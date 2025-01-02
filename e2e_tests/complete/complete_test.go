package main_test

import (
	"e2e_tests/checks"
	"e2e_tests/tester"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestComplete(t *testing.T) {
	zone := scw.ZoneFrPar2

	tester.Test(t, &tester.TestConfig{
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
