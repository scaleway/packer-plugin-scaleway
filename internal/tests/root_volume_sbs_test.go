package tests_test

import (
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestRootVolumeSBS(t *testing.T) {
	zone := scw.ZoneFrPar1

	tester.Test(t, &tester.TestConfig{
		Config: `
source "scaleway" "basic" {
  communicator = "none"
  commercial_type = "PLAY2-PICO"
  zone = "fr-par-1"
  image = "ubuntu_jammy"
  image_name = "packer-e2e-root-volume-sbs"
  ssh_username = "root"
  remove_volume = true

  root_volume {
    size_in_gb = 50
    iops = 15000
  }
}

build {
  sources = ["source.scaleway.basic"]
}
`,
		Checks: []tester.PackerCheck{
			checks.Image(zone, "packer-e2e-root-volume-sbs").
				RootVolumeType("sbs_snapshot").
				RootVolumeBlockSnapshot(checks.BlockSnapshot().SizeInGB(50)),
			checks.NoVolume(zone),
		},
	})
}
