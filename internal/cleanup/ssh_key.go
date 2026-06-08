package cleanup

import (
	"context"
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*SSHKeyCleanup)(nil)

type SSHKeyCleanup struct {
	id string
}

func SSHKey(name string) *SSHKeyCleanup {
	return &SSHKeyCleanup{
		id: name,
	}
}

func (i *SSHKeyCleanup) Cleanup(ctx context.Context, t *testing.T) error {
	t.Helper()

	testCtx := tester.ExtractCtx(ctx)
	api := iam.NewAPI(testCtx.ScwClient)

	err := api.DeleteSSHKey(&iam.DeleteSSHKeyRequest{
		SSHKeyID: i.id,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}

	t.Logf("deleted SSH key %s\n", i.id)

	return nil
}
