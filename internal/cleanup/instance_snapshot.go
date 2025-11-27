package cleanup

import (
	"context"
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*InstanceSnapshotCleanup)(nil)

type InstanceSnapshotCleanup struct {
	zone       scw.Zone
	namePrefix string
}

func InstanceSnapshot(zone scw.Zone, namePrefix string) *InstanceSnapshotCleanup {
	return &InstanceSnapshotCleanup{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (i *InstanceSnapshotCleanup) Cleanup(ctx context.Context, t *testing.T) error {
	t.Helper()

	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&instance.ListSnapshotsRequest{
		Name:    &i.namePrefix,
		Zone:    i.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list instance snapshots: %w", err)
	}

	if len(resp.Snapshots) == 0 {
		return fmt.Errorf("could not find any instance snapshot prefixed with %q", i.namePrefix)
	}

	err = api.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		Zone:       i.zone,
		SnapshotID: resp.Snapshots[0].ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete instance snapshot: %w", err)
	}

	t.Logf("deleted instance snapshot %q\n", resp.Snapshots[0].Name)

	return nil
}
