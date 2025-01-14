package tester

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/vcr"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/require"
)

type PackerCtxKey struct{}

type PackerCtx struct {
	ScwClient *scw.Client
	ProjectID string
}

func getActiveProfile() *scw.Profile {
	cfg, err := scw.LoadConfig()
	if err != nil {
		return &scw.Profile{}
	}
	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return &scw.Profile{}
	}

	return activeProfile
}

func NewTestContext(ctx context.Context, httpClient *http.Client) (context.Context, error) {
	activeProfile := getActiveProfile()
	profile := scw.MergeProfiles(activeProfile, scw.LoadEnvProfile())
	client, err := scw.NewClient(scw.WithProfile(profile), scw.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("error creating scw client: %w", err)
	}
	projectID, exists := client.GetDefaultProjectID()
	if !exists {
		if vcr.UpdateCassettes == false {
			projectID = "11111111-1111-1111-1111-111111111111"
		} else {
			return nil, errors.New("error getting default project ID")
		}
	}

	return context.WithValue(ctx, PackerCtxKey{}, &PackerCtx{
		ScwClient: client,
		ProjectID: projectID,
	}), nil
}

func ExtractCtx(ctx context.Context) *PackerCtx {
	return ctx.Value(PackerCtxKey{}).(*PackerCtx)
}

type TestConfig struct {
	Config string
	Checks []PackerCheck
}

func Test(t *testing.T, config *TestConfig) {
	httpClient, cleanup, err := vcr.GetHTTPRecorder(vcr.GetTestFilePath(t, "."), vcr.UpdateCassettes)
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()
	ctx, err = NewTestContext(ctx, httpClient)
	require.NoError(t, err)

	// Create TMP Dir
	tmpDir, err := os.MkdirTemp(os.TempDir(), "packer_e2e_test")
	require.NoError(t, err)
	t.Logf("Created tmp dir: %s", tmpDir)

	err = packerExec(tmpDir, config.Config)
	require.NoError(t, err, "error executing packer command: %s", err)

	for i, check := range config.Checks {
		t.Logf("Running check %d/%d", i+1, len(config.Checks))
		err := check.Check(ctx)
		if err != nil {
			t.Fail()
			t.Errorf("Packer check %d failed: %s", i+1, err.Error())
		}
	}

	t.Logf("Deleting tmp dir: %s", tmpDir)
	require.NoError(t, os.RemoveAll(tmpDir), "failed to remote tmp dir %s", tmpDir)
}
