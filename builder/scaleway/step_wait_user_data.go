package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepWaitUserData struct{}

func (s *stepWaitUserData) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	ui.Say("Waiting for any user data apply to finish if provided...")

	time.Sleep(c.UserDataTimeout)

	return multistep.ActionContinue
}

func (s *stepWaitUserData) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
