package scaleway_test

import (
	"strconv"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
)

func TestBuilderPrepare_SnapshotName(t *testing.T) {
	var b scaleway.Builder

	config := map[string]any{
		"project_id":      "00000000-1111-2222-3333-444444444444",
		"access_key":      "SCWABCXXXXXXXXXXXXXX",
		"secret_key":      "00000000-1111-2222-3333-444444444444",
		"zone":            "fr-par-3",
		"commercial_type": "PRO2-S",
		"ssh_username":    "root",
		"image":           "image-uuid",
		"root_volume": map[string]any{
			"snapshot_name": "default",
		},
	}

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}

	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.Config.RootVolume.SnapshotName == "" {
		t.Errorf("invalid: %s", b.Config.RootVolume.SnapshotName)
	}

	config["root_volume"] = map[string]any{"snapshot_name": "foobarbaz"}
	b = scaleway.Builder{}

	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}

	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["root_volume"] = map[string]any{"snapshot_name": "{{timestamp}}"}
	b = scaleway.Builder{}

	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}

	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	_, err = strconv.ParseInt(b.Config.RootVolume.SnapshotName, 0, 0)
	if err != nil {
		t.Fatalf("failed to parse int in template: %s", err)
	}
}
