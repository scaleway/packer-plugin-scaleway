package main_test

import (
	"e2e_tests/checks"
	"e2e_tests/tester"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestSimple(t *testing.T) {
	zone := scw.ZoneFrPar1

	tester.Test(t, &tester.TestConfig{
		Checks: []tester.PackerCheck{
			checks.Image(zone, "packer-e2e-simple").
				RootVolumeType("unified"),
			checks.NoVolume(zone),
		},
	})
}
