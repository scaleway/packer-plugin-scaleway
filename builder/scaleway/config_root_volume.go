//go:generate go tool packer-sdc struct-markdown
//go:generate go tool packer-sdc mapstructure-to-hcl2 -type ConfigRootVolume

package scaleway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// ConfigRootVolume is the configuration for your server's root volume
type ConfigRootVolume struct {
	// The name of the root volume. Default packer-UUID
	Name string `mapstructure:"name" required:"false"`
	// The type of the root volume
	Type string `mapstructure:"type"`
	// IOPS of the root volume if using SBS, will only affect runtime. Image's volumes cannot have a configured IOPS.
	IOPS *uint32 `mapstructure:"iops"`
	// Size of the root volume
	SizeInGB uint64 `mapstructure:"size_in_gb"`
	// The name of the resulting snapshot that will appear in your account. Default packer-TIMESTAMP
	SnapshotName string `mapstructure:"snapshot_name" required:"false"`
}

// Prepare sets the default names of the root volume and its snapshot
func (c *ConfigRootVolume) Prepare() {
	if c.Name == "" {
		c.Name = "packer-" + uuid.TimeOrderedUUID()
	}

	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		c.SnapshotName = def
	}
}

// IsConfigured returns true if root volume has been manually configured.
// If true, the volume template should be used when creating the server.
func (c *ConfigRootVolume) IsConfigured() bool {
	return c.Type != "" || c.IOPS != nil || c.SizeInGB != 0
}

// VolumeServerTemplate returns the template to create the volume in a CreateServerRequest
func (c *ConfigRootVolume) VolumeServerTemplate() *instance.VolumeServerTemplate {
	// Name is not set here: a named template makes the API create a new volume instead of using the image's one.
	// The volume is renamed in PostServerCreationSetup instead.
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

// PostServerCreationSetup renames the root volume and applies its IOPS once the server is created
func (c *ConfigRootVolume) PostServerCreationSetup(ctx context.Context, instanceAPI *instance.API, blockAPI *block.API, server *instance.Server, timeout time.Duration) error {
	rootVolume, exists := server.Volumes["0"]
	if !exists {
		return errors.New("root volume not found")
	}

	rename := rootVolume.Name == nil || *rootVolume.Name != c.Name

	if rootVolume.VolumeType != instance.VolumeServerVolumeTypeSbsVolume {
		if !rename {
			return nil
		}

		_, err := instanceAPI.UpdateVolume(&instance.UpdateVolumeRequest{
			Zone:     rootVolume.Zone,
			VolumeID: rootVolume.ID,
			Name:     &c.Name,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to rename root volume: %w", err)
		}

		return nil
	}

	if !rename && c.IOPS == nil {
		return nil
	}

	_, err := blockAPI.WaitForVolume(&block.WaitForVolumeRequest{
		VolumeID: rootVolume.ID,
		Zone:     rootVolume.Zone,
		Timeout:  &timeout,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to wait for root volume: %w", err)
	}

	req := &block.UpdateVolumeRequest{
		Zone:     rootVolume.Zone,
		VolumeID: rootVolume.ID,
		PerfIops: c.IOPS,
	}
	if rename {
		req.Name = &c.Name
	}

	_, err = blockAPI.UpdateVolume(req, scw.WithContext(ctx))

	return err
}
