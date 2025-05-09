package scaleway

import (
	"fmt"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func prepareBlockVolumes(volumes []ConfigBlockVolume) *packersdk.MultiError {
	var errs *packersdk.MultiError

	for i := range volumes {
		volume := &volumes[i]

		if volume.Name == "" {
			volume.Name = "packer-" + uuid.TimeOrderedUUID()
		}

		if volume.SizeInGB != 0 && volume.SnapshotID != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("volume (index: %d) can't have a snapshot_id and a size", i))
		}

		if volume.SizeInGB == 0 && volume.SnapshotID == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("volume (index: %d) must have a snapshot_id or a size", i))
		}
	}

	return errs
}

func (blockVolume *ConfigBlockVolume) VolumeTemplate() *instance.VolumeServerTemplate {
	return &instance.VolumeServerTemplate{
		Name:         &blockVolume.Name,
		Size:         scw.SizePtr(scw.Size(blockVolume.SizeInGB) * scw.GB),
		BaseSnapshot: &blockVolume.SnapshotID,
	}
}
