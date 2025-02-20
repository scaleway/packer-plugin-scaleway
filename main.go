package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
	"github.com/scaleway/packer-plugin-scaleway/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(scaleway.Builder))
	pps.SetVersion(version.PluginVersion)

	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
