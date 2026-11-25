package tests_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain builds the plugin from the current code and installs it in a temporary
// PACKER_PLUGIN_PATH, so that tests never run against another installed version
func TestMain(m *testing.M) {
	if _, err := exec.LookPath("packer"); err != nil {
		log.Println("packer binary not found in PATH, skipping e2e tests")

		return
	}

	pluginDir, err := os.MkdirTemp("", "packer-plugin-scaleway-e2e")
	if err != nil {
		log.Fatalf("failed to create plugin dir: %s", err)
	}

	code := 1

	if err := installPlugin(pluginDir); err != nil {
		log.Println(err)
	} else {
		code = m.Run()
	}

	_ = os.RemoveAll(pluginDir)

	os.Exit(code)
}

func installPlugin(pluginDir string) error {
	ctx := context.Background()
	binary := filepath.Join(pluginDir, "packer-plugin-scaleway")

	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "../..") //nolint:gosec
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build plugin: %w: %s", err, output)
	}

	pluginPath := filepath.Join(pluginDir, "plugins")
	if err := os.Setenv("PACKER_PLUGIN_PATH", pluginPath); err != nil {
		return fmt.Errorf("failed to set PACKER_PLUGIN_PATH: %w", err)
	}

	install := exec.CommandContext(ctx, "packer", "plugins", "install", "--path", binary, "github.com/scaleway/scaleway") //nolint:gosec
	if output, err := install.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install plugin: %w: %s", err, output)
	}

	return nil
}
