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

	for i, volume := range volumes {
		if volume.Name == "" {
			volume.Name = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
		}
		if volume.Size != 0 && volume.SnapshotID != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("volume (index: %d) can't have a snapshot_id and a size", i))
		}
		if volume.Size == 0 && volume.SnapshotID == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("volume (index: %d) must have a snapshot_id or a size", i))
		}
	}

	return errs
}

func (blockVolume *ConfigBlockVolume) VolumeTemplate() *instance.VolumeServerTemplate {
	return &instance.VolumeServerTemplate{
		Name:         &blockVolume.Name,
		Size:         scw.SizePtr(scw.Size(blockVolume.Size)),
		BaseSnapshot: &blockVolume.SnapshotID,
	}
}
