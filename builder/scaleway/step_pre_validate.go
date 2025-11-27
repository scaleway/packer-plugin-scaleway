package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const DefaultZoneInMock = scw.ZoneFrPar1

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time-consuming work
type StepPreValidate struct {
	Force          bool
	ImageName      string
	SnapshotsNames []string
}

func (s *StepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*scw.Client)
	instanceAPI := instance.NewAPI(client)
	blockAPI := block.NewAPI(client)
	ui := state.Get("ui").(packersdk.Ui)

	zone := DefaultZoneInMock
	if c := state.Get("config"); c != nil {
		zone = scw.Zone(c.(*Config).Zone)
	}

	if s.Force {
		ui.Say("Force flag found, skipping pre-validating image name")

		return multistep.ActionContinue
	}

	ui.Say("Pre-validating image name: " + s.ImageName + " in zone " + zone.String())

	images, err := instanceAPI.ListImages(
		&instance.ListImagesRequest{
			Name: &s.ImageName,
			Zone: zone,
		},
		scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error: getting image list: %w", err))
	}

	for _, im := range images.Images {
		if im.Name == s.ImageName {
			return putErrorAndHalt(state, ui, fmt.Errorf("error: image name: '%s' is used by existing image with ID %s", s.ImageName, im.ID))
		}
	}

	ui.Say("Pre-validating snapshot names")

	instanceSnapshots, err := instanceAPI.ListSnapshots(&instance.ListSnapshotsRequest{
		Zone: zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error: getting snapshot list: %w", err))
	}

	for _, sn := range instanceSnapshots.Snapshots {
		// Only root volume can be a local one so we only need to check the name at index 0 against instance snapshots
		if sn.Name == s.SnapshotsNames[0] {
			return putErrorAndHalt(state, ui, fmt.Errorf("error: snapshot name: '%s' is used by existing snapshot with ID %s", s.SnapshotsNames[0], sn.ID))
		}
	}

	blockSnapshots, err := blockAPI.ListSnapshots(&block.ListSnapshotsRequest{
		Zone: zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error: getting snapshot list: %w", err))
	}

	for _, sn := range blockSnapshots.Snapshots {
		for _, snapshotName := range s.SnapshotsNames {
			if sn.Name == snapshotName {
				return putErrorAndHalt(state, ui, fmt.Errorf("error: snapshot name: '%s' is used by existing snapshot with ID %s", snapshotName, sn.ID))
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
