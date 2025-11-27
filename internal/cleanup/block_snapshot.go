package cleanup

import (
	"context"
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*BlockSnapshotCleanup)(nil)

type BlockSnapshotCleanup struct {
	zone       scw.Zone
	namePrefix string
}

func BlockSnapshot(zone scw.Zone, namePrefix string) *BlockSnapshotCleanup {
	return &BlockSnapshotCleanup{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (b *BlockSnapshotCleanup) Cleanup(ctx context.Context, t *testing.T) error {
	t.Helper()

	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&block.ListSnapshotsRequest{
		Name:      &b.namePrefix,
		Zone:      b.zone,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list block snapshots: %w", err)
	}

	if len(resp.Snapshots) == 0 {
		return fmt.Errorf("could not find any block snapshot prefixed with %q", b.namePrefix)
	}

	for _, snapshot := range resp.Snapshots {
		err = api.DeleteSnapshot(&block.DeleteSnapshotRequest{
			Zone:       b.zone,
			SnapshotID: snapshot.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to delete block snapshot: %w", err)
		}
	}

	t.Logf("deleted block snapshot %q\n", resp.Snapshots[0].Name)

	return nil
}
