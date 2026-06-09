package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/packer-plugin-scaleway/internal/vcr"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/require"
)

func TestComplete(t *testing.T) {
	zone := scw.ZoneNlAms1
	imageName := "packer-e2e-complete"
	serverName := "packer-tmp-server"
	rootVolumeSize := 12
	blockVolumeSize := 10

	httpClient, vcrCleanupFunc, err := vcr.GetHTTPRecorder(vcr.GetTestFilePath(t, "."), vcr.UpdateCassettes)
	require.NoError(t, err)

	defer vcrCleanupFunc()

	tester.Test(t, httpClient, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "DEV1-M"
			  zone = "%[1]s"
			  image = "ubuntu_jammy"
			  ssh_username = "root"

			  image_name = "%[2]s"
              tags = [ "%[5]s", "%[6]s", "%[7]s" ]
			  server_name = "%[8]s"
			  server_tags = [ "packer-build", "tmp-server" ]
			  remove_volume = false
			  keep_server = true

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
			`, zone, imageName, rootVolumeSize, blockVolumeSize, tagTest, tagPacker, tagComplete, serverName),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				SizeInGB(42).
				Tags(e2eTagsComplete).
				RootVolumeSnapshot(
					checks.InstanceSnapshot(zone, "named-snap-root").
						SizeInGB(uint64(rootVolumeSize)).
						Name("named-snap-root-volume-0").
						Tags(e2eTagsComplete),
				).
				ExtraVolumeSnapshot("1", checks.BlockSnapshot(zone, packerGeneratedResourceNamePrefix).
					SizeInGB(uint64(blockVolumeSize)).
					Tags(e2eTagsComplete),
				).
				ExtraVolumeSnapshot("2", checks.BlockSnapshot(zone, packerGeneratedResourceNamePrefix).
					SizeInGB(uint64(blockVolumeSize)).
					Tags(e2eTagsComplete),
				).
				ExtraVolumeSnapshot("3", checks.BlockSnapshot(zone, "named-snap-extra").
					SizeInGB(uint64(blockVolumeSize)).
					Name("named-snap-extra-volume-3").
					Tags(e2eTagsComplete),
				),
			checks.InstanceVolume(zone, rootVolumeNamePrefix).
				SizeInGB(12).
				Name(rootVolumeFromUbuntuJammyNamePrefix).
				Tags(e2eTagsComplete),
			checks.BlockVolume(zone, "named-extra-volume").
				SizeInGB(uint64(blockVolumeSize)).
				Name("named-extra-volume-1").
				IOPS(5000).
				Tags(e2eTagsComplete),
			checks.BlockVolume(zone, packerGeneratedResourceNamePrefix).
				SizeInGB(uint64(blockVolumeSize)).
				IOPS(15000).
				Tags(e2eTagsComplete),
			checks.BlockVolume(zone, packerGeneratedResourceNamePrefix).
				SizeInGB(uint64(blockVolumeSize)).
				IOPS(5000).
				Tags(e2eTagsComplete),
			checks.Server(zone, serverName).
				Tags([]string{"packer-build", "tmp-server"}),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Server(zone, serverName),
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
