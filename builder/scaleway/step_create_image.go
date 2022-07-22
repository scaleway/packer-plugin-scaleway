package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepImage struct{}

func (s *stepImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	snapshotID := state.Get("snapshot_id").(string)
	bootscriptID := ""

	ui.Say(fmt.Sprintf("Creating image: %v", c.ImageName))

	imageID := c.Image

	// if not a UUID, we check the Marketplace API
	_, err := uuid.ParseUUID(c.Image)
	if err != nil {
		apiMarketplace := marketplace.NewAPI(state.Get("client").(*scw.Client))
		imageID, err = apiMarketplace.GetLocalImageIDByLabel(&marketplace.GetLocalImageIDByLabelRequest{
			ImageLabel:     c.Image,
			Zone:           scw.Zone(c.Zone),
			CommercialType: c.CommercialType,
		})
		if err != nil {
			err := fmt.Errorf("error getting initial image info from marketplace: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	imageResp, err := instanceAPI.GetImage(&instance.GetImageRequest{
		ImageID: imageID,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error getting initial image info: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if imageResp.Image.DefaultBootscript != nil {
		bootscriptID = imageResp.Image.DefaultBootscript.ID
	}

	createImageResp, err := instanceAPI.CreateImage(&instance.CreateImageRequest{
		Arch:              imageResp.Image.Arch,
		DefaultBootscript: bootscriptID,
		Name:              c.ImageName,
		RootVolume:        snapshotID,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	waitImageRequest := &instance.WaitForImageRequest{
		ImageID: createImageResp.Image.ID,
		Zone:    scw.Zone(c.Zone),
		Timeout: &c.ImageCreationTimeout,
	}

	image, err := instanceAPI.WaitForImage(waitImageRequest)
	if err != nil {
		err := fmt.Errorf("image is not available: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if image.State != instance.ImageStateAvailable {
		err := fmt.Errorf("image is in error state")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Image ID: %s", createImageResp.Image.ID)
	state.Put("image_id", createImageResp.Image.ID)
	state.Put("image_name", c.ImageName)
	state.Put("region", c.Zone) // Deprecated
	state.Put("zone", c.Zone)

	return multistep.ActionContinue
}

func (s *stepImage) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
