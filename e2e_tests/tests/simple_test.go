package tests_test

import (
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/e2e_tests/checks"
	"github.com/scaleway/packer-plugin-scaleway/e2e_tests/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestSimple(t *testing.T) {
	zone := scw.ZoneFrPar1

	tester.Test(t, &tester.TestConfig{
		Config: `
source "scaleway" "basic" {
  communicator = "none"
  commercial_type = "PRO2-XXS"
  zone = "fr-par-1"
  image = "ubuntu_jammy"
  image_name = "packer-e2e-simple"
  ssh_username = "root"
  remove_volume = true
}

build {
  sources = ["source.scaleway.basic"]
}
`,
		Checks: []tester.PackerCheck{
			checks.Image(zone, "packer-e2e-simple").
				RootVolumeType("unified"),
			checks.NoVolume(zone),
		},
	})
}
