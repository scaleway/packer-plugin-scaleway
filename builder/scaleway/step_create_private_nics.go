package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepCreatePrivateNICs struct{}

func (s *stepCreatePrivateNICs) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	serverID := state.Get("server_id").(string)

	if len(c.PrivateNetworkIDs) == 0 {
		return multistep.ActionContinue
	}

	for _, pnID := range c.PrivateNetworkIDs {
		ui.Say(fmt.Sprintf("Attaching private network %s...", pnID))

		nicResp, err := instanceAPI.CreatePrivateNIC(&instance.CreatePrivateNICRequest{
			Zone:             scw.Zone(c.Zone),
			ServerID:         serverID,
			PrivateNetworkID: pnID,
		}, scw.WithContext(ctx))
		if err != nil {
			return putErrorAndHalt(state, ui, fmt.Errorf("error attaching private network %s: %w", pnID, err))
		}

		_, err = instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
			ServerID:     serverID,
			PrivateNicID: nicResp.PrivateNic.ID,
			Zone:         scw.Zone(c.Zone),
		}, scw.WithContext(ctx))
		if err != nil {
			return putErrorAndHalt(state, ui, fmt.Errorf("error waiting for private NIC %s: %w", nicResp.PrivateNic.ID, err))
		}

		ui.Say(fmt.Sprintf("Private network %s attached successfully", pnID))
	}

	return multistep.ActionContinue
}

func (s *stepCreatePrivateNICs) Cleanup(_ multistep.StateBag) {
	// Private NICs are automatically removed when the server is deleted.
}
