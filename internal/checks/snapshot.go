package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type SnapshotCheck struct {
	zone         scw.Zone
	snapshotName string
	tags         []string

	size *scw.Size
}

func (c *SnapshotCheck) SizeInGB(size uint64) *SnapshotCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *SnapshotCheck) Tags(tags []string) *SnapshotCheck {
	c.tags = tags
	return c
}

func (c *SnapshotCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&instance.ListSnapshotsRequest{
		Zone:    c.zone,
		Name:    &c.snapshotName,
		Project: &testCtx.ProjectID,
	})
	if err != nil {
		return err
	}

	if len(resp.Snapshots) == 0 {
		return fmt.Errorf("snapshot %s not found, no snapshots found", c.snapshotName)
	}

	if len(resp.Snapshots) > 1 {
		return fmt.Errorf("multiple snapshots found with name %s", c.snapshotName)
	}

	snapshot := resp.Snapshots[0]

	if snapshot.Name != c.snapshotName {
		return fmt.Errorf("snapshot name %s does not match expected snapshot name %s", snapshot.Name, c.snapshotName)
	}

	if c.size != nil && snapshot.Size != *c.size {
		return fmt.Errorf("snapshot size %d does not match expected size %d", snapshot.Size, *c.size)
	}

	if len(c.tags) > 0 {
		for _, expectedTag := range c.tags {
			found := false
			for _, actualTag := range snapshot.Tags {
				if actualTag == expectedTag {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("expected tag %q not found on snapshot %s", expectedTag, c.snapshotName)
			}
		}
	}

	return nil
}

func Snapshot(zone scw.Zone, snapshotName string) *SnapshotCheck {
	return &SnapshotCheck{
		zone:         zone,
		snapshotName: snapshotName,
	}
}
