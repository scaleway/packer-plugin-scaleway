package cleanup

import (
	"context"
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*InstanceVolumeCleanup)(nil)

type InstanceVolumeCleanup struct {
	zone       scw.Zone
	namePrefix string
}

func InstanceVolume(zone scw.Zone, namePrefix string) *InstanceVolumeCleanup {
	return &InstanceVolumeCleanup{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (i *InstanceVolumeCleanup) Cleanup(ctx context.Context, t *testing.T) error {
	t.Helper()

	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&instance.ListVolumesRequest{
		Name:    &i.namePrefix,
		Zone:    i.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list instance volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return fmt.Errorf("could not find any instance volume prefixed with %q", i.namePrefix)
	}

	err = api.DeleteVolume(&instance.DeleteVolumeRequest{
		Zone:     i.zone,
		VolumeID: resp.Volumes[0].ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete instance volume: %w", err)
	}

	t.Logf("deleted instance volume %q\n", resp.Volumes[0].Name)

	return nil
}
