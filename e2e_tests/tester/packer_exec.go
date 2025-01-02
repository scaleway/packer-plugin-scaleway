package tester

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const PackerFileHeader = `
packer {
  required_plugins {
  }
}
`

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build image with packer: %w", err)
	}

	return nil
}
