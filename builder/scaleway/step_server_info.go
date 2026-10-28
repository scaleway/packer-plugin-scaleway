package scaleway

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
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
		return putErrorAndHalt(state, ui, fmt.Errorf("error waiting for server to become booted: %w", err))
	}

	if instanceResp.State != instance.ServerStateRunning {
		return putErrorAndHalt(state, ui, fmt.Errorf("server is in state %s", instanceResp.State.String()))
	}

	if instanceResp.PublicIP == nil && !c.DisablePublicIP {
		return putErrorAndHalt(state, ui, errors.New("server does not have a public IP"))
	}

	for _, pNIC := range instanceResp.PrivateNics {
		ui.Say(fmt.Sprintf("PrivateNICs %s ", pNIC.ID))
	}

	if c.DisablePublicIP {
		if len(instanceResp.PrivateNics) == 0 ||
			len(instanceResp.PrivateNics[0].IpamIPIDs) == 0 {
			return putErrorAndHalt(state, ui, errors.New("no private IP found for instance"))
		}

		zone := scw.Zone(c.Zone)

		region, err := zone.Region()
		if err != nil {
			return putErrorAndHalt(state, ui, fmt.Errorf("getting region from zone %s: %w", zone, err))
		}

		ipamAPI := ipam.NewAPI(state.Get("client").(*scw.Client))

		var firstPrivateIP string

		for _, nic := range instanceResp.PrivateNics {
			for _, ipID := range nic.IpamIPIDs {
				ipResp, err := ipamAPI.GetIP(&ipam.GetIPRequest{
					Region: region,
					IPID:   ipID,
				}, scw.WithContext(ctx))
				if err != nil {
					return putErrorAndHalt(state, ui, fmt.Errorf("error getting IP %s: %w", ipID, err))
				}

				ui.Say(fmt.Sprintf("NIC %s — IP %s", nic.ID, ipResp.Address))

				// Mémoriser la première IP récupérée.
				if firstPrivateIP == "" {
					firstPrivateIP = ipResp.Address.IP.String()
				}
			}
		}

		if firstPrivateIP == "" {
			return putErrorAndHalt(state, ui, errors.New("no private IP found for instance"))
		}

		state.Put("server_ip", firstPrivateIP)
	} else {
		state.Put("server_ip", instanceResp.PublicIP.Address.String())
	}

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
