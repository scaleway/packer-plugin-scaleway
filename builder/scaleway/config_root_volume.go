package scaleway

import (
	"fmt"

	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// IsConfigured returns true if root volume has been manually configured.
// If true, the volume template should be used when creating the server.
func (c *ConfigRootVolume) IsConfigured() bool {
	return c.Type != "" || c.IOPS != nil || c.SizeInGB != 0
}

// VolumeServerTemplate returns the template to create the volume in a CreateServerRequest
func (c *ConfigRootVolume) VolumeServerTemplate() *instance.VolumeServerTemplate {
	tmpl := &instance.VolumeServerTemplate{}

	if c.Type != "" {
		tmpl.VolumeType = instance.VolumeVolumeType(c.Type)
	} else {
		tmpl.VolumeType = instance.VolumeVolumeTypeSbsVolume
	}

	if c.SizeInGB > 0 {
		tmpl.Size = scw.SizePtr(scw.Size(c.SizeInGB) * scw.GB)
	}

	return tmpl
}

func (c *ConfigRootVolume) PostServerCreationSetup(blockAPI *block.API, server *instance.Server) error {
	if c.IOPS != nil {
		rootVolume, exists := server.Volumes["0"]
		if !exists {
			return fmt.Errorf("root Volume not found")
		}
		_, err := blockAPI.UpdateVolume(&block.UpdateVolumeRequest{
			Zone:     rootVolume.Zone,
			VolumeID: rootVolume.ID,
			PerfIops: c.IOPS,
		})
		return err
	}

	return nil
}
