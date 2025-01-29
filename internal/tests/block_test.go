package tests_test

import (
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestBlock(t *testing.T) {
	zone := scw.ZoneFrPar1

	tester.Test(t, &tester.TestConfig{
		Config: `
source "scaleway" "basic" {
  communicator = "none"
  commercial_type = "PRO2-XXS"
  zone = "fr-par-1"
  image = "ubuntu_jammy"
  image_name = "packer-e2e-block"
  ssh_username = "root"
  remove_volume = true

  block_volume {
    name = "packer-e2e-block-vol1"
    size_in_gb = 20
    iops = 5000
  }
}

build {
  sources = ["source.scaleway.basic"]
}
`,
		Checks: []tester.PackerCheck{
			checks.Image(zone, "packer-e2e-block").
				RootVolumeType("b_ssd").
				ExtraVolumeType("1", "sbs_snapshot"),
			checks.NoVolume(zone),
		},
	})
}
