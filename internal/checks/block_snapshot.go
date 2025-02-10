package checks

import (
	"context"
	"errors"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type BlockSnapshotCheck struct {
	zone       scw.Zone
	snapshotID *string

	size *scw.Size
}

func (c *BlockSnapshotCheck) SizeInGB(size uint64) *BlockSnapshotCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *BlockSnapshotCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	if c.snapshotID == nil {
		return errors.New("snapshot ID is required")
	}

	snapshot, err := api.GetSnapshot(&block.GetSnapshotRequest{
		Zone:       c.zone,
		SnapshotID: *c.snapshotID,
	})
	if err != nil {
		return fmt.Errorf("error getting snapshot %s: %w", *c.snapshotID, err)
	}

	if c.size != nil && snapshot.Size != *c.size {
		return fmt.Errorf("snapshot size %d does not match expected size %d", snapshot.Size, *c.size)
	}

	return nil
}

// BlockSnapshot returns an empty check, to be passed to another check to fill ID and zone
func BlockSnapshot() *BlockSnapshotCheck {
	return &BlockSnapshotCheck{}
}
