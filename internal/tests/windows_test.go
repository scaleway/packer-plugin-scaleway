package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// This ID corresponds to the SSH key named "opensource@scaleway.com" of the "packer-e2e" project on the "terraform-provider-scaleway" organization.
// It should be removed when we're able to implement pre-tasks and share the same client/context between the pre-tasks and the actual test tasks, without one returning early.
// Then we'll be able to create and upload the key at each run of the test.
const sshPublicKeyID = "c31f516d-ed85-4aca-bfd6-ca352060f536"

func TestWindows(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-windows"
	serverName := "packer-tmp-server-windows"

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "POP2-2C-8G-WIN"
			  image = "windows_server_2022"
			  zone = "%s"
			  image_name = "%s"
			  ssh_username = "root"
			  tags = ["%s", "%s", "%s"]
			  keep_server = true
			  remove_volume = false

			  server_name = "%s"
              server_tags = [ "with-ssh" ]
			  admin_password_encryption_ssh_key_id = %q
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, tagDevtools, tagProvider, tagPacker, serverName, sshPublicKeyID),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				Tags(e2eTagsDevtools).
				RootVolumeSnapshot(checks.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix).
					Tags(e2eTagsDevtools),
				),
			checks.Server(zone, serverName).
				Tags([]string{"with-ssh"}).
				AdminPasswordEncryptionSSHKeyID(sshPublicKeyID),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix),
			cleanup.Server(zone, serverName),
		},
	})
}
