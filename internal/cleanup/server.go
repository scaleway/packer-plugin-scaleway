package cleanup

import (
	"context"
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*ServerCleanup)(nil)

type ServerCleanup struct {
	zone scw.Zone
	name string
}

func Server(zone scw.Zone, name string) *ServerCleanup {
	return &ServerCleanup{
		zone: zone,
		name: name,
	}
}

func (i *ServerCleanup) Cleanup(ctx context.Context, t *testing.T) error {
	t.Helper()

	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListServers(&instance.ListServersRequest{
		Name:    &i.name,
		Zone:    i.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list instance servers: %w", err)
	}

	if len(resp.Servers) == 0 {
		return fmt.Errorf("could not find any instance server named %q", i.name)
	}

	if len(resp.Servers) > 1 {
		return fmt.Errorf("found multiple instance servers named %q", i.name)
	}

	err = api.DeleteServer(&instance.DeleteServerRequest{
		Zone:     i.zone,
		ServerID: resp.Servers[0].ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete instance server: %w", err)
	}

	t.Logf("deleted instance server %q\n", resp.Servers[0].Name)

	return nil
}
