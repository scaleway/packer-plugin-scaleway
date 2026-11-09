package tester

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
		switch {
		case strings.HasPrefix(envVariable, scw.ScwDefaultProjectIDEnv):
			hasProject = true
		case strings.HasPrefix(envVariable, scw.ScwAccessKeyEnv):
			hasAccessKey = true
		case strings.HasPrefix(envVariable, scw.ScwSecretKeyEnv):
			hasSecretKey = true
		case strings.HasPrefix(envVariable, vcr.UpdateCassettesEnvVariable):
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

// packerExec runs `packer build` with the given args on packerConfig, using cassettePath as the plugin's cassette
// It returns the combined output of the command
func packerExec(ctx context.Context, folder, packerConfig string, args []string, cassettePath string, fakeEnv bool) (string, error) {
	// Create Packer file
	packerFile := filepath.Join(folder, "build_scaleway.pkr.hcl")
	packerFileContent := PackerFileHeader + packerConfig

	err := os.WriteFile(packerFile, []byte(packerFileContent), 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to create packer file: %w", err)
	}

	// Run Packer
	cmdArgs := append(append([]string{"build"}, args...), packerFile)

	cmd := exec.CommandContext(ctx, "packer", cmdArgs...) //nolint:gosec
	cmd.Env = os.Environ()

	if fakeEnv {
		cmd.Env = preparePackerEnv(cmd.Env)
	}

	cmd.Env = append(cmd.Env, vcr.CassettePathEnvVariable+"="+cassettePath)

	output := new(bytes.Buffer)
	cmd.Stdout = io.MultiWriter(os.Stdout, output)
	cmd.Stderr = io.MultiWriter(os.Stderr, output)

	err = cmd.Run()
	if err != nil {
		return output.String(), fmt.Errorf("failed to build image with packer: %w", err)
	}

	return output.String(), nil
}
