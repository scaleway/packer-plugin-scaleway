package scaleway

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// StepRemoveOldImages deletes the private images of the project that share the
// name of the image just created, so that only the new one remains
// It is only run when the force flag is set
type StepRemoveOldImages struct{}

func (s *StepRemoveOldImages) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*scw.Client)
	instanceAPI := instance.NewAPI(client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	imageName := state.Get("image_name").(string)
	newImageID := state.Get("image_id").(string)
	zone := scw.Zone(c.Zone)

	ui.Say("Force flag found, removing old images named: " + imageName + " in zone " + zone.String())

	images, err := instanceAPI.ListImages(&instance.ListImagesRequest{
		Zone:    zone,
		Name:    &imageName,
		Project: &c.ProjectID,
		Public:  new(false),
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error: getting image list: %w", err))
	}

	var errs []error

	for _, im := range images.Images {
		// The name filter is not an exact match
		if im.Name != imageName || im.ID == newImageID {
			continue
		}

		ui.Say(fmt.Sprintf("Deleting old image: %s (ID: %s)", im.Name, im.ID))

		err := instanceAPI.DeleteImage(&instance.DeleteImageRequest{
			Zone:    zone,
			ImageID: im.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to delete old image %s: %w", im.ID, err))
		}
	}

	if len(errs) > 0 {
		return putErrorAndHalt(state, ui, errors.Join(errs...))
	}

	return multistep.ActionContinue
}

func (s *StepRemoveOldImages) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
