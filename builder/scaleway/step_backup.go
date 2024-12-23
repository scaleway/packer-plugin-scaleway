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

	ui.Say(fmt.Sprintf("Backing up server to image: %v", c.ImageName))

	actionResp, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
		ServerID: server.ID,
		Action:   instance.ServerActionBackup,
		Name:     &c.ImageName,
		Volumes:  backupVolumesFromServer(server),
		Zone:     scw.Zone(c.Zone),
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("failed to backup server: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imageID, err := imageIDFromBackupResult(actionResp.Task.HrefResult)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

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

	snapshots := artifactSnapshotFromImage(image)

	// Apply tags to image, volumes and snapshots
	if len(c.Tags) != 0 {
		err = applyTags(ctx, instanceAPI, scw.Zone(c.Zone), imageID, server.Volumes, snapshots, c.Tags)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
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

func artifactSnapshotFromImage(image *instance.Image) []ArtifactSnapshot {
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
	return snapshots
}

func backupVolumesFromServer(server *instance.Server) map[string]*instance.ServerActionRequestVolumeBackupTemplate {
	backupVolumes := map[string]*instance.ServerActionRequestVolumeBackupTemplate{}

	for _, volume := range server.Volumes {
		backupVolumes[volume.ID] = &instance.ServerActionRequestVolumeBackupTemplate{
			//VolumeType: instance.SnapshotVolumeTypeUnified,
		}
	}
	return backupVolumes
}

func imageIDFromBackupResult(hrefResult string) (string, error) {
	// HrefResult format is /images/<uuid>
	hrefSplit := strings.Split(hrefResult, "/")
	if len(hrefSplit) != 3 {
		return "", fmt.Errorf("failed to parse backup request response (%s)", hrefResult)
	}
	imageID := hrefSplit[2]

	return imageID, nil
}

func applyTags(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, imageID string, volumes map[string]*instance.VolumeServer, snapshots []ArtifactSnapshot, tags []string) error {
	if _, err := instanceAPI.UpdateImage(&instance.UpdateImageRequest{
		ImageID: imageID,
		Zone:    zone,
		Tags:    &tags,
	}, scw.WithContext(ctx)); err != nil {
		return fmt.Errorf("failed to set tags on the image: %w", err)
	}

	for _, volume := range volumes {
		if _, err := instanceAPI.UpdateVolume(&instance.UpdateVolumeRequest{
			VolumeID: volume.ID,
			Zone:     zone,
			Tags:     &tags,
		}, scw.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to set tags on the volume: %w", err)
		}
	}

	for _, snapshot := range snapshots {
		if _, err := instanceAPI.UpdateSnapshot(&instance.UpdateSnapshotRequest{
			SnapshotID: snapshot.ID,
			Zone:       zone,
			Tags:       &tags,
		}, scw.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to set tags on the snapshot: %w", err)
		}
	}

	return nil
}
