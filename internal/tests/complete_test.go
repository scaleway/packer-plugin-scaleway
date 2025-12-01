package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestComplete(t *testing.T) {
	zone := scw.ZoneNlAms1
	imageName := "packer-e2e-complete"
	rootVolumeSize := 12
	blockVolumeSize := 10

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "DEV1-M"
			  zone = "%[1]s"
			  image = "ubuntu_jammy"
			  image_name = "%[2]s"
			  ssh_username = "root"
			  remove_volume = false
              tags = [ "test", "packer", "complete" ]

			  root_volume {
				type = "l_ssd"
			    snapshot_name = "named-snap-root-volume-0"
			    size_in_gb = %[3]d
			  }

			  block_volume {
				name = "named-extra-volume-1"
			    size_in_gb = %[4]d
			    iops = 5000
			  }

			  block_volume {
			    size_in_gb = %[4]d
			    iops = 15000
			  }

			  block_volume {
			    snapshot_name = "named-snap-extra-volume-3"
			    size_in_gb = %[4]d
			    iops = 5000
			  }
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, rootVolumeSize, blockVolumeSize),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				SizeInGB(42).
				Tags([]string{"test", "packer", "complete"}).
				RootVolumeSnapshot(
					checks.InstanceSnapshot(zone, "named-snap-root").
						SizeInGB(uint64(rootVolumeSize)).
						Name("named-snap-root-volume-0").
						Tags([]string{"test", "packer", "complete"}),
				).
				ExtraVolumeSnapshot("1", checks.BlockSnapshot(zone, packerGeneratedResourceNamePrefix).
					SizeInGB(uint64(blockVolumeSize)).
					Tags([]string{"test", "packer", "complete"}),
				).
				ExtraVolumeSnapshot("2", checks.BlockSnapshot(zone, packerGeneratedResourceNamePrefix).
					SizeInGB(uint64(blockVolumeSize)).
					Tags([]string{"test", "packer", "complete"}),
				).
				ExtraVolumeSnapshot("3", checks.BlockSnapshot(zone, "named-snap-extra").
					SizeInGB(uint64(blockVolumeSize)).
					Name("named-snap-extra-volume-3").
					Tags([]string{"test", "packer", "complete"}),
				),
			checks.InstanceVolume(zone, rootVolumeNamePrefix).
				SizeInGB(12).
				Name(rootVolumeFromUbuntuJammyNamePrefix).
				Tags([]string{"test", "packer", "complete"}),
			checks.BlockVolume(zone, "named-extra-volume").
				SizeInGB(uint64(blockVolumeSize)).
				Name("named-extra-volume-1").
				IOPS(5000).
				Tags([]string{"test", "packer", "complete"}),
			checks.BlockVolume(zone, packerGeneratedResourceNamePrefix).
				SizeInGB(uint64(blockVolumeSize)).
				IOPS(15000).
				Tags([]string{"test", "packer", "complete"}),
			checks.BlockVolume(zone, packerGeneratedResourceNamePrefix).
				SizeInGB(uint64(blockVolumeSize)).
				IOPS(5000).
				Tags([]string{"test", "packer", "complete"}),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.InstanceSnapshot(zone, "named"),
			cleanup.InstanceVolume(zone, rootVolumeNamePrefix),
			cleanup.BlockSnapshot(zone, packerGeneratedResourceNamePrefix),
			cleanup.BlockVolume(zone, packerGeneratedResourceNamePrefix),
			cleanup.BlockSnapshot(zone, "named"),
			cleanup.BlockVolume(zone, "named"),
		},
	})
}
