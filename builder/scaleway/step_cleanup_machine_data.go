package scaleway

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCleanupMachineData struct{}

// Machine ID file locations
var (
	sysdID = "/etc/machine-id"
	dbusID = "/var/lib/dbus/machine-id"
)

func (s *stepCleanupMachineData) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packersdk.Communicator)
	c := state.Get("config").(*Config)
	cmd := new(packersdk.RemoteCmd)

	str, err := strconv.ParseBool(c.CleanupMachineRelatedData)
	if err != nil {
		err := fmt.Errorf("value must be: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if !str {
		return multistep.ActionContinue
	}

	ui.Say("Trying to remove machine-related data...")

	// Remove the machine-id file under /etc
	cmd.Command = fmt.Sprintf("sudo truncate -s 0 %s", sysdID)
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		log.Printf("Error cleaning up %s: %s", sysdID, err)
	}

	// Remove the machine-id file under /var/lib/dbus
	cmd = new(packersdk.RemoteCmd)
	cmd.Command = fmt.Sprintf("sudo truncate -s 0 %s", dbusID)
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		log.Printf("Error cleaning up %s: %s", dbusID, err)
	}

	return multistep.ActionContinue
}

func (s *stepCleanupMachineData) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
