package scaleway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestStepCreateVolumeFromSnapshot(t *testing.T) {
	t.Parallel()

	const snapshotID = "33333333-3333-3333-3333-333333333333"

	var got block.CreateVolumeRequest

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/block/v1alpha1/zones/fr-par-1/volumes" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)

			return
		}

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decoding request: %s", err)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(block.Volume{ID: "44444444-4444-4444-4444-444444444444"}); err != nil {
			t.Errorf("encoding reply: %s", err)
		}
	}))
	t.Cleanup(ts.Close)

	client, err := scw.NewClient(scw.WithDefaultZone(scaleway.DefaultZone), scw.WithAPIURL(ts.URL))
	if err != nil {
		t.Fatalf("client: %s", err)
	}

	state := new(multistep.BasicStateBag)
	state.Put("client", client)
	state.Put("ui", &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer)})
	state.Put("config", &scaleway.Config{
		Zone:         scw.ZoneFrPar1.String(),
		BlockVolumes: []scaleway.ConfigBlockVolume{{Name: "data", SnapshotID: snapshotID}},
	})

	if action := scaleway.NewStepCreateVolume().Run(t.Context(), state); action != multistep.ActionContinue {
		t.Fatalf("expected continue, got %v: %v", action, state.Get("error"))
	}

	if got.FromSnapshot == nil || got.FromSnapshot.SnapshotID != snapshotID {
		t.Fatalf("expected the volume to be created from snapshot %s, got %+v", snapshotID, got.FromSnapshot)
	}

	if got.FromSnapshot.Size != nil {
		t.Errorf("expected no size, the snapshot size applies, got %v", *got.FromSnapshot.Size)
	}
}
