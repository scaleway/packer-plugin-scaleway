package scaleway

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepServerInfo struct{}

func (s *stepServerInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Waiting for server to become active...")

	instanceResp, err := instanceAPI.WaitForServer(&instance.WaitForServerRequest{
		ServerID: serverID,
		Zone:     scw.Zone(c.Zone),
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error waiting for server to become booted: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instanceResp.State != instance.ServerStateRunning {
		err := fmt.Errorf("server is in state %s", instanceResp.State.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instanceResp.PublicIP == nil {
		err := errors.New("server does not have a public IP")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("server_ip", instanceResp.PublicIP.Address.String())
	state.Put("server", instanceResp)

	volumes := []*instance.VolumeServer(nil)
	for _, volume := range instanceResp.Volumes {
		volumes = append(volumes, volume)
	}

	state.Put("volumes", volumes)

	return multistep.ActionContinue
}

func (s *stepServerInfo) Cleanup(_ multistep.StateBag) {
	// no cleanup
}
