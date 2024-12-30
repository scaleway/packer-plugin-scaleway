package main

import (
	"context"
	"e2e_tests/checks"
	"e2e_tests/tester"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func main() {
	zone := scw.ZoneFrPar2

	tester.Run(context.Background(),
		checks.Image(zone, "packer-e2e-complete").
			RootVolumeType("unified").
			SizeInGb(42),
		checks.Snapshot(zone, "packer-e2e-complete-snapshot").
			SizeInGB(42),
		checks.Volume(zone, "Ubuntu 22.04 Jammy Jellyfish").
			SizeInGB(42),
	)
}
