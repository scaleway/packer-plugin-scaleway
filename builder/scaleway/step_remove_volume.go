package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepRemoveVolume struct{}

func (s *stepRemoveVolume) Run(_ context.Context, _ multistep.StateBag) multistep.StepAction {
	// nothing to do ... only cleanup interests us
	return multistep.ActionContinue
}

func (s *stepRemoveVolume) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("snapshots"); !ok {
		// volume will be detached from server only after snapshotting ... so we don't
		// need to remove volume before snapshot step.
		return
	}

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	if !c.RemoveVolume {
		return
	}

	ui.Say("Removing Volumes ...")

	volumes := state.Get("volumes").([]*instance.VolumeServer)
	for _, volume := range volumes {
		err := instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
			VolumeID: volume.ID,
			Zone:     scw.Zone(c.Zone),
		})
		if err != nil {
			err := fmt.Errorf("error removing block volume %s: %s", volume.ID, err)
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error removing block volume %s: %s. (Ignored)", volume.ID, err))
		}
	}
}
