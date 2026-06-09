package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	rootVolumeNamePrefix                = "Ubuntu"
	rootVolumeFromUbuntuJammyNamePrefix = "Ubuntu 22.04 Jammy Jellyfish"
	packerGeneratedResourceNamePrefix   = "packer-"
	apiGeneratedSnapshotNamePrefix      = "snp-"
	tagTest                             = "test"
	tagPacker                           = "packer"
	tagComplete                         = "complete"
	tagDevtools                         = "devtools"
	tagProvider                         = "provider"
	tagSnapshotName                     = "snapshot-name"
	tagBlock                            = "block"
	tagLocal                            = "local"
)

var (
	e2eTagsComplete          = []string{tagTest, tagPacker, tagComplete}
	e2eTagsDevtools          = []string{tagDevtools, tagProvider, tagPacker}
	e2eTagsSnapshotNameBlock = []string{tagTest, tagSnapshotName, tagBlock}
	e2eTagsSnapshotNameLocal = []string{tagTest, tagSnapshotName, tagLocal}
)

func TestSimple(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-simple"

	tester.Test(t.Context(), t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "PRO2-XXS"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = true
			  tags = ["%s", "%s", "%s"]
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, tagDevtools, tagProvider, tagPacker),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				Tags(e2eTagsDevtools).
				RootVolumeSnapshot(checks.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix).
					Tags(e2eTagsDevtools),
				),
			checks.NoVolume(zone),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix),
		},
	})
}
