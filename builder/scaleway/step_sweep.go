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

func (s *stepSweep) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// nothing to do ... only cleanup interests us
	return multistep.ActionContinue
}

func (s *stepSweep) Cleanup(state multistep.StateBag) {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	snapshotID := state.Get("snapshot_id").(string)
	imageID := state.Get("image_id").(string)

	ui.Say("Deleting Image...")

	err := instanceAPI.DeleteImage(&instance.DeleteImageRequest{
		ImageID: imageID,
	})
	if err != nil {
		err := fmt.Errorf("error deleting image: %s", err)
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error deleting image: %s. (Ignored)", err))
	}

	ui.Say("Deleting Snapshot...")

	err = instanceAPI.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		SnapshotID: snapshotID,
	})
	if err != nil {
		err := fmt.Errorf("error deleting snapshot: %s", err)
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error deleting snapshot: %s. (Ignored)", err))
	}

}
