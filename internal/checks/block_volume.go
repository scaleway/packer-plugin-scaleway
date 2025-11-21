package checks

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*BlockVolumeCheck)(nil)

type BlockVolumeCheck struct {
	zone       scw.Zone
	namePrefix string

	volumeName *string
	tags       []string
	size       *scw.Size
	iops       *uint32
}

func BlockVolume(zone scw.Zone, namePrefix string) *BlockVolumeCheck {
	return &BlockVolumeCheck{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (c *BlockVolumeCheck) Name(volumeName string) *BlockVolumeCheck {
	c.volumeName = &volumeName

	return c
}

func (c *BlockVolumeCheck) Tags(tags []string) *BlockVolumeCheck {
	c.tags = tags

	return c
}

func (c *BlockVolumeCheck) SizeInGB(size uint64) *BlockVolumeCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *BlockVolumeCheck) IOPS(iops uint32) *BlockVolumeCheck {
	c.iops = &iops

	return c
}

func findBlockVolumes(ctx context.Context, zone scw.Zone, namePrefix string) ([]*block.Volume, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&block.ListVolumesRequest{
		Zone:      zone,
		Name:      &namePrefix,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list block volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return nil, fmt.Errorf("could not find any block volume prefixed with %q", namePrefix)
	}

	return resp.Volumes, nil
}

func (c *BlockVolumeCheck) compareSingleBlockVolume(volume *block.Volume) error {
	if c.volumeName != nil && volume.Name != *c.volumeName {
		return fmt.Errorf("volume name %q does not match expected volume name %q", volume.Name, *c.volumeName)
	}

	if len(c.tags) > 0 && !reflect.DeepEqual(c.tags, volume.Tags) {
		return fmt.Errorf("volume tags did not match, expected %v, got %v", c.tags, volume.Tags)
	}

	if c.size != nil && volume.Size != *c.size {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *c.size)
	}

	if c.iops != nil && volume.Specs != nil && volume.Specs.PerfIops != nil && *volume.Specs.PerfIops != *c.iops {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *c.size)
	}

	return nil
}

func (c *BlockVolumeCheck) CheckName() string {
	return fmt.Sprintf("Block volume \"%s...\"", c.namePrefix)
}

func (c *BlockVolumeCheck) Check(ctx context.Context) error {
	volumes, err := findBlockVolumes(ctx, c.zone, c.namePrefix)
	if err != nil {
		return err
	}

	volumeMatchingErrors := []error(nil)
	for _, volume := range volumes {
		err = c.compareSingleBlockVolume(volume)
		if err != nil {
			volumeMatchingErrors = append(volumeMatchingErrors, err)
		}
	}

	if len(volumeMatchingErrors) < len(volumes) {
		return nil
	}

	return fmt.Errorf("no block volume matched the expected specs, got the following matching errors:\n%s", errors.Join(volumeMatchingErrors...))
}
