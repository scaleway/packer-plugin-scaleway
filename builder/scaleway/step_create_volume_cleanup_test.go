package scaleway_test

import (
	"bytes"
	"encoding/json"
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
	cleanupInstanceVolumesPath = "/instance/v1/zones/fr-par-1/volumes/"
	cleanupBlockVolumesPath    = "/block/v1alpha1/zones/fr-par-1/volumes/"
)

// fakeVolumeAPI serves the instance and block volumes it holds, and answers
// 404 for any other volume.
type fakeVolumeAPI struct {
	mu              sync.Mutex
	instanceVolumes []string
	blockVolumes    []string
	failDelete      string // volume whose deletion gets a 500
	requests        []string
	deleted         []string
}

func (f *fakeVolumeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	var (
		volumes *[]string
		body    any
	)

	id := r.URL.Path

	switch {
	case strings.HasPrefix(r.URL.Path, cleanupInstanceVolumesPath):
		volumes = &f.instanceVolumes
		id = strings.TrimPrefix(id, cleanupInstanceVolumesPath)
		f.requests = append(f.requests, "instance "+id)
		body = instance.GetVolumeResponse{Volume: &instance.Volume{ID: id, State: instance.VolumeStateAvailable}}
	case strings.HasPrefix(r.URL.Path, cleanupBlockVolumesPath):
		volumes = &f.blockVolumes
		id = strings.TrimPrefix(id, cleanupBlockVolumesPath)
		f.requests = append(f.requests, "block "+id)
		body = block.Volume{ID: id, Status: block.VolumeStatusAvailable}
	}

	switch {
	case volumes == nil || !slices.Contains(*volumes, id):
		w.WriteHeader(http.StatusNotFound)

		body = map[string]string{"type": "not_found", "message": "resource is not found"}
	case r.Method == http.MethodDelete && id == f.failDelete:
		w.WriteHeader(http.StatusInternalServerError)

		body = map[string]string{"type": "internal_error", "message": "internal error"}
	case r.Method == http.MethodDelete:
		*volumes = slices.DeleteFunc(*volumes, func(v string) bool { return v == id })

		f.deleted = append(f.deleted, id)

		w.WriteHeader(http.StatusNoContent)

		return
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestStepCreateVolumeCleanupAfterImage(t *testing.T) {
	t.Parallel()

	const (
		goneVolume     = "11111111-1111-1111-1111-111111111111"
		instanceVolume = "22222222-2222-2222-2222-222222222222"
		blockVolume    = "33333333-3333-3333-3333-333333333333"
	)

	for name, tc := range map[string]struct {
		api         *fakeVolumeAPI
		wantDeleted []string
		wantErr     bool
	}{
		"volume already deleted with the server": {
			api: &fakeVolumeAPI{},
		},
		"instance and block volumes": {
			api:         &fakeVolumeAPI{instanceVolumes: []string{instanceVolume}, blockVolumes: []string{blockVolume}},
			wantDeleted: []string{instanceVolume, blockVolume},
		},
		"instance volume deletion fails": {
			api:         &fakeVolumeAPI{instanceVolumes: []string{instanceVolume}, blockVolumes: []string{blockVolume}, failDelete: instanceVolume},
			wantDeleted: []string{blockVolume},
			wantErr:     true,
		},
		"block volume deletion fails": {
			api:     &fakeVolumeAPI{blockVolumes: []string{blockVolume}, failDelete: blockVolume},
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(tc.api)
			t.Cleanup(ts.Close)

			client, err := scw.NewClient(scw.WithDefaultZone(scaleway.DefaultZone), scw.WithAPIURL(ts.URL))
			if err != nil {
				t.Fatalf("client: %s", err)
			}

			state := new(multistep.BasicStateBag)
			state.Put("client", client)
			state.Put("ui", &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer), ErrorWriter: new(bytes.Buffer)})
			state.Put("config", &scaleway.Config{Zone: scw.ZoneFrPar1.String(), RemoveVolume: true})
			state.Put("snapshots", []*instance.Snapshot{})
			state.Put("volumes", []*instance.VolumeServer{
				{ID: goneVolume, VolumeType: instance.VolumeServerVolumeTypeLSSD},
				{ID: instanceVolume, VolumeType: instance.VolumeServerVolumeTypeLSSD},
				{ID: blockVolume, VolumeType: instance.VolumeServerVolumeTypeSbsVolume},
			})

			scaleway.NewStepCreateVolume().Cleanup(state)

			if !slices.Equal(tc.api.deleted, tc.wantDeleted) {
				t.Errorf("deleted volumes: got %v, want %v", tc.api.deleted, tc.wantDeleted)
			}

			// Each volume goes to the API of its type only.
			for _, wrongAPI := range []string{"block " + goneVolume, "block " + instanceVolume, "instance " + blockVolume} {
				if slices.Contains(tc.api.requests, wrongAPI) {
					t.Errorf("unexpected request %q, got %v", wrongAPI, tc.api.requests)
				}
			}

			if _, gotErr := state.GetOk("error"); gotErr != tc.wantErr {
				t.Errorf("error in state: got %t, want %t (%v)", gotErr, tc.wantErr, state.Get("error"))
			}
		})
	}
}
