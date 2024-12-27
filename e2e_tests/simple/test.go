package main

import (
	"context"
	"e2e_tests/checks"
	"e2e_tests/tester"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func main() {
	zone := scw.ZoneFrPar1

	tester.Run(context.Background(),
		checks.Image(zone, "temp-build-packer").
			RootVolumeType("unified"),
		checks.NoVolume(zone),
	)
}
