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
			req.FromSnapshot = &block.CreateVolumeRequestFromSnapshot{
				SnapshotID: requestedVolume.SnapshotID,
			}
		} else {
			req.FromEmpty = &block.CreateVolumeRequestFromEmpty{
				Size: scw.Size(requestedVolume.SizeInGB) * scw.GB,
			}
		}

		volume, err := blockAPI.CreateVolume(req, scw.WithContext(ctx))
		if err != nil {
			return putErrorAndHalt(state, ui, err)
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
		var err error
		if volume.VolumeType == instance.VolumeServerVolumeTypeSbsVolume {
			err = waitAndDeleteBlockVolume(blockAPI, scw.Zone(c.Zone), volume.ID)
		} else {
			err = waitAndDeleteInstanceVolume(instanceAPI, scw.Zone(c.Zone), volume.ID)
		}

		// A 404 means the volume is already gone, for example a local volume
		// deleted by terminate.
		if err != nil && !httperrors.Is404(err) {
			state.Put("error", fmt.Errorf("error removing volume %s: %w", volume.ID, err))
			ui.Error(fmt.Sprintf("Error removing volume %s: %s", volume.ID, err))
		}
	}
}

func waitAndDeleteInstanceVolume(instanceAPI *instance.API, zone scw.Zone, volumeID string) error {
	_, err := instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
		VolumeID: volumeID,
		Zone:     zone,
	})
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	err = instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: volumeID,
		Zone:     zone,
	})
	if err != nil {
		return err
	}

	return nil
}

func waitAndDeleteBlockVolume(blockAPI *block.API, zone scw.Zone, volumeID string) error {
	volumeTerminalStatus := block.VolumeStatusAvailable

	_, err := blockAPI.WaitForVolumeAndReferences(&block.WaitForVolumeAndReferencesRequest{
		VolumeID:             volumeID,
		Zone:                 zone,
		VolumeTerminalStatus: &volumeTerminalStatus,
	})
	if err != nil {
		return err
	}

	err = blockAPI.DeleteVolume(&block.DeleteVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	})
	if err != nil {
		return err
	}

	return nil
}
