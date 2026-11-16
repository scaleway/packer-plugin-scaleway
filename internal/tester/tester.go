package tester

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
		if !vcr.UpdateCassettes {
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

type TestStep struct {
	Config string
	Args []string
	ExpectError *regexp.Regexp
	Checks      []PackerCheck
}

type TestConfig struct {
	Config string
	Checks []PackerCheck

	Steps   []TestStep
	Cleanup []PackerCleanup
}

func Test(t *testing.T, config *TestConfig) {
	cassettesDir, err := vcr.GetTestDirPath(t, ".")
	require.NoError(t, err)

	if vcr.UpdateCassettes {
		require.NoError(t, os.MkdirAll(cassettesDir, 0o750))
	}

	httpClient, vcrCleanupFunc, err := vcr.GetHTTPRecorder(filepath.Join(cassettesDir, "tester.cassette"), vcr.UpdateCassettes)
	require.NoError(t, err)

	defer vcrCleanupFunc()

	ctx := t.Context()
	ctx, err = NewTestContext(ctx, httpClient)
	require.NoError(t, err)

	defer runCleanup(ctx, t, config.Cleanup)

	steps := config.Steps
	if len(steps) == 0 {
		steps = []TestStep{{Config: config.Config, Checks: config.Checks}}
	}

	for i, step := range steps {
		t.Logf("Running step %d/%d", i+1, len(steps))

		cassettePath := filepath.Join(cassettesDir, fmt.Sprintf("step-%d.cassette", i+1))
		if !runStep(ctx, t, step, cassettePath) {
			t.Errorf("Step %d failed, skipping the next steps", i+1)

			return
		}
	}
}

func runStep(ctx context.Context, t *testing.T, step TestStep, cassettePath string) bool {
	t.Helper()

	output, err := packerExec(ctx, t.TempDir(), step.Config, step.Args, cassettePath, !vcr.UpdateCassettes)

	switch {
	case step.ExpectError == nil && err != nil:
		t.Errorf("error executing packer command: %s", err)

		return false
	case step.ExpectError != nil && err == nil:
		t.Errorf("expected packer command to fail with an output matching %q", step.ExpectError)

		return false
	case step.ExpectError != nil && !step.ExpectError.MatchString(output):
		t.Errorf("expected packer output to match %q", step.ExpectError)

		return false
	}

	for i, check := range step.Checks {
		t.Logf("Running check %d/%d: %s", i+1, len(step.Checks), check.CheckName())

		err = check.Check(ctx)
		if err != nil {
			t.Errorf("Packer check %d failed: %s", i+1, err.Error())
		}
	}

	return true
}

func runCleanup(ctx context.Context, t *testing.T, cleanups []PackerCleanup) {
	t.Helper()

	for i, cleanup := range cleanups {
		t.Logf("Running cleanup func %d/%d", i+1, len(cleanups))

		err := cleanup.Cleanup(ctx, t)
		if err != nil {
			t.Errorf("Packer cleanup %d failed: %s", i+1, err.Error())
		}
	}
}
