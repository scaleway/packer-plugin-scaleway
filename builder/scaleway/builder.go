// The scaleway package contains a packersdk.Builder implementation
// that builds Scaleway images (snapshots).

package scaleway

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/packer-plugin-scaleway/internal/vcr"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// BuilderID is the unique id for the builder
const BuilderID = "hashicorp.scaleway"

var acceptanceTests = flag.Bool("run-acceptance-tests", os.Getenv("PACKER_ACC") == "1", "Run acceptance tests")

type Builder struct {
	Config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.Config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.Config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) { //nolint:ireturn
	scwZone, err := scw.ParseZone(b.Config.Zone)
	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}

	clientOpts := []scw.ClientOption{
		scw.WithDefaultProjectID(b.Config.ProjectID),
		scw.WithAuth(b.Config.AccessKey, b.Config.SecretKey),
		scw.WithDefaultZone(scwZone),
		scw.WithUserAgent(b.Config.UserAgent),
	}

	if b.Config.APIURL != "" {
		clientOpts = append(clientOpts, scw.WithAPIURL(b.Config.APIURL))
	}

	// Only use cassette if vcr.UpdateCassettesEnvVariable env variable is used.
	// It must at least be set to false when wanting to use local cassettes.
	if _, isSet := os.LookupEnv(vcr.UpdateCassettesEnvVariable); isSet {
		client, cleanup, err := vcr.GetHTTPRecorder(filepath.Join("testdata", b.Config.ImageName+".cassette"), vcr.UpdateCassettes)
		if err != nil {
			ui.Error(err.Error())
			return nil, err
		}
		defer cleanup()

		clientOpts = append(clientOpts, scw.WithHTTPClient(client))
	}

	client, err := scw.NewClient(clientOpts...)
	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.Config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepPreValidate{
			Force:        b.Config.PackerForce,
			ImageName:    b.Config.ImageName,
			SnapshotName: b.Config.SnapshotName,
		},
		&stepCreateSSHKey{
			Debug:        b.Config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("scw_%s.pem", b.Config.PackerBuildName),
		},
		new(stepCreateVolume),
		new(stepCreateServer),
		new(stepServerInfo),
		&communicator.StepConnect{
			Config:    &b.Config.Comm,
			Host:      communicator.CommHost(b.Config.Comm.Host(), "server_ip"),
			SSHConfig: b.Config.Comm.SSHConfigFunc(),
		},
		new(commonsteps.StepProvision),
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.Config.Comm,
		},
		new(stepWaitUserData),
		new(stepCleanupMachineData),
		new(stepShutdown),
		new(stepBackup),
	}

	if *acceptanceTests {
		steps = append(steps, new(stepSweep))
	}

	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.Config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("build was cancelled")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("build was halted")
	}

	if _, ok := state.GetOk("snapshots"); !ok {
		return nil, errors.New("cannot find snapshot_name in state")
	}

	artifact := &Artifact{
		ImageName: state.Get("image_name").(string),
		ImageID:   state.Get("image_id").(string),
		Snapshots: state.Get("snapshots").([]ArtifactSnapshot),
		ZoneName:  b.Config.Zone,
		Client:    client,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}
