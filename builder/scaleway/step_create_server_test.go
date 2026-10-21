package scaleway_test

import (
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// The API rejects a poweron as out of stock right away, or accepts it and the
// server falls back to stopped without any reason once node allocation fails.
func TestStepCreateServerFailedStartCleansUp(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		acceptedPoweron int
		wantRun         []string
		wantErr         string
	}{
		"rejected as out of stock": {
			acceptedPoweron: 0,
			wantRun:         []string{reqCreateServer, reqPoweron},
			wantErr:         "resource GP1-M is out of stock",
		},
		"accepted then stopped, retry rejected as out of stock": {
			acceptedPoweron: 1,
			wantRun:         []string{reqCreateServer, reqPoweron, reqGetServer, reqPoweron},
			wantErr:         "resource GP1-M is out of stock",
		},
		"accepted then stopped twice": {
			acceptedPoweron: 2,
			wantRun:         []string{reqCreateServer, reqPoweron, reqGetServer, reqPoweron, reqGetServer},
			wantErr:         "server " + fakeServerID + " stopped while allocating a node, usually because GP1-M is out of stock in fr-par-1",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			api := &fakeScalewayAPI{
				serverVolumes: map[string]*instance.VolumeServer{
					"0": {ID: fakeRootVolumeID, VolumeType: instance.VolumeServerVolumeTypeSbsVolume},
				},
				acceptedPoweron:   tc.acceptedPoweron,
				stateAfterPoweron: instance.ServerStateStopped,
			}
			state := newFakeState(t, api, &scaleway.Config{
				Zone:                  scw.ZoneFrPar1.String(),
				CommercialType:        "GP1-M",
				Image:                 "ubuntu_focal",
				ServerName:            "packer-test",
				BootType:              "local",
				RemoveVolume:          true,
				ServerCreationTimeout: time.Minute,
			})

			step := scaleway.NewStepCreateServer()

			if action := step.Run(t.Context(), state); action != multistep.ActionHalt {
				t.Fatalf("expected halt, got %v", action)
			}

			runErr, _ := state.Get("error").(error)
			if runErr == nil || !strings.Contains(runErr.Error(), tc.wantErr) {
				t.Fatalf("expected %q in the error, got %v", tc.wantErr, runErr)
			}

			assertRequests(t, api, tc.wantRun...)

			step.Cleanup(state)

			assertRequests(t, api,
				reqGetServer,
				reqDeleteServer,
				reqGetBlockVolume(fakeRootVolumeID),
				reqDeleteBlockVolume(fakeRootVolumeID),
			)

			if _, ok := state.GetOk(scaleway.StateKeyServerID); ok {
				t.Errorf("expected %s to be removed from the state", scaleway.StateKeyServerID)
			}
		})
	}
}

func TestStepCreateServerCleanupTerminatesRunningServer(t *testing.T) {
	t.Parallel()

	api := &fakeScalewayAPI{
		serverVolumes: map[string]*instance.VolumeServer{
			"0": {ID: fakeLocalVolumeID, VolumeType: instance.VolumeServerVolumeTypeLSSD},
			"1": {ID: fakeInstanceVolumeID, VolumeType: instance.VolumeServerVolumeTypeBSSD},
			"2": {ID: createdVolumeID(1), VolumeType: instance.VolumeServerVolumeTypeSbsVolume},
		},
		acceptedPoweron:   1,
		stateAfterPoweron: instance.ServerStateRunning,
		refuseDelete:      true,
	}
	state := newFakeState(t, api, &scaleway.Config{
		Zone:                  scw.ZoneFrPar1.String(),
		CommercialType:        "GP1-M",
		Image:                 "ubuntu_focal",
		ServerName:            "packer-test",
		BootType:              "local",
		RemoveVolume:          true,
		ServerCreationTimeout: time.Minute,
		BlockVolumes:          []scaleway.ConfigBlockVolume{{Name: "data", SizeInGB: 10}},
	})

	createdVolume := createdVolumeID(1)
	state.Put(scaleway.StateKeyCreatedVolumes, []*instance.VolumeServerTemplate{
		{ID: &createdVolume, VolumeType: instance.VolumeVolumeTypeSbsVolume},
	})

	step := scaleway.NewStepCreateServer()

	if action := step.Run(t.Context(), state); action != multistep.ActionContinue {
		t.Fatalf("expected continue, got %v: %v", action, state.Get("error"))
	}

	assertRequests(t, api, reqCreateServer, reqPoweron, reqGetServer)

	step.Cleanup(state)

	// Terminate deletes the local volume, and stepCreateVolume.Cleanup
	// removes the block volume it created.
	assertRequests(t, api,
		reqGetServer,
		reqDeleteServer,
		reqTerminate,
		reqGetInstanceVolume(fakeInstanceVolumeID),
		reqDeleteInstanceVolume(fakeInstanceVolumeID),
	)

	if _, ok := state.GetOk(scaleway.StateKeyServerID); ok {
		t.Errorf("expected %s to be removed from the state", scaleway.StateKeyServerID)
	}
}

func TestStepCreateServerCleanupKeepsServerIDWhenNotDestroyed(t *testing.T) {
	t.Parallel()

	api := &fakeScalewayAPI{
		serverVolumes: map[string]*instance.VolumeServer{
			"0": {ID: fakeRootVolumeID, VolumeType: instance.VolumeServerVolumeTypeSbsVolume},
		},
		acceptedPoweron:   1,
		stateAfterPoweron: instance.ServerStateRunning,
		refuseDelete:      true,
		refuseTerminate:   true,
	}
	state := newFakeState(t, api, &scaleway.Config{
		Zone:                  scw.ZoneFrPar1.String(),
		CommercialType:        "GP1-M",
		Image:                 "ubuntu_focal",
		ServerName:            "packer-test",
		BootType:              "local",
		RemoveVolume:          true,
		ServerCreationTimeout: time.Minute,
	})

	step := scaleway.NewStepCreateServer()

	if action := step.Run(t.Context(), state); action != multistep.ActionContinue {
		t.Fatalf("expected continue, got %v: %v", action, state.Get("error"))
	}

	assertRequests(t, api, reqCreateServer, reqPoweron, reqGetServer)

	step.Cleanup(state)

	// The root volume is still attached to the server, so it is kept.
	assertRequests(t, api, reqGetServer, reqDeleteServer, reqTerminate)

	// stepCreateVolume.Cleanup relies on it to keep the volumes.
	if _, ok := state.GetOk(scaleway.StateKeyServerID); !ok {
		t.Errorf("expected %s to stay in the state", scaleway.StateKeyServerID)
	}
}
