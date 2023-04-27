package scaleway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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
		tags = []string{"AUTHORIZED_KEY=" + strings.ReplaceAll(strings.TrimSpace(string(c.Comm.SSHPublicKey)), " ", "_")}
	}

	bootType := instance.BootType(c.BootType)

	createServerReq := &instance.CreateServerRequest{
		BootType:       &bootType,
		Bootscript:     bootscript,
		CommercialType: c.CommercialType,
		Name:           c.ServerName,
		Image:          c.Image,
		Tags:           tags,
		Zone:           scw.Zone(c.Zone),
	}

	if c.ImageSizeInGB != 0 {
		size := scw.Size(c.ImageSizeInGB) * scw.GB
		createServerReq.Volumes = map[string]*instance.VolumeServerTemplate{
			"0": {
				VolumeType: instance.VolumeVolumeTypeBSSD,
				Size:       &size,
			},
		}
	}

	if len(c.BlockVolumes) > 0 {
		if createServerReq.Volumes == nil {
			createServerReq.Volumes = make(map[string]*instance.VolumeServerTemplate)
		}
		for i, blockVolume := range c.BlockVolumes {
			volumeIndex := strconv.FormatInt(int64(i+1), 10)
			createServerReq.Volumes[volumeIndex] = blockVolume.VolumeTemplate()
		}
	}

	createServerResp, err := instanceAPI.CreateServer(createServerReq, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(c.UserData) > 0 {
		m := map[string]io.Reader{}
		for k, v := range c.UserData {
			m[k] = bytes.NewBufferString(v)
		}

		createUserDataReq := &instance.SetAllServerUserDataRequest{
			Zone:     scw.Zone(c.Zone),
			ServerID: createServerResp.Server.ID,
			UserData: m,
		}

		err = instanceAPI.SetAllServerUserData(createUserDataReq, scw.WithContext(ctx))
		if err != nil {
			err := fmt.Errorf("error creating server: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
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

	waitServerRequest := &instance.WaitForServerRequest{
		ServerID: createServerResp.Server.ID,
		Zone:     scw.Zone(c.Zone),
		Timeout:  &c.ServerCreationTimeout,
	}
	server, err := instanceAPI.WaitForServer(waitServerRequest)
	if err != nil {
		err := fmt.Errorf("server is not available: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if server.State != instance.ServerStateRunning {
		err := errors.New("servert is in error state")
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
	if s.serverID == "" {
		return
	}

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Destroying server...")

	deleteErr := instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		ServerID: s.serverID,
	})
	if deleteErr != nil {
		_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
			Action:   instance.ServerActionTerminate,
			ServerID: s.serverID,
		})
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting server: %s", deleteErr))
			ui.Error(fmt.Sprintf(
				"Error terminating server. Please destroy it manually: %s", err))
		}
	}
}
