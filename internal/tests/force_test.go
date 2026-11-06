package tests_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestForceRemovesOldImage(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-force"
	config := fmt.Sprintf(`
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
		`, zone, imageName, tagDevtools, tagProvider, tagPacker)

	tester.Test(t, &tester.TestConfig{
		Steps: []tester.TestStep{
			{
				Config: config,
				Checks: []tester.PackerCheck{
					checks.Image(zone, imageName),
				},
			},
			{
				Config:      config,
				ExpectError: regexp.MustCompile("is used by existing image"),
			},
			{
				Config: config,
				Args:   []string{"-force"},
				Checks: []tester.PackerCheck{
					// Fails if the image of the first step is still there
					checks.Image(zone, imageName),
					checks.NoVolume(zone),
				},
			},
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix),
		},
	})
}
