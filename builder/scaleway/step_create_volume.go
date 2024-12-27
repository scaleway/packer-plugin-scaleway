package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/internal/httperrors"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const StateKeyCreatedVolumes = "created_volumes"

type stepCreateVolume struct{}

func (s *stepCreateVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*scw.Client)
	blockAPI := block.NewAPI(client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	volumeTemplates := []*instance.VolumeServerTemplate(nil)
	for _, requestedVolume := range c.BlockVolumes {
		req := &block.CreateVolumeRequest{
			Zone:      scw.Zone(c.Zone),
			Name:      requestedVolume.Name,
			PerfIops:  requestedVolume.IOPS,
			ProjectID: c.ProjectID,
		}
		if requestedVolume.SnapshotID != "" {
			req.FromSnapshot = &block.CreateVolumeRequestFromSnapshot{}
		} else {
			req.FromEmpty = &block.CreateVolumeRequestFromEmpty{
				Size: scw.Size(requestedVolume.Size),
			}
		}
		volume, err := blockAPI.CreateVolume(req, scw.WithContext(ctx))
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		volumeTemplates = append(volumeTemplates, &instance.VolumeServerTemplate{
			ID:         &volume.ID,
			VolumeType: instance.VolumeVolumeTypeSbsVolume,
		})
	}

	state.Put(StateKeyCreatedVolumes, volumeTemplates)

	return multistep.ActionContinue
}

func (s *stepCreateVolume) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	if !c.RemoveVolume {
		return
	}

	blockAPI := block.NewAPI(state.Get("client").(*scw.Client))

	_, serverWasCreated := state.GetOk("server_id")
	createdVolumesI, createdVolumesExists := state.GetOk("created_volumes")
	if !serverWasCreated && createdVolumesExists {
		// If server was not created, we need to clean up manually created volumes
		createdVolumes := createdVolumesI.([]*instance.VolumeServerTemplate)
		for _, volume := range createdVolumes {
			err := blockAPI.DeleteVolume(&block.DeleteVolumeRequest{
				Zone:     scw.Zone(c.Zone),
				VolumeID: *volume.ID,
			})
			if err != nil {
				ui.Error(fmt.Sprintf("failed to cleanup block volume %s: %s", *volume.ID, err))
			}
		}
	}

	if _, ok := state.GetOk("snapshots"); !ok {
		// volume will be detached from server only after snapshotting ... so we don't
		// need to remove volume before snapshot step.
		return
	}

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))

	ui.Say("Removing Volumes ...")

	volumes := state.Get("volumes").([]*instance.VolumeServer)
	for _, volume := range volumes {
		err := instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
			VolumeID: volume.ID,
			Zone:     scw.Zone(c.Zone),
		})
		if err != nil && !httperrors.Is404(err) {
			err := fmt.Errorf("error removing block volume %s: %s", volume.ID, err)
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error removing block volume %s: %s. (Ignored)", volume.ID, err))
		}
		if err == nil {
			continue
		}

		err = blockAPI.DeleteVolume(&block.DeleteVolumeRequest{
			Zone:     scw.Zone(c.Zone),
			VolumeID: volume.ID,
		})
		if err != nil {
			err := fmt.Errorf("error removing block volume %s: %s", volume.ID, err)
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error removing block volume %s: %s. (Ignored)", volume.ID, err))
		}
	}
}
