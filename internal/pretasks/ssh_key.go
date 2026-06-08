package pretasks

import (
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type SSHKeyTask struct {
	publicKey  string
	sshKeyName string
	sshKeyID   string
}

func SSHKey(publicKey string, sshKeyName string) *SSHKeyTask {
	return &SSHKeyTask{
		publicKey:  publicKey,
		sshKeyName: sshKeyName,
	}
}

func (s *SSHKeyTask) Create(t *testing.T) error {
	ctx := tester.CreateClientAndContext(t)
	testCtx := tester.ExtractCtx(ctx)
	api := iam.NewAPI(testCtx.ScwClient)

	t.Log("Running pre-task: Create SSH key")

	sshKey, err := api.CreateSSHKey(&iam.CreateSSHKeyRequest{
		Name:      s.sshKeyName,
		PublicKey: s.publicKey,
		ProjectID: testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	s.sshKeyID = sshKey.ID

	return nil
}

func (s *SSHKeyTask) GetID() string {
	return s.sshKeyID
}
