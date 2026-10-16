package scaleway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/internal/httperrors"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
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

	// Add specific server_tags if provided, else apply the tags
	tags := c.ServerTags
	if len(tags) == 0 {
		tags = c.Tags
	}

	ui.Say("Creating server...")

	if c.Comm.SSHPublicKey != nil {
		tags = append(tags, "AUTHORIZED_KEY="+strings.ReplaceAll(strings.TrimSpace(string(c.Comm.SSHPublicKey)), " ", "_"))
	}

	bootType := instance.BootType(c.BootType)

	createServerReq := &instance.CreateServerRequest{
		BootType:       &bootType,
		CommercialType: c.CommercialType,
		Name:           c.ServerName,
		Image:          scw.StringPtr(c.Image),
		Tags:           tags,
		Zone:           scw.Zone(c.Zone),
	}

	if c.RootVolume.IsConfigured() {
		if createServerReq.Volumes == nil {
			createServerReq.Volumes = make(map[string]*instance.VolumeServerTemplate)
		}

		createServerReq.Volumes["0"] = c.RootVolume.VolumeServerTemplate()
	}

	if len(c.BlockVolumes) > 0 {
		if createServerReq.Volumes == nil {
			createServerReq.Volumes = make(map[string]*instance.VolumeServerTemplate)
		}

		createdVolumes := state.Get(StateKeyCreatedVolumes).([]*instance.VolumeServerTemplate)
		for i, blockVolume := range createdVolumes {
			volumeIndex := strconv.FormatInt(int64(i+1), 10)
			createServerReq.Volumes[volumeIndex] = blockVolume
		}
	}

	if c.AdminPasswordEncryptionSSHKeyID != "" {
		createServerReq.AdminPasswordEncryptionSSHKeyID = &c.AdminPasswordEncryptionSSHKeyID
	}

	createServerResp, err := createServer(instanceAPI, createServerReq, scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error creating server: %w", formatInstanceError(err)))
	}

	// Track the server as soon as it exists so that every later halt
	// destroys it in Cleanup. server_id stays in the state until Cleanup
	// deletes the server, which tells stepCreateVolume.Cleanup whether a
	// server still holds its volumes.
	s.serverID = createServerResp.Server.ID
	state.Put("server_id", s.serverID)

	err = c.RootVolume.PostServerCreationSetup(block.NewAPI(state.Get("client").(*scw.Client)), createServerResp.Server)
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error during post server creation setup server: %w", err))
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
			return putErrorAndHalt(state, ui, fmt.Errorf("error creating server: %w", err))
		}
	}

	server, err := startServer(ctx, instanceAPI, c, createServerResp.Server.ID)
	if err == nil && server.State == instance.ServerStateStopped {
		// The poweron was accepted, then failed asynchronously without
		// recording a reason on the server. Out of stock is the usual cause,
		// and a new poweron reports it synchronously, or starts the server if
		// capacity came back.
		ui.Say("Server stopped right after poweron, retrying once...")

		server, err = startServer(ctx, instanceAPI, c, createServerResp.Server.ID)
	}

	if err != nil {
		return putErrorAndHalt(state, ui, err)
	}

	if server.State != instance.ServerStateRunning {
		return putErrorAndHalt(state, ui, fmt.Errorf("server %s is %s instead of running", server.ID, server.State))
	}

	// instance_id is the generic term used so that users can have access to the
	// instance id inside the provisioners, used in step_provision.
	state.Put("instance_id", s.serverID)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	c := state.Get("config").(*Config)
	if s.serverID == "" || c.KeepServer {
		return
	}

	client := state.Get("client").(*scw.Client)
	instanceAPI := instance.NewAPI(client)
	ui := state.Get("ui").(packersdk.Ui)
	zone := scw.Zone(c.Zone)

	// Once the image exists, stepCreateVolume.Cleanup removes every server
	// volume. Before that, the volumes created along with the server (root
	// volume, extra volumes of the source image) are removed here, and
	// stepCreateVolume.Cleanup removes the block volumes it created.
	var serverVolumes []*instance.VolumeServer

	if _, imageCreated := state.GetOk("snapshots"); !imageCreated {
		// A halt can leave the server in a transient state (start timeout,
		// cancellation) in which both delete and terminate are refused.
		server, err := instanceAPI.WaitForServer(&instance.WaitForServerRequest{
			ServerID: s.serverID,
			Zone:     zone,
		})
		if err != nil && !httperrors.Is404(err) {
			ui.Error(fmt.Sprintf("Error waiting for server %s to settle, its volumes may need manual removal: %s", s.serverID, err))
		}

		if err == nil && c.RemoveVolume {
			serverVolumes = volumesCreatedWithServer(state, server)
		}
	}

	ui.Say("Destroying server...")

	terminated := false

	deleteErr := instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		ServerID: s.serverID,
		Zone:     zone,
	})
	if deleteErr != nil && !httperrors.Is404(deleteErr) {
		_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
			Action:   instance.ServerActionTerminate,
			ServerID: s.serverID,
			Zone:     zone,
		})
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting server: %s", deleteErr))
			ui.Error(fmt.Sprintf(
				"Error terminating server. Please destroy it manually: %s", err))

			return
		}

		terminated = true
	}

	state.Remove("server_id")

	for _, volume := range serverVolumes {
		isLocal := volume.VolumeType == instance.VolumeServerVolumeTypeLSSD ||
			volume.VolumeType == instance.VolumeServerVolumeTypeScratch
		if terminated && isLocal {
			// terminate deletes local volumes along with the server.
			continue
		}

		ui.Say(fmt.Sprintf("Removing volume %s...", volume.ID))

		var err error
		if volume.VolumeType == instance.VolumeServerVolumeTypeSbsVolume {
			// Waits for the volume to be detached from the deleted server.
			err = waitAndDeleteBlockVolume(block.NewAPI(client), zone, volume.ID)
		} else {
			err = waitAndDeleteInstanceVolume(instanceAPI, zone, volume.ID)
		}

		if err != nil && !httperrors.Is404(err) {
			ui.Error(fmt.Sprintf("Error removing volume %s. Please remove it manually: %s", volume.ID, err))
		}
	}
}

// volumesCreatedWithServer returns the server volumes, except the block
// volumes created by stepCreateVolume, which its own Cleanup removes.
func volumesCreatedWithServer(state multistep.StateBag, server *instance.Server) []*instance.VolumeServer {
	createdByStep := map[string]struct{}{}

	if createdVolumes, ok := state.GetOk(StateKeyCreatedVolumes); ok {
		for _, volume := range createdVolumes.([]*instance.VolumeServerTemplate) {
			createdByStep[*volume.ID] = struct{}{}
		}
	}

	volumes := make([]*instance.VolumeServer, 0, len(server.Volumes))

	for _, volume := range server.Volumes {
		if _, ok := createdByStep[volume.ID]; !ok {
			volumes = append(volumes, volume)
		}
	}

	return volumes
}

// startServer powers the server on and waits for it to settle in a terminal
// state, which can be stopped if the start failed.
func startServer(ctx context.Context, instanceAPI *instance.API, c *Config, serverID string) (*instance.Server, error) {
	_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
		Action:   instance.ServerActionPoweron,
		ServerID: serverID,
		Zone:     scw.Zone(c.Zone),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error starting server: %w", formatInstanceError(err))
	}

	server, err := instanceAPI.WaitForServer(&instance.WaitForServerRequest{
		ServerID: serverID,
		Zone:     scw.Zone(c.Zone),
		Timeout:  &c.ServerCreationTimeout,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("server is not available: %w", err)
	}

	return server, nil
}
