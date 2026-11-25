package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestSnapshotNameBlock(t *testing.T) {
	zone := scw.ZoneFrPar2
	imageName := "packer-e2e-snap-name-block"
	snapshotName := "named-block-snapshot"

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "PLAY2-PICO"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = false
              tags = [ "%s", "%s", "%s" ]

			  root_volume {
			    type = "sbs_volume"
			  	snapshot_name = "%s"
			  }
			}

			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, tagTest, tagSnapshotName, tagBlock, snapshotName),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				Tags(e2eTagsSnapshotNameBlock).
				RootVolumeSnapshot(
					checks.BlockSnapshot(zone, snapshotName).
						Name(snapshotName).
						Tags(e2eTagsSnapshotNameBlock),
				),
			checks.BlockVolume(zone, packerGeneratedResourceNamePrefix).
				Tags(e2eTagsSnapshotNameBlock),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, snapshotName),
			cleanup.BlockVolume(zone, packerGeneratedResourceNamePrefix),
		},
	})
}

func TestSnapshotNameLocal(t *testing.T) {
	zone := scw.ZoneFrPar2
	imageName := "packer-e2e-snap-name-local"
	snapshotName := "named-local-snapshot"

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "DEV1-S"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = false
              tags = [ "%s", "%s", "%s" ]

			  root_volume {
			    type = "l_ssd"
			  	snapshot_name = "%s"
			  }
			}

			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, tagTest, tagSnapshotName, tagLocal, snapshotName),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				Tags(e2eTagsSnapshotNameLocal).
				RootVolumeSnapshot(
					checks.InstanceSnapshot(zone, snapshotName).
						Name(snapshotName).
						Tags(e2eTagsSnapshotNameLocal),
				),
			checks.InstanceVolume(zone, packerGeneratedResourceNamePrefix).
				Tags(e2eTagsSnapshotNameLocal),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.InstanceSnapshot(zone, snapshotName),
			cleanup.InstanceVolume(zone, packerGeneratedResourceNamePrefix),
		},
	})
}
