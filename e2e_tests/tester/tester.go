package tester

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/require"
)

const PackerCtxKey = "PACKER_CTX_KEY"

type PackerCtx struct {
	ScwClient *scw.Client
	ProjectID string
}

func NewContext(ctx context.Context) (context.Context, error) {
	cfg, err := scw.LoadConfig()
	if err != nil {
		return nil, err
	}
	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return nil, err
	}

	profile := scw.MergeProfiles(activeProfile, scw.LoadEnvProfile())
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		return nil, err
	}
	projectID, exists := client.GetDefaultProjectID()
	if !exists {
		return nil, errors.New("error getting default project ID")
	}

	return context.WithValue(ctx, PackerCtxKey, &PackerCtx{
		ScwClient: client,
		ProjectID: projectID,
	}), nil
}

func ExtractCtx(ctx context.Context) *PackerCtx {
	return ctx.Value(PackerCtxKey).(*PackerCtx)
}

type TestConfig struct {
	Config string
	Checks []PackerCheck
}

func Test(t *testing.T, config *TestConfig) {
	ctx := context.Background()
	ctx, err := NewContext(ctx)
	require.Nil(t, err)

	// Create TMP Dir
	tmpDir, err := os.MkdirTemp(os.TempDir(), "packer_e2e_test")
	require.Nil(t, err)
	t.Logf("Created tmp dir: %s", tmpDir)

	err = packerExec(tmpDir, config.Config)
	require.Nil(t, err, "error executing packer command")
	
	for i, check := range config.Checks {
		t.Logf("Running check %d/%d", i+1, len(config.Checks))
		err := check.Check(ctx)
		if err != nil {
			t.Fail()
			t.Errorf("Packer check %d failed: %s", i+1, err.Error())
		}
	}

	t.Logf("Deleting tmp dir: %s", tmpDir)
	require.Nil(t, os.RemoveAll(tmpDir), "failed to remote tmp dir %s", tmpDir)
}
