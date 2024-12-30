package checks

import (
	"context"
	"e2e_tests/tester"
	"fmt"

	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*NoVolumesCheck)(nil)

type NoVolumesCheck struct {
	zone scw.Zone
}

func (c *NoVolumesCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	instanceAPI := instance.NewAPI(testCtx.ScwClient)
	blockAPI := block.NewAPI(testCtx.ScwClient)

	resp, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{
		Zone: c.zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list instance volumes: %w", err)
	}

	if len(resp.Volumes) != 0 {
		return fmt.Errorf("expected 0 instance volumes, got %d", len(resp.Volumes))
	}

	blockResp, err := blockAPI.ListVolumes(&block.ListVolumesRequest{
		Zone: c.zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list block volumes: %w", err)
	}

	if len(blockResp.Volumes) != 0 {
		return fmt.Errorf("expected 0 block volumes, got %d", len(blockResp.Volumes))
	}

	return nil
}

// NoVolume checks that the current project does not contain any volume, block or instance.
func NoVolume(zone scw.Zone) *NoVolumesCheck {
	return &NoVolumesCheck{
		zone: zone,
	}
}

type VolumeCheck struct {
	zone       scw.Zone
	volumeName string

	size *scw.Size
}

func (c *VolumeCheck) SizeInGB(size uint64) *VolumeCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *VolumeCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&instance.ListVolumesRequest{
		Zone:    c.zone,
		Name:    &c.volumeName,
		Project: &testCtx.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("failed to list instance volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return fmt.Errorf("volume %s not found, no volumes found", c.volumeName)
	}

	if len(resp.Volumes) > 1 {
		return fmt.Errorf("multiple volumes found with name %s", c.volumeName)
	}

	volume := resp.Volumes[0]

	if volume.Name != c.volumeName {
		return fmt.Errorf("volume name %s does not match expected volume name %s", volume.Name, c.volumeName)
	}

	if c.size != nil && volume.Size != *c.size {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *c.size)
	}

	return nil
}

func Volume(zone scw.Zone, name string) *VolumeCheck {
	return &VolumeCheck{
		zone:       zone,
		volumeName: name,
	}
}
