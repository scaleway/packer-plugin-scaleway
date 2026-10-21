package scaleway_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	fakeServerID         = "11111111-1111-1111-1111-111111111111"
	fakeRootVolumeID     = "22222222-2222-2222-2222-222222222222"
	fakeLocalVolumeID    = "55555555-5555-5555-5555-555555555555"
	fakeInstanceVolumeID = "66666666-6666-6666-6666-666666666666"

	fakeInstancePath = "/instance/v1/zones/fr-par-1"
	fakeBlockPath    = "/block/v1alpha1/zones/fr-par-1"
	fakeServerPath   = fakeInstancePath + "/servers/" + fakeServerID

	reqCreateServer = http.MethodPost + " " + fakeInstancePath + "/servers"
	reqGetServer    = http.MethodGet + " " + fakeServerPath
	reqDeleteServer = http.MethodDelete + " " + fakeServerPath
	reqPoweron      = http.MethodPost + " " + fakeServerPath + "/action poweron"
	reqTerminate    = http.MethodPost + " " + fakeServerPath + "/action terminate"
	reqCreateVolume = http.MethodPost + " " + fakeBlockPath + "/volumes"
)

// createdVolumeID is the ID of the nth block volume created through the fake
// API, from 1.
func createdVolumeID(n int) string {
	return fmt.Sprintf("44444444-4444-4444-4444-%012d", n)
}

func reqGetBlockVolume(id string) string {
	return http.MethodGet + " " + fakeBlockPath + "/volumes/" + id
}

func reqDeleteBlockVolume(id string) string {
	return http.MethodDelete + " " + fakeBlockPath + "/volumes/" + id
}

func reqGetInstanceVolume(id string) string {
	return http.MethodGet + " " + fakeInstancePath + "/volumes/" + id
}

func reqDeleteInstanceVolume(id string) string {
	return http.MethodDelete + " " + fakeInstancePath + "/volumes/" + id
}

// fakeScalewayAPI fakes the instance and block APIs for a single server, and
// records the requests it gets. The fields before mu set the scenario.
type fakeScalewayAPI struct {
	// serverVolumes are the volumes of the server.
	serverVolumes map[string]*instance.VolumeServer
	// acceptedPoweron is the number of poweron calls accepted. Later ones are
	// rejected as out of stock.
	acceptedPoweron int
	// stateAfterPoweron is the server state after an accepted poweron:
	// running, or stopped for a start that failed asynchronously.
	stateAfterPoweron instance.ServerState
	// refuseDelete makes DeleteServer fail, as for a running server.
	refuseDelete bool
	// refuseTerminate makes the terminate action fail.
	refuseTerminate bool
	// failCreateVolume is the CreateVolume call, from 1, that fails on quota.
	// 0 accepts them all.
	failCreateVolume int

	t            *testing.T
	mu           sync.Mutex
	serverState  instance.ServerState
	poweron      int
	createVolume int
	requests     []string
}

func newFakeState(t *testing.T, api *fakeScalewayAPI, config *scaleway.Config) *multistep.BasicStateBag {
	t.Helper()

	api.t = t
	api.serverState = instance.ServerStateStopped

	ts := httptest.NewServer(api)
	t.Cleanup(ts.Close)

	client, err := scw.NewClient(scw.WithDefaultZone(scaleway.DefaultZone), scw.WithAPIURL(ts.URL))
	if err != nil {
		t.Fatalf("client: %s", err)
	}

	state := new(multistep.BasicStateBag)
	state.Put("client", client)
	state.Put("ui", &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer), ErrorWriter: new(bytes.Buffer)})
	state.Put("config", config)

	return state
}

// assertRequests checks the requests received since the last call, in order.
func assertRequests(t *testing.T, api *fakeScalewayAPI, want ...string) {
	t.Helper()

	api.mu.Lock()
	defer api.mu.Unlock()

	if !slices.Equal(api.requests, want) {
		t.Errorf("requests:\n  got:  %s\n  want: %s", strings.Join(api.requests, "\n        "), strings.Join(want, "\n        "))
	}

	api.requests = nil
}

func (f *fakeScalewayAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	req := r.Method + " " + r.URL.Path

	if strings.HasSuffix(r.URL.Path, "/action") {
		var action struct {
			Action instance.ServerAction `json:"action"`
		}

		if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
			f.t.Errorf("decoding action: %s", err)
		}

		req += " " + action.Action.String()
	}

	f.requests = append(f.requests, req)

	w.Header().Set("Content-Type", "application/json")

	status, body := f.reply(req, r)

	w.WriteHeader(status)

	if body == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		f.t.Errorf("encoding reply: %s", err)
	}
}

func (f *fakeScalewayAPI) reply(req string, r *http.Request) (int, any) {
	server := &instance.Server{
		ID:             fakeServerID,
		CommercialType: "GP1-M",
		State:          f.serverState,
		Zone:           scw.ZoneFrPar1,
		Volumes:        f.serverVolumes,
	}
	task := instance.ServerActionResponse{Task: &instance.Task{ID: "00000000-0000-0000-0000-000000000000"}}

	switch {
	case req == reqCreateServer:
		return http.StatusCreated, instance.CreateServerResponse{Server: server}
	case req == reqGetServer:
		return http.StatusOK, instance.GetServerResponse{Server: server}
	case req == reqPoweron:
		f.poweron++
		if f.poweron > f.acceptedPoweron {
			return http.StatusPreconditionFailed, map[string]string{"type": "out_of_stock", "message": "resource is out of stock", "resource": "GP1-M"}
		}

		f.serverState = f.stateAfterPoweron

		return http.StatusAccepted, task
	case req == reqTerminate && f.refuseTerminate:
		return http.StatusBadRequest, map[string]string{"type": "invalid_request_error", "message": "server is locked"}
	case req == reqTerminate:
		return http.StatusAccepted, task
	case req == reqDeleteServer && f.refuseDelete:
		return http.StatusBadRequest, map[string]string{"type": "invalid_request_error", "message": "server should be stopped"}
	case req == reqDeleteServer:
		return http.StatusNoContent, nil
	case req == reqCreateVolume:
		f.createVolume++
		if f.createVolume == f.failCreateVolume {
			return http.StatusForbidden, map[string]any{"type": "quotas_exceeded", "message": "quota exceeded", "details": []any{}}
		}

		return http.StatusOK, block.Volume{ID: createdVolumeID(f.createVolume)}
	case strings.HasPrefix(r.URL.Path, fakeBlockPath+"/volumes/"):
		if r.Method == http.MethodDelete {
			return http.StatusNoContent, nil
		}

		return http.StatusOK, block.Volume{ID: strings.TrimPrefix(r.URL.Path, fakeBlockPath+"/volumes/"), Status: block.VolumeStatusAvailable}
	case strings.HasPrefix(r.URL.Path, fakeInstancePath+"/volumes/"):
		if r.Method == http.MethodDelete {
			return http.StatusNoContent, nil
		}

		id := strings.TrimPrefix(r.URL.Path, fakeInstancePath+"/volumes/")

		return http.StatusOK, instance.GetVolumeResponse{Volume: &instance.Volume{ID: id, State: instance.VolumeStateAvailable}}
	default:
		// Cleanup treats 404 as already deleted, so fail the test instead of
		// letting it pass silently.
		f.t.Errorf("unexpected request %s", req)

		return http.StatusNotFound, map[string]string{"type": "not_found", "message": "unexpected request"}
	}
}
