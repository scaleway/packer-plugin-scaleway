package scaleway

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *stepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if config.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")

		privateKeyBytes, err := config.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			return putErrorAndHalt(state, nil, err)
		}

		config.Comm.SSHPrivateKey = privateKeyBytes
		rawPrivateKey, err := ssh.ParseRawPrivateKey(privateKeyBytes)
		if err != nil {
			return putErrorAndHalt(state, ui, fmt.Errorf("error parsing SSH private key: %w", err))
		}

		signer, err := ssh.NewSignerFromKey(rawPrivateKey)
		if err != nil {
			return putErrorAndHalt(state, ui, fmt.Errorf("error creating SSH signer from private key: %w", err))
		}

		config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(signer.PublicKey())

		if isWindowsCommercialType(config.CommercialType) {
			if err := s.createWindowsIAMSSHKey(state); err != nil {
				return putErrorAndHalt(state, ui, err)
			}
		}

		return multistep.ActionContinue
	}

	ui.Say("Creating temporary ssh key for server...")

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error creating temporary SSH key: %w", err))
	}

	// ASN.1 DER encoded form
	privateDER := x509.MarshalPKCS1PrivateKey(priv)
	privateBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateDER,
	}

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("error creating temporary SSH key: %w", err))
	}

	log.Printf("temporary ssh key created")

	// Remember some state for the future
	config.Comm.SSHPrivateKey = pem.EncodeToMemory(&privateBlock)
	config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	if isWindowsCommercialType(config.CommercialType) {
		if err := s.createWindowsIAMSSHKey(state); err != nil {
			return putErrorAndHalt(state, ui, err)
		}
	}

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Message("Saving key for debug purposes: " + s.DebugKeyPath)

		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			return putErrorAndHalt(state, nil, fmt.Errorf("error saving debug key: %w", err))
		}

		defer f.Close() //nolint

		// Write the key out
		if _, err := f.Write(pem.EncodeToMemory(&privateBlock)); err != nil {
			return putErrorAndHalt(state, nil, fmt.Errorf("error saving debug key: %w", err))
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0o600); err != nil {
				return putErrorAndHalt(state, nil, fmt.Errorf("error setting permissions of debug key: %w", err))
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	sshKeyID, ok := state.GetOk("windows_iam_ssh_key_id")
	if !ok {
		return
	}

	client := state.Get("client").(*scw.Client)
	iamAPI := iam.NewAPI(client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting temporary Windows IAM SSH key...")

	if err := iamAPI.DeleteSSHKey(&iam.DeleteSSHKeyRequest{
		SSHKeyID: sshKeyID.(string),
	}); err != nil {
		ui.Error(fmt.Sprintf("Error deleting Windows IAM SSH key: %s", err))
	}
}

func (s *stepCreateSSHKey) createWindowsIAMSSHKey(state multistep.StateBag) error {
	client := state.Get("client").(*scw.Client)
	config := state.Get("config").(*Config)
	iamAPI := iam.NewAPI(client)

	sshKey, err := iamAPI.CreateSSHKey(&iam.CreateSSHKeyRequest{
		Name:      config.ServerName + "-windows-ssh",
		PublicKey: string(config.Comm.SSHPublicKey),
		ProjectID: config.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("error creating Windows IAM SSH key: %w", err)
	}

	state.Put("windows_iam_ssh_key_id", sshKey.ID)

	return nil
}
