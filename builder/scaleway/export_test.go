package scaleway

import "github.com/hashicorp/packer-plugin-sdk/multistep"

func NewStepCreateServer() multistep.Step { //nolint:ireturn
	return new(stepCreateServer)
}

func NewStepCreateVolume() multistep.Step { //nolint:ireturn
	return new(stepCreateVolume)
}
