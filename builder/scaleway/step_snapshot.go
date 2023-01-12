package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepSnapshot struct{}

func (s *stepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	volumeID := state.Get("root_volume_id").(string)

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))
	createSnapshotResp, err := instanceAPI.CreateSnapshot(&instance.CreateSnapshotRequest{
		Name:       c.SnapshotName,
		VolumeID:   &volumeID,
		VolumeType: instance.SnapshotVolumeTypeUnified,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	waitSnapshotRequest := &instance.WaitForSnapshotRequest{
		SnapshotID: createSnapshotResp.Snapshot.ID,
		Zone:       scw.Zone(c.Zone),
		Timeout:    &c.SnapshotCreationTimeout,
	}

	snapshot, err := instanceAPI.WaitForSnapshot(waitSnapshotRequest)
	if err != nil {
		err := fmt.Errorf("snapshot is not available: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if snapshot.State != instance.SnapshotStateAvailable {
		err := fmt.Errorf("snapshot is in error state")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Snapshot ID: %s", createSnapshotResp.Snapshot.ID)
	state.Put("snapshot_id", createSnapshotResp.Snapshot.ID)
	state.Put("snapshot_name", c.SnapshotName)
	state.Put("region", c.Zone) // Deprecated
	state.Put("zone", c.Zone)

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
