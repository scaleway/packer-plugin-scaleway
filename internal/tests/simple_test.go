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
)

func TestSimple(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-simple"

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "PRO2-XXS"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = true
			  tags = ["devtools", "provider", "packer"]
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				Tags([]string{"devtools", "provider", "packer"}).
				RootVolumeSnapshot(checks.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix).
					Tags([]string{"devtools", "provider", "packer"}),
				),
			checks.NoVolume(zone),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix),
		},
	})
}
