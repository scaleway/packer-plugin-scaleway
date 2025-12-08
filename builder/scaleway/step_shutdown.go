package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepShutdown struct{}

func (s *stepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	serverID := state.Get("server_id").(string)

	ui.Say("Shutting down server...")

	_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
		Action:   instance.ServerActionPoweroff,
		ServerID: serverID,
		Zone:     scw.Zone(c.Zone),
	}, scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error stopping server: %w", err))
	}

	waitRequest := &instance.WaitForServerRequest{
		ServerID: serverID,
		Zone:     scw.Zone(c.Zone),
		Timeout:  &c.ServerShutdownTimeout,
	}

	instanceResp, err := instanceAPI.WaitForServer(waitRequest, scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error shutting down server: %w", err))
	}

	if instanceResp.State != instance.ServerStateStopped {
		return putErrorAndHalt(state, ui, fmt.Errorf("server is in state %s instead of stopped", instanceResp.State.String()))
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
