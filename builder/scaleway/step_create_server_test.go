package scaleway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	fakeServerID     = "11111111-1111-1111-1111-111111111111"
	fakeRootVolumeID = "22222222-2222-2222-2222-222222222222"
	fakeBlockVolume  = "33333333-3333-3333-3333-333333333333"
	fakeBlockPath    = "/block/v1alpha1/zones/fr-par-1/volumes/"
	fakeZonePath     = "/instance/v1/zones/fr-par-1"
)

// fakeInstanceAPI mimics a server whose first poweron is accepted then fails
// asynchronously, while later poweron calls are rejected as out of stock.
type fakeInstanceAPI struct {
	mu         sync.Mutex
	poweron    int
	requests   []string
	unexpected []string
}

func (f *fakeInstanceAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requests = append(f.requests, r.Method+" "+r.URL.Path)

	server := &instance.Server{
		ID:             fakeServerID,
		CommercialType: "GP1-M",
		State:          instance.ServerStateStopped,
		Zone:           scw.ZoneFrPar1,
		Volumes: map[string]*instance.VolumeServer{
			"0": {ID: fakeRootVolumeID, VolumeType: instance.VolumeServerVolumeTypeSbsVolume},
		},
	}

	w.Header().Set("Content-Type", "application/json")

	var body any

	switch r.Method + " " + r.URL.Path {
	case "POST " + fakeZonePath + "/servers":
		w.WriteHeader(http.StatusCreated)

		body = instance.CreateServerResponse{Server: server}
	case "POST " + fakeZonePath + "/servers/" + fakeServerID + "/action":
		f.poweron++
		if f.poweron > 1 {
			w.WriteHeader(http.StatusPreconditionFailed)

			body = map[string]string{"type": "out_of_stock", "message": "resource is out of stock", "resource": "GP1-M"}

			break
		}

		w.WriteHeader(http.StatusAccepted)

		body = instance.ServerActionResponse{Task: &instance.Task{ID: "00000000-0000-0000-0000-000000000000"}}
	case "GET " + fakeZonePath + "/servers/" + fakeServerID:
		body = instance.GetServerResponse{Server: server}
	case "DELETE " + fakeZonePath + "/servers/" + fakeServerID:
		w.WriteHeader(http.StatusNoContent)

		return
	case "GET " + fakeBlockPath + fakeRootVolumeID, "GET " + fakeBlockPath + fakeBlockVolume:
		body = block.Volume{ID: strings.TrimPrefix(r.URL.Path, fakeBlockPath), Status: block.VolumeStatusAvailable}
	case "DELETE " + fakeBlockPath + fakeRootVolumeID, "DELETE " + fakeBlockPath + fakeBlockVolume:
		w.WriteHeader(http.StatusNoContent)

		return
	default:
		// Cleanup treats 404 as already deleted, so record the request for
		// the test to fail on instead of letting it pass silently.
		f.unexpected = append(f.unexpected, r.Method+" "+r.URL.Path)

		w.WriteHeader(http.StatusNotFound)

		body = map[string]string{"type": "not_found", "message": "unexpected request"}
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newFakeState(t *testing.T, config *scaleway.Config) (*fakeInstanceAPI, *multistep.BasicStateBag) {
	t.Helper()

	api := &fakeInstanceAPI{}
	ts := httptest.NewServer(api)
	t.Cleanup(ts.Close)

	client, err := scw.NewClient(scw.WithDefaultZone(scaleway.DefaultZone), scw.WithAPIURL(ts.URL))
	if err != nil {
		t.Fatalf("client: %s", err)
	}

	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer), ErrorWriter: new(bytes.Buffer)}

	state := new(multistep.BasicStateBag)
	state.Put("client", client)
	state.Put("ui", ui)
	state.Put("config", config)

	return api, state
}

func TestStepCreateServerOutOfStockCleansUp(t *testing.T) {
	t.Parallel()

	api, state := newFakeState(t, &scaleway.Config{
		Zone:                  scw.ZoneFrPar1.String(),
		CommercialType:        "GP1-M",
		Image:                 "ubuntu_focal",
		ServerName:            "packer-test",
		BootType:              "local",
		RemoveVolume:          true,
		ServerCreationTimeout: time.Minute,
	})

	step := scaleway.NewStepCreateServer()

	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("expected halt, got %v", action)
	}

	runErr, _ := state.Get("error").(error)
	if runErr == nil || !strings.Contains(runErr.Error(), "resource GP1-M is out of stock") {
		t.Fatalf("expected the out of stock reason in the error, got %v", runErr)
	}

	step.Cleanup(state)

	got := strings.Join(api.requests, "\n")
	for _, want := range []string{
		"DELETE " + fakeZonePath + "/servers/" + fakeServerID,
		"DELETE /block/v1alpha1/zones/fr-par-1/volumes/" + fakeRootVolumeID,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected request %q, got:\n%s", want, got)
		}
	}

	if strings.Index(got, "DELETE "+fakeZonePath) > strings.Index(got, "DELETE /block") {
		t.Errorf("root volume must be deleted after the server, got:\n%s", got)
	}

	if len(api.unexpected) > 0 {
		t.Errorf("unexpected requests:\n%s", strings.Join(api.unexpected, "\n"))
	}
}

func TestStepCreateVolumeCleanupKeepServer(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		serverExists bool
		wantDelete   bool
	}{
		"no server was created": {serverExists: false, wantDelete: true},
		"kept server":           {serverExists: true, wantDelete: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			api, state := newFakeState(t, &scaleway.Config{
				Zone:         scw.ZoneFrPar1.String(),
				RemoveVolume: true,
				KeepServer:   true,
			})

			volumeID := fakeBlockVolume
			state.Put(scaleway.StateKeyCreatedVolumes, []*instance.VolumeServerTemplate{{ID: &volumeID}})

			if tc.serverExists {
				state.Put("server_id", fakeServerID)
			}

			scaleway.NewStepCreateVolume().Cleanup(state)

			deleted := strings.Contains(strings.Join(api.requests, "\n"), "DELETE "+fakeBlockPath+fakeBlockVolume)
			if deleted != tc.wantDelete {
				t.Errorf("block volume deleted: got %t, want %t, requests:\n%s", deleted, tc.wantDelete, strings.Join(api.requests, "\n"))
			}

			if len(api.unexpected) > 0 {
				t.Errorf("unexpected requests:\n%s", strings.Join(api.unexpected, "\n"))
			}
		})
	}
}
