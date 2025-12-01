package scaleway

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func putErrorAndHalt(state multistep.StateBag, ui packersdk.Ui, err error) multistep.StepAction {
	state.Put("error", err)

	if ui != nil {
		ui.Error(err.Error())
	}

	return multistep.ActionHalt
}
