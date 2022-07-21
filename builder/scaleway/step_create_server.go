package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepCreateServer struct {
	serverID string
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	var tags []string
	var bootscript *string

	ui.Say("Creating server...")

	if c.Bootscript != "" {
		bootscript = &c.Bootscript
	}

	if c.Comm.SSHPublicKey != nil {
		tags = []string{fmt.Sprintf("AUTHORIZED_KEY=%s", strings.Replace(strings.TrimSpace(string(c.Comm.SSHPublicKey)), " ", "_", -1))}
	}

	bootType := instance.BootType(c.BootType)

	createServerReq := &instance.CreateServerRequest{
		BootType:       &bootType,
		Bootscript:     bootscript,
		CommercialType: c.CommercialType,
		Name:           c.ServerName,
		Image:          c.Image,
		Tags:           tags,
	}

	if c.ImageSizeInGB != 0 {
		createServerReq.Volumes = map[string]*instance.VolumeServerTemplate{
			"0": {
				VolumeType: instance.VolumeVolumeTypeBSSD,
				Size:       scw.Size(c.ImageSizeInGB) * scw.GB,
			},
		}
	}

	createServerResp, err := instanceAPI.CreateServer(createServerReq, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	waitServerRequest := &instance.WaitForServerRequest{
		ServerID: createServerResp.Server.ID,
		Zone:     scw.Zone(c.Zone),
	}
	timeout := c.ServerTimeout
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		err := fmt.Errorf("error: %s could not parse string %s as a duration", err, timeout)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if timeout != "" {
		waitServerRequest.Timeout = scw.TimeDurationPtr(duration)
	}

	_, err = instanceAPI.ServerAction(&instance.ServerActionRequest{
		Action:   instance.ServerActionPoweron,
		ServerID: createServerResp.Server.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error starting server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	server, err := instanceAPI.WaitForServer(waitServerRequest)
	if err != nil {
		err := fmt.Errorf("server is not available: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if server.State != instance.ServerStateRunning {
		err := fmt.Errorf("servert is in error state")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.serverID = createServerResp.Server.ID

	state.Put("server_id", createServerResp.Server.ID)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside the provisioners, used in step_provision.
	state.Put("instance_id", s.serverID)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	volumeID := state.Get("root_volume_id").(string)

	if s.serverID == "" {
		return
	}

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Destroying server...")

	err := instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		ServerID: s.serverID,
	})
	if err != nil {
		_, err = instanceAPI.ServerAction(&instance.ServerActionRequest{
			Action:   instance.ServerActionTerminate,
			ServerID: s.serverID,
		})
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying server. Please destroy it manually: %s", err))
		}
	}

	ui.Say("Removing volume ...")

	err = instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: volumeID,
	})
	if err != nil {
		err := fmt.Errorf("error removing volume: %s", err)
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error removing volume: %s. Please destroy it manually", err))
	}
}
