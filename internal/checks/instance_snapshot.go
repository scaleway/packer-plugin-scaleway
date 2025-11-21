package checks

import (
	"context"
	"fmt"
	"reflect"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*InstanceSnapshotCheck)(nil)
var _ SnapshotCheck = (*InstanceSnapshotCheck)(nil)

type InstanceSnapshotCheck struct {
	zone       scw.Zone
	namePrefix string

	snapshotName *string
	tags         []string
	size         *scw.Size
}

func InstanceSnapshot(zone scw.Zone, namePrefix string) *InstanceSnapshotCheck {
	return &InstanceSnapshotCheck{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (c *InstanceSnapshotCheck) Name(name string) *InstanceSnapshotCheck {
	c.snapshotName = &name

	return c
}

func (c *InstanceSnapshotCheck) Tags(tags []string) *InstanceSnapshotCheck {
	c.tags = tags

	return c
}

func (c *InstanceSnapshotCheck) SizeInGB(size uint64) *InstanceSnapshotCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func findInstanceSnapshot(ctx context.Context, zone scw.Zone, namePrefix string) (*instance.Snapshot, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&instance.ListSnapshotsRequest{
		Zone:    zone,
		Name:    &namePrefix,
		Project: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error listing instance snapshots: %v", err)
	}

	if len(resp.Snapshots) == 0 {
		return nil, fmt.Errorf("could not find any instance snapshot prefixed with %q", namePrefix)
	}

	if len(resp.Snapshots) > 1 {
		return nil, fmt.Errorf("multiple instance snapshots found with prefix %q", namePrefix)
	}

	return resp.Snapshots[0], nil
}

func (c *InstanceSnapshotCheck) CheckName() string {
	return fmt.Sprintf("Instance snapshot \"%s...\"", c.namePrefix)
}

func (c *InstanceSnapshotCheck) Check(ctx context.Context) error {
	snapshot, err := findInstanceSnapshot(ctx, c.zone, c.namePrefix)
	if err != nil {
		return err
	}

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
