package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepSweep struct{}

func (s *stepSweep) Run(_ context.Context, _ multistep.StateBag) multistep.StepAction {
	// nothing to do ... only cleanup interests us
	return multistep.ActionContinue
}

func (s *stepSweep) Cleanup(state multistep.StateBag) {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	snapshots := state.Get("snapshots").([]ArtifactSnapshot)
	imageID := state.Get("image_id").(string)

	ui.Say("Deleting Image...")

	err := instanceAPI.DeleteImage(&instance.DeleteImageRequest{
		ImageID: imageID,
		Zone:    scw.Zone(c.Zone),
	})
	if err != nil {
		err := fmt.Errorf("error deleting image: %s", err)
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error deleting image: %s. (Ignored)", err))
	}

	ui.Say("Deleting Snapshot...")

	for _, snapshot := range snapshots {
		err = instanceAPI.DeleteSnapshot(&instance.DeleteSnapshotRequest{
			SnapshotID: snapshot.ID,
			Zone:       scw.Zone(c.Zone),
		})
		if err != nil {
			err := fmt.Errorf("error deleting snapshot: %s", err)
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error deleting snapshot: %s. (Ignored)", err))
		}
	}
}
