package scaleway_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/scaleway/builder_acc_test.go  -timeout=120m
func TestAccScalewayBuilder(t *testing.T) {
	if skip := testAccPreCheck(t); skip {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-scaleway-builder-basic",
		Template: testBuilderAccBasic,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func testAccPreCheck(t *testing.T) bool {
	t.Helper()
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			acctest.TestEnvVar)
		return true
	}
	return false
}

const testBuilderAccBasic = `
source "scaleway" "basic" {
  commercial_type      = "PRO2-XXS"
  image                = "ubuntu_focal"
  image_name 		   = "Acceptance test"
  ssh_username         = "root"
  zone                 = "fr-par-1"
  remove_volume        = true
  tags                 = ["devtools", "provider", "packer"]
}

build {
  sources = ["source.scaleway.basic"]
}
`
