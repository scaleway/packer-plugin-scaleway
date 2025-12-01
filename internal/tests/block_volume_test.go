package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestLocalWithSBSVolume(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-mixed-volumes"
	rootVolumeSize := 20
	blockVolumeSize := 50

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "GP1-S"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = false
			
			  root_volume {
				type = "l_ssd"
			    size_in_gb = %d
			  }
			
			  block_volume {
			    size_in_gb = %d
			    iops = 15000
			  }
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, rootVolumeSize, blockVolumeSize),
		Checks: []tester.PackerCheck{
			checks.InstanceVolume(zone, rootVolumeNamePrefix).
				SizeInGB(uint64(rootVolumeSize)),
			checks.BlockVolume(zone, packerGeneratedResourceNamePrefix).
				IOPS(15000).
				SizeInGB(uint64(blockVolumeSize)),
			checks.Image(zone, imageName).
				RootVolumeSnapshot(
					checks.InstanceSnapshot(zone, packerGeneratedResourceNamePrefix).
						SizeInGB(uint64(rootVolumeSize)),
				).
				ExtraVolumeSnapshot("1", checks.BlockSnapshot(zone, packerGeneratedResourceNamePrefix).
					SizeInGB(uint64(blockVolumeSize)),
				),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.InstanceSnapshot(zone, packerGeneratedResourceNamePrefix),
			cleanup.InstanceVolume(zone, rootVolumeNamePrefix),
			cleanup.BlockSnapshot(zone, packerGeneratedResourceNamePrefix),
			cleanup.BlockVolume(zone, packerGeneratedResourceNamePrefix),
		},
	})
}

func TestBlockOnly(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-block"
	volumeName := "volume-with-name"
	volumeSize := 20

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "PRO2-XXS"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = false
			
			  block_volume {
			    name = "%s"
			    size_in_gb = %d
			    iops = 5000
			  }
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, volumeName, volumeSize),
		Checks: []tester.PackerCheck{
			checks.BlockVolume(zone, rootVolumeNamePrefix).
				SizeInGB(uint64(10)),
			checks.BlockVolume(zone, volumeName).
				Name(volumeName).
				IOPS(5000).
				SizeInGB(uint64(volumeSize)),
			checks.Image(zone, imageName).
				SizeInGB(uint64(30)).
				RootVolumeSnapshot(checks.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix).
					SizeInGB(uint64(10))).
				ExtraVolumeSnapshot("1", checks.BlockSnapshot(zone, packerGeneratedResourceNamePrefix).
					SizeInGB(uint64(volumeSize)),
				),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix),
			cleanup.BlockVolume(zone, rootVolumeNamePrefix),
			cleanup.BlockSnapshot(zone, packerGeneratedResourceNamePrefix),
			cleanup.BlockVolume(zone, volumeName),
		},
	})
}
