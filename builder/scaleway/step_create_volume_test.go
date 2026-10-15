package scaleway_test

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestStepCreateVolumeCleanupKeepServer(t *testing.T) {
	t.Parallel()

	createdVolume := createdVolumeID(1)

	for name, tc := range map[string]struct {
		keepServer   bool
		serverExists bool
		want         []string
	}{
		"no server was created": {
			keepServer: true,
			want:       []string{reqGetBlockVolume(createdVolume), reqDeleteBlockVolume(createdVolume)},
		},
		"kept server": {
			keepServer:   true,
			serverExists: true,
		},
		"server not destroyed": {
			serverExists: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			api := &fakeScalewayAPI{}
			state := newFakeState(t, api, &scaleway.Config{
				Zone:         scw.ZoneFrPar1.String(),
				RemoveVolume: true,
				KeepServer:   tc.keepServer,
			})

			state.Put(scaleway.StateKeyCreatedVolumes, []*instance.VolumeServerTemplate{{ID: &createdVolume}})

			if tc.serverExists {
				state.Put(scaleway.StateKeyServerID, fakeServerID)
			}

			scaleway.NewStepCreateVolume().Cleanup(state)

			assertRequests(t, api, tc.want...)
		})
	}
}

func TestStepCreateVolumeCleansUpPartialCreation(t *testing.T) {
	t.Parallel()

	api := &fakeScalewayAPI{failCreateVolume: 2}
	state := newFakeState(t, api, &scaleway.Config{
		Zone:         scw.ZoneFrPar1.String(),
		RemoveVolume: true,
		BlockVolumes: []scaleway.ConfigBlockVolume{
			{Name: "data", SizeInGB: 10},
			{Name: "logs", SizeInGB: 10},
		},
	})

	step := scaleway.NewStepCreateVolume()

	if action := step.Run(t.Context(), state); action != multistep.ActionHalt {
		t.Fatalf("expected halt, got %v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Error("expected an error in the state")
	}

	assertRequests(t, api, reqCreateVolume, reqCreateVolume)

	created, _ := state.Get(scaleway.StateKeyCreatedVolumes).([]*instance.VolumeServerTemplate)
	if len(created) != 1 || *created[0].ID != createdVolumeID(1) {
		t.Fatalf("expected the first volume in the state, got %v", created)
	}

	step.Cleanup(state)

	assertRequests(t, api, reqGetBlockVolume(createdVolumeID(1)), reqDeleteBlockVolume(createdVolumeID(1)))
}
