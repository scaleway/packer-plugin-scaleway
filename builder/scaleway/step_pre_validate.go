package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time-consuming work
type StepPreValidate struct {
	Force        bool
	ImageName    string
	SnapshotName string
}

func (s *StepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.Force {
		ui.Say("Force flag found, skipping pre-validating image name")

		return multistep.ActionContinue
	}

	ui.Say("Pre-validating image name: " + s.ImageName)

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))

	images, err := instanceAPI.ListImages(
		&instance.ListImagesRequest{Name: &s.ImageName},
		scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error: getting image list: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())

		return multistep.ActionHalt
	}

	for _, im := range images.Images {
		if im.Name == s.ImageName {
			err := fmt.Errorf("error: image name: '%s' is used by existing image with ID %s",
				s.ImageName, im.ID)
			state.Put("error", err)
			ui.Error(err.Error())

			return multistep.ActionHalt
		}
	}

	ui.Say("Pre-validating snapshot name: " + s.SnapshotName)

	snapshots, err := instanceAPI.ListSnapshots(
		&instance.ListSnapshotsRequest{Name: &s.SnapshotName},
		scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error: getting snapshot list: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())

		return multistep.ActionHalt
	}

	for _, sn := range snapshots.Snapshots {
		if sn.Name == s.SnapshotName {
			err := fmt.Errorf("error: snapshot name: '%s' is used by existing snapshot with ID %s",
				s.SnapshotName, sn.ID)
			state.Put("error", err)
			ui.Error(err.Error())

			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
