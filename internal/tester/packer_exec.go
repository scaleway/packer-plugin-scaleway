package tester

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/scaleway/packer-plugin-scaleway/internal/vcr"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const PackerFileHeader = `
packer {
  required_plugins {
  }
}
`

// preparePackerEnv will prepare an Environ to run packer in tests.
// Some scaleway config variable are required for the client to be created.
// Cassettes must also be configured to be used, disabling the recording is enough.
func preparePackerEnv(currentEnv []string) []string {
	hasProject := false
	hasAccessKey := false
	hasSecretKey := false
	hasCassettesConfigured := false

	env := make([]string, 0, len(currentEnv))
	for _, envVariable := range currentEnv {
		if strings.HasPrefix(envVariable, scw.ScwDefaultProjectIDEnv) {
			hasProject = true
		} else if strings.HasPrefix(envVariable, scw.ScwAccessKeyEnv) {
			hasAccessKey = true
		} else if strings.HasPrefix(envVariable, scw.ScwSecretKeyEnv) {
			hasSecretKey = true
		} else if strings.HasPrefix(envVariable, vcr.UpdateCassettesEnvVariable) {
			hasCassettesConfigured = true
		}

		env = append(env, envVariable)
	}
	if !hasProject {
		env = append(env, scw.ScwDefaultProjectIDEnv+"=11111111-1111-1111-1111-111111111111")
	}
	if !hasAccessKey {
		env = append(env, scw.ScwAccessKeyEnv+"=SCWXXXXXXXXXXXXXFAKE")
	}
	if !hasSecretKey {
		env = append(env, scw.ScwSecretKeyEnv+"=11111111-1111-1111-1111-111111111111")
	}
	if !hasCassettesConfigured {
		env = append(env, vcr.UpdateCassettesEnvVariable+"=false")
	}

	return env
}

func packerExec(folder, packerConfig string) error {
	// Create Packer file
	packerFile := filepath.Join(folder, "build_scaleway.pkr.hcl")
	packerFileContent := PackerFileHeader + packerConfig
	err := os.WriteFile(packerFile, []byte(packerFileContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create packer file: %w", err)
	}

	// Run Packer
	cmd := exec.Command("packer", "build", packerFile)
	cmd.Env = preparePackerEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build image with packer: %w", err)
	}

	return nil
}
