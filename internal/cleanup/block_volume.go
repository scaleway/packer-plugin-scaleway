package cleanup

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*BlockVolumeCleanup)(nil)

type BlockVolumeCleanup struct {
	zone       scw.Zone
	namePrefix string
}

func BlockVolume(zone scw.Zone, namePrefix string) *BlockVolumeCleanup {
	return &BlockVolumeCleanup{
		zone:       zone,
		namePrefix: namePrefix,
	}
}

func (b *BlockVolumeCleanup) Cleanup(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&block.ListVolumesRequest{
		Name:      &b.namePrefix,
		Zone:      b.zone,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list block volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return fmt.Errorf("could not find any block volume prefixed with %q", b.namePrefix)
	}

	for _, volume := range resp.Volumes {
		err = api.DeleteVolume(&block.DeleteVolumeRequest{
			Zone:     b.zone,
			VolumeID: volume.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to delete block volume: %w", err)
		}
	}

	fmt.Printf("deleted block volume %q\n", resp.Volumes[0].Name)

	return nil
}
