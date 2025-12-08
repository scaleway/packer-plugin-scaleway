package cleanup

import (
	"context"
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*ImageCleanup)(nil)

type ImageCleanup struct {
	zone      scw.Zone
	imageName string
}

func Image(zone scw.Zone, name string) *ImageCleanup {
	return &ImageCleanup{
		zone:      zone,
		imageName: name,
	}
}

func (i *ImageCleanup) Cleanup(ctx context.Context, t *testing.T) error {
	t.Helper()

	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListImages(&instance.ListImagesRequest{
		Name:    &i.imageName,
		Zone:    i.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if len(resp.Images) == 0 {
		return fmt.Errorf("could not find any image by the name %q", i.imageName)
	}

	err = api.DeleteImage(&instance.DeleteImageRequest{
		Zone:    i.zone,
		ImageID: resp.Images[0].ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	t.Logf("deleted image %q\n", resp.Images[0].Name)

	return nil
}
