package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*ImageCheck)(nil)

func Image(zone scw.Zone, name string) *ImageCheck {
	return &ImageCheck{
		zone:      zone,
		imageName: name,
	}
}

type ImageCheck struct {
	zone      scw.Zone
	imageName string
	tags      []string

	rootVolumeType     *string
	rootVolumeSnapshot *BlockSnapshotCheck
	size               *scw.Size
	extraVolumesType   map[string]string
}

func (c *ImageCheck) RootVolumeType(rootVolumeType string) *ImageCheck {
	c.rootVolumeType = &rootVolumeType

	return c
}

func (c *ImageCheck) RootVolumeBlockSnapshot(snapshotCheck *BlockSnapshotCheck) *ImageCheck {
	c.rootVolumeSnapshot = snapshotCheck

	return c
}

func (c *ImageCheck) ExtraVolumeType(key string, volumeType string) *ImageCheck {
	if c.extraVolumesType == nil {
		c.extraVolumesType = map[string]string{}
	}

	c.extraVolumesType[key] = volumeType

	return c
}

func (c *ImageCheck) SizeInGb(size uint64) *ImageCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *ImageCheck) Tags(tags []string) *ImageCheck {
	c.tags = tags
	return c
}

func (c *ImageCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)
	images := []*instance.Image(nil)

	resp, err := api.ListImages(&instance.ListImagesRequest{
		Name:    &c.imageName,
		Zone:    c.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range resp.Images {
		if img.Name == c.imageName {
			images = append(images, img)
		}
	}

	if len(images) == 0 {
		return fmt.Errorf("image %s not found, no images found", c.imageName)
	}

	if len(images) > 1 {
		return fmt.Errorf("multiple images found with name %s", c.imageName)
	}

	image := images[0]

	if image.Name != c.imageName {
		return fmt.Errorf("image name %s does not match expected %s", image.Name, c.imageName)
	}

	if c.rootVolumeType != nil && string(image.RootVolume.VolumeType) != *c.rootVolumeType {
		return fmt.Errorf("image root volume type %s does not match expected %s", image.RootVolume.VolumeType, *c.rootVolumeType)
	}

	if c.size != nil && image.RootVolume.Size != *c.size {
		return fmt.Errorf("image size %d does not match expected %d", image.RootVolume.Size, *c.size)
	}

	if c.extraVolumesType != nil {
		for k, v := range c.extraVolumesType {
			vol, exists := image.ExtraVolumes[k]
			if !exists {
				return fmt.Errorf("extra volume %s does not exist", k)
			}

			if string(vol.VolumeType) != v {
				return fmt.Errorf("extra volume %s type %s does not match expected %s", k, vol.VolumeType, v)
			}
		}
	}

	if c.tags != nil {
		for _, expectedTag := range c.tags {
			found := false
			for _, actualTag := range image.Tags {
				if actualTag == expectedTag {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("expected tag %q not found on image %s", expectedTag, c.imageName)
			}
		}
	}

	return nil
}
