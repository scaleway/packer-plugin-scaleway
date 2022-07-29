package e2e

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/e2e"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) *multistep.BasicStateBag {
	c, cleanup, err := e2e.NewClient(t)
	defer func() {
		cleanup()
	}()

	require.NoError(t, err)

	state := multistep.BasicStateBag{}
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("client", c)

	return &state
}

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/scaleway/builder_acc_test.go  -timeout=120m
func TestAccScalewayBuilder(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-scaleway-builder-basic",
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) bool {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			acctest.TestEnvVar))
		return true
	}
	return false
}

const testBuilderAccBasic = `
source "scaleway" "basic" {
  commercial_type      = "DEV1-S"
  image                = "ubuntu_focal"
  image_name 		   = "boo"
  ssh_username         = "root"
  zone                 = "fr-par-1"
}

build {
  sources = ["source.scaleway.basic"]
}
`
