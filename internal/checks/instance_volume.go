package checks

import (
	"context"
	"fmt"
	"reflect"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*InstanceVolumeCheck)(nil)

type InstanceVolumeCheck struct {
	zone       scw.Zone
	namePrefix string

	volumeName *string
	size       *scw.Size
	tags       []string
}

func InstanceVolume(zone scw.Zone, name string) *InstanceVolumeCheck {
	return &InstanceVolumeCheck{
		zone:       zone,
		namePrefix: name,
	}
}

func (c *InstanceVolumeCheck) SizeInGB(size uint64) *InstanceVolumeCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *InstanceVolumeCheck) Tags(tags []string) *InstanceVolumeCheck {
	c.tags = tags

	return c
}

func (c *InstanceVolumeCheck) Name(volumeName string) *InstanceVolumeCheck {
	c.volumeName = &volumeName

	return c
}

func findInstanceVolume(ctx context.Context, zone scw.Zone, namePrefix string) (*instance.Volume, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&instance.ListVolumesRequest{
		Zone:    zone,
		Name:    &namePrefix,
		Project: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list instance volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return nil, fmt.Errorf("instance volume with prefix %q not found, no volumes found", namePrefix)
	}

	if len(resp.Volumes) > 1 {
		return nil, fmt.Errorf("multiple instance volumes found with name %q", namePrefix)
	}

	return resp.Volumes[0], nil
}

func (c *InstanceVolumeCheck) CheckName() string {
	return fmt.Sprintf("Instance volume \"%s...\"", c.namePrefix)
}

func (c *InstanceVolumeCheck) Check(ctx context.Context) error {
	volume, err := findInstanceVolume(ctx, c.zone, c.namePrefix)
	if err != nil {
		return err
	}

	if c.volumeName != nil && volume.Name != *c.volumeName {
		return fmt.Errorf("volume name %q does not match expected volume name %q", volume.Name, *c.volumeName)
	}

	if len(c.tags) > 0 && !reflect.DeepEqual(c.tags, volume.Tags) {
		return fmt.Errorf("volume tags did not match, expected %v, got %v", c.tags, volume.Tags)
	}

	if c.size != nil && volume.Size != *c.size {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *c.size)
	}

	return nil
}
