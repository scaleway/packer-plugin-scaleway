package checks

import (
	"context"
	"fmt"
	"reflect"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*ImageCheck)(nil)

type ImageCheck struct {
	zone      scw.Zone
	imageName string

	tags                  []string
	size                  *scw.Size
	rootVolumeSnapshot    SnapshotCheck
	extraVolumesSnapshots map[string]SnapshotCheck
}

func Image(zone scw.Zone, name string) *ImageCheck {
	return &ImageCheck{
		zone:      zone,
		imageName: name,
	}
}

func (c *ImageCheck) SizeInGB(size uint64) *ImageCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *ImageCheck) Tags(tags []string) *ImageCheck {
	c.tags = tags

	return c
}

func (c *ImageCheck) RootVolumeSnapshot(snapshotCheck SnapshotCheck) *ImageCheck {
	c.rootVolumeSnapshot = snapshotCheck

	return c
}

func (c *ImageCheck) ExtraVolumeSnapshot(index string, snapshotCheck SnapshotCheck) *ImageCheck {
	if c.extraVolumesSnapshots == nil {
		c.extraVolumesSnapshots = map[string]SnapshotCheck{index: snapshotCheck}
	} else {
		c.extraVolumesSnapshots[index] = snapshotCheck
	}

	return c
}

func (c *ImageCheck) CheckName() string {
	return "Image"
}

func findImage(ctx context.Context, zone scw.Zone, imageName string) (*instance.Image, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)
	images := []*instance.Image(nil)

	resp, err := api.ListImages(&instance.ListImagesRequest{
		Name:    &imageName,
		Zone:    zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	// Filtering by name returns all images which name is prefixed with imageName,
	// so we need to select the ones which name is strictly equal.
	for _, img := range resp.Images {
		if img.Name == imageName {
			images = append(images, img)
		}
	}

	if len(resp.Images) == 0 {
		return nil, fmt.Errorf("image %s not found, no images found", imageName)
	}

	if len(resp.Images) > 1 {
		return nil, fmt.Errorf("multiple images found with name %s", imageName)
	}

	return resp.Images[0], nil
}

func computeImageSize(ctx context.Context, image *instance.Image) (scw.Size, error) {
	testCtx := tester.ExtractCtx(ctx)
	blockAPI := block.NewAPI(testCtx.ScwClient)
	imageSize := image.RootVolume.Size

	if image.RootVolume.VolumeType == instance.VolumeVolumeTypeSbsSnapshot {
		blockSnapshot, err := blockAPI.GetSnapshot(&block.GetSnapshotRequest{
			Zone:       image.Zone,
			SnapshotID: image.RootVolume.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return 0, fmt.Errorf("could not get block snapshot %s: %w", image.RootVolume.ID, err)
		}

		imageSize += blockSnapshot.Size
	}

	for _, extraVolume := range image.ExtraVolumes {
		blockSnapshot, err := blockAPI.GetSnapshot(&block.GetSnapshotRequest{
			Zone:       image.Zone,
			SnapshotID: extraVolume.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return 0, fmt.Errorf("could not get block snapshot %s: %w", extraVolume.ID, err)
		}

		imageSize += blockSnapshot.Size
	}

	return imageSize, nil
}

func (c *ImageCheck) Check(ctx context.Context) error {
	image, err := findImage(ctx, c.zone, c.imageName)
	if err != nil {
		return err
	}

	if len(c.tags) > 0 && !reflect.DeepEqual(c.tags, image.Tags) {
		return fmt.Errorf("image tags did not match, expected %v, got %v", c.tags, image.Tags)
	}

	imageSize, err := computeImageSize(ctx, image)
	if err != nil {
		return fmt.Errorf("could not calculate image size: %w", err)
	}

	if c.size != nil && imageSize != *c.size {
		return fmt.Errorf("image size %d does not match expected %d", image.RootVolume.Size, *c.size)
	}

	if c.rootVolumeSnapshot != nil {
		err = c.rootVolumeSnapshot.Check(ctx)
		if err != nil {
			return fmt.Errorf("root volume check failed: %w", err)
		}
	}

	if c.extraVolumesSnapshots != nil {
		for imageExtraVolumeKey := range image.ExtraVolumes {
			if _, exists := c.extraVolumesSnapshots[imageExtraVolumeKey]; !exists {
				return fmt.Errorf("expected extra volume %q does not exist in image", imageExtraVolumeKey)
			}
		}

		for checkExtraVolumeKey, snapshotCheck := range c.extraVolumesSnapshots {
			if _, exists := image.ExtraVolumes[checkExtraVolumeKey]; !exists {
				return fmt.Errorf("unexpected extra volume %q found in image", checkExtraVolumeKey)
			}

			err = snapshotCheck.Check(ctx)
			if err != nil {
				return fmt.Errorf("extra volume %q check failed: %w", checkExtraVolumeKey, err)
			}
		}
	}

	return nil
}
