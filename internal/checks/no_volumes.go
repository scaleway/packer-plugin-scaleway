package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*NoVolumesCheck)(nil)

type NoVolumesCheck struct {
	zone scw.Zone
}

// NoVolume checks that the current project does not contain any volume, block or instance.
func NoVolume(zone scw.Zone) *NoVolumesCheck {
	return &NoVolumesCheck{
		zone: zone,
	}
}

func (c *NoVolumesCheck) CheckName() string {
	return "No volumes"
}

func (c *NoVolumesCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	instanceAPI := instance.NewAPI(testCtx.ScwClient)
	blockAPI := block.NewAPI(testCtx.ScwClient)

	resp, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{
		Zone:    c.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list instance volumes: %w", err)
	}

	if len(resp.Volumes) != 0 {
		return fmt.Errorf("expected 0 instance volumes, got %d", len(resp.Volumes))
	}

	blockResp, err := blockAPI.ListVolumes(&block.ListVolumesRequest{
		Zone:      c.zone,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list block volumes: %w", err)
	}

	if len(blockResp.Volumes) != 0 {
		return fmt.Errorf("expected 0 block volumes, got %d", len(blockResp.Volumes))
	}

	return nil
}
