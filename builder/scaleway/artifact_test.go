package scaleway

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packersdk.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactId(t *testing.T) {
	generatedData := make(map[string]interface{})
	a := &Artifact{
		"packer-foobar-image",
		"cc586e45-5156-4f71-b223-cf406b10dd1d",
		[]ArtifactSnapshot{{
			"packer-foobar-snapshot",
			"cc586e45-5156-4f71-b223-cf406b10dd1c",
		}},
		"ams1",
		nil,
		generatedData}
	expected := "ams1:cc586e45-5156-4f71-b223-cf406b10dd1d"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	generatedData := make(map[string]interface{})
	a := &Artifact{
		"packer-foobar-image",
		"cc586e45-5156-4f71-b223-cf406b10dd1d",
		[]ArtifactSnapshot{
			{
				"cc586e45-5156-4f71-b223-cf406b10dd1c",
				"packer-foobar-snapshot",
			},
			{
				"cc586e45-5156-4f71-b223-cf406b10dd1e",
				"packer-foobar-snapshot2",
			},
		},
		"ams1",
		nil,
		generatedData}
	expected := "An image was created: 'packer-foobar-image' (ID: cc586e45-5156-4f71-b223-cf406b10dd1d) in zone 'ams1' based on snapshots [(packer-foobar-snapshot: cc586e45-5156-4f71-b223-cf406b10dd1c) (packer-foobar-snapshot2: cc586e45-5156-4f71-b223-cf406b10dd1e)]"

	if a.String() != expected {
		t.Fatalf("artifact string (%v) should match: %v", a.String(), expected)
	}
}

func TestArtifactState_StateData(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = artifact.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}
