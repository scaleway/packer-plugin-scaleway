package scaleway

import "github.com/hashicorp/packer-plugin-sdk/multistep"

func NewStepCreateVolume() multistep.Step { //nolint:ireturn
	return new(stepCreateVolume)
}
