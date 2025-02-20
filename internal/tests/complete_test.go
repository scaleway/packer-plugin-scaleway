package tests_test

import (
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestComplete(t *testing.T) {
	t.Skip("snapshot_name argument does not work")

	zone := scw.ZoneFrPar2

	tester.Test(t, &tester.TestConfig{
		Config: `
source "scaleway" "basic" {
  communicator = "none"
  commercial_type = "PLAY2-PICO"
  zone = "fr-par-2"
  image = "ubuntu_jammy"
  image_name = "packer-e2e-complete"
  ssh_username = "root"

  remove_volume = false
  image_size_in_gb = 42
  snapshot_name = "packer-e2e-complete-snapshot"
}

build {
  sources = ["source.scaleway.basic"]
}
`,
		Checks: []tester.PackerCheck{
			checks.Image(zone, "packer-e2e-complete").
				RootVolumeType("unified").
				SizeInGb(42),
			checks.Snapshot(zone, "packer-e2e-complete-snapshot").
				SizeInGB(42),
			checks.Volume(zone, "Ubuntu 22.04 Jammy Jellyfish").
				SizeInGB(42),
		},
	})
}
