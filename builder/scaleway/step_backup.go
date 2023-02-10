package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepBackup struct{}

func (s *stepBackup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	server := state.Get("server").(*instance.Server)

	backupVolumes := map[string]*instance.ServerActionRequestVolumeBackupTemplate{}

	for _, volume := range server.Volumes {
		backupVolumes[volume.ID] = &instance.ServerActionRequestVolumeBackupTemplate{
			VolumeType: instance.SnapshotVolumeTypeUnified,
		}
	}

	ui.Say(fmt.Sprintf("Backing up server to image: %v", c.ImageName))

	actionResp, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
		ServerID: server.ID,
		Action:   instance.ServerActionBackup,
		Name:     &c.ImageName,
		Volumes:  backupVolumes,
		Zone:     scw.Zone(c.Zone),
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("failed to backup server: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// HrefResult format is /images/<uuid>
	hrefSplit := strings.Split(actionResp.Task.HrefResult, "/")
	if len(hrefSplit) != 3 {
		err := fmt.Errorf("failed to parse backup request response (%s)", actionResp.Task.HrefResult)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	imageID := hrefSplit[2]

	image, err := instanceAPI.WaitForImage(&instance.WaitForImageRequest{
		ImageID: imageID,
		Zone:    scw.Zone(c.Zone),
		Timeout: &c.ImageCreationTimeout,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("failed to fetch generated image: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	snapshots := []ArtifactSnapshot{
		{
			ID:   image.RootVolume.ID,
			Name: image.RootVolume.Name,
		},
	}
	for _, extraVolume := range image.ExtraVolumes {
		snapshots = append(snapshots, ArtifactSnapshot{
			ID:   extraVolume.ID,
			Name: extraVolume.Name,
		})
	}

	state.Put("snapshots", snapshots)
	state.Put("image_id", image.ID)
	state.Put("image_name", c.ImageName)
	state.Put("region", c.Zone) // Deprecated
	state.Put("zone", c.Zone)

	return multistep.ActionContinue
}

func (s *stepBackup) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
