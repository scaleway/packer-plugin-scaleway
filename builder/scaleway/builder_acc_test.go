package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

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
