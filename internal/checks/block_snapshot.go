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

var (
	_ tester.PackerCheck = (*BlockSnapshotCheck)(nil)
	_ SnapshotCheck      = (*BlockSnapshotCheck)(nil)
)

type BlockSnapshotCheck struct {
	zone       scw.Zone
	namePrefix string

	snapshotName *string
	tags         []string
	size         *scw.Size
}

func BlockSnapshot(zone scw.Zone, namePrefix string) *BlockSnapshotCheck {
	return &BlockSnapshotCheck{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (c *BlockSnapshotCheck) Name(name string) *BlockSnapshotCheck {
	c.snapshotName = &name

	return c
}

func (c *BlockSnapshotCheck) Tags(tags []string) *BlockSnapshotCheck {
	c.tags = tags

	return c
}

func (c *BlockSnapshotCheck) SizeInGB(size uint64) *BlockSnapshotCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func findBlockSnapshots(ctx context.Context, zone scw.Zone, namePrefix string) ([]*block.Snapshot, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&block.ListSnapshotsRequest{
		Zone:      zone,
		Name:      &namePrefix,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error listing block snapshots: %w", err)
	}

	if len(resp.Snapshots) == 0 {
		return nil, fmt.Errorf("could not find any block snapshot prefixed with %q", namePrefix)
	}

	return resp.Snapshots, nil
}

func (c *BlockSnapshotCheck) compareSingleBlockSnapshot(snapshot *block.Snapshot) error {
	if c.snapshotName != nil && snapshot.Name != *c.snapshotName {
		return fmt.Errorf("snapshot name %q does not match expected snapshot name %q", snapshot.Name, *c.snapshotName)
	}

	if len(c.tags) > 0 && !reflect.DeepEqual(c.tags, snapshot.Tags) {
		return fmt.Errorf("snapshot tags did not match, expected %v, got %v", c.tags, snapshot.Tags)
	}

	if c.size != nil && snapshot.Size != *c.size {
		return fmt.Errorf("snapshot size %d does not match expected size %d", snapshot.Size, *c.size)
	}

	return nil
}

func (c *BlockSnapshotCheck) CheckName() string {
	return fmt.Sprintf("Block snapshot \"%s...\"", c.namePrefix)
}

func (c *BlockSnapshotCheck) Check(ctx context.Context) error {
	snapshots, err := findBlockSnapshots(ctx, c.zone, c.namePrefix)
	if err != nil {
		return err
	}

	snapshotMatchingErrors := []error(nil)

	for _, snapshot := range snapshots {
		err = c.compareSingleBlockSnapshot(snapshot)
		if err != nil {
			snapshotMatchingErrors = append(snapshotMatchingErrors, err)
		}
	}

	if len(snapshotMatchingErrors) < len(snapshots) {
		return nil
	}

	return fmt.Errorf("no block snapshot matched the expected specs, got the following matching errors:\n%w", errors.Join(snapshotMatchingErrors...))
}
