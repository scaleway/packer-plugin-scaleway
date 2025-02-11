package tests_test

import (
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestRootVolume(t *testing.T) {
	zone := scw.ZoneFrPar1

	tester.Test(t, &tester.TestConfig{
		Config: `
source "scaleway" "basic" {
  communicator = "none"
  commercial_type = "PLAY2-PICO"
  zone = "fr-par-1"
  image = "ubuntu_jammy"
  image_name = "packer-e2e-root-volume"
  ssh_username = "root"
  remove_volume = true

  root_volume {
    type = "b_ssd"
    size_in_gb = 50
  }
}

build {
  sources = ["source.scaleway.basic"]
}
`,
		Checks: []tester.PackerCheck{
			checks.Image(zone, "packer-e2e-root-volume").
				RootVolumeType("b_ssd").
				SizeInGb(50),
			checks.NoVolume(zone),
		},
	})
}
