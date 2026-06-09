package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/pretasks"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/require"
)

func TestWindows(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-windows"
	serverName := "packer-tmp-server-windows"

	ctx, vcrCleanupFunc := tester.CreateRecordedClientAndContext(t)

	defer vcrCleanupFunc()

	t.Log("Running pre-task: Create SSH key")

	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFNaFderD6JUbMr6LoL7SdTaQ31gLcXwKv07Zyw0t4pq6Y8CGaeEvevS54TBR2iNJHa3hlIIUmA2qvH7Oh4v1QmMG2djWi2cD1lDEl8/8PYakaEBGh6snp3TMyhoqHOZqqKwDhPW0gJbe2vXfAgWSEzI8h1fs1D7iEkC1L/11hZjkqbUX/KduWFLyIRWdSuI3SWk4CXKRXwIkeYeSYb8AiIGY21u2z8H2J7YmhRzE85Kj/Fk4tST5gLW/IfLD4TMJjC/cZiJevETjs+XVmzTMIyU2sTQKufSQTj2qZ7RfgGwTHDoOeFvylgAdMGLZ/Un+gzeEPj9xUSPvvnbA9UPIKV4AffgtT1y5gcSWuHaqRxpUTY204mh6kq0EdVN2UsiJTgX+xnJgnOrKg6G3dkM8LSi2QtbjYbRXcuDJ9YUbUFK8M5Vo7LhMsMFb1hPtY68kbDUqD01RuMD5KhGIngCRRBZJriRQclUCJS4D3jr/Frw9ruNGh+NTIvIwdv0Y2brU= opensource@scaleway.com"
	sshKeyName := "packer-tests-windows"
	sshKeyPreTask := pretasks.SSHKey(sshPublicKey, sshKeyName)
	err := sshKeyPreTask.Create(ctx)
	require.NoError(t, err)

	tester.Test(ctx, t, &tester.TestConfig{
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
			`, zone, imageName, tagDevtools, tagProvider, tagPacker, serverName, sshKeyPreTask.GetID()),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				Tags(e2eTagsDevtools).
				RootVolumeSnapshot(checks.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix).
					Tags(e2eTagsDevtools),
				),
			checks.Server(zone, serverName).
				Tags([]string{"with-ssh"}).
				AdminPasswordEncryptionSSHKeyID(sshKeyPreTask.GetID()),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, apiGeneratedSnapshotNamePrefix),
			cleanup.Server(zone, serverName),
			cleanup.SSHKey(sshKeyPreTask.GetID()),
		},
	})
}
