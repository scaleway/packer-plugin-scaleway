package checks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*ServerCheck)(nil)

type ServerCheck struct {
	zone scw.Zone
	name string

	tags []string
}

func Server(zone scw.Zone, name string) *ServerCheck {
	return &ServerCheck{
		zone: zone,
		name: name,
	}
}

func (c *ServerCheck) Tags(tags []string) *ServerCheck {
	c.tags = tags

	return c
}

func findServer(ctx context.Context, zone scw.Zone, name string) (*instance.Server, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListServers(&instance.ListServersRequest{
		Zone:    zone,
		Name:    &name,
		Project: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list instance servers: %w", err)
	}

	if len(resp.Servers) == 0 {
		return nil, fmt.Errorf("no instance server named %q was found", name)
	}

	if len(resp.Servers) > 1 {
		return nil, fmt.Errorf("multiple instance servers found with name %q", name)
	}

	return resp.Servers[0], nil
}

func (c *ServerCheck) CheckName() string {
	return fmt.Sprintf("Instance server \"%s...\"", c.name)
}

func (c *ServerCheck) Check(ctx context.Context) error {
	server, err := findServer(ctx, c.zone, c.name)
	if err != nil {
		return err
	}

	actualTags := make(map[string]bool, len(server.Tags))
	for _, tag := range server.Tags {
		actualTags[tag] = false
	}

	expectedTags := make(map[string]bool, len(c.tags))
	for _, tag := range c.tags {
		expectedTags[tag] = false
	}

	for serverTag := range actualTags {
		for tagToFind := range expectedTags {
			if serverTag == tagToFind {
				actualTags[serverTag] = true
				expectedTags[tagToFind] = true

				break
			}

			if strings.HasPrefix(serverTag, "AUTHORIZED_KEY=") {
				actualTags[serverTag] = true

				break
			}
		}
	}

	errs := []error{errors.New("server tags did not match")}

	for expectedTag, found := range expectedTags {
		if !found {
			errs = append(errs, fmt.Errorf("- missing expected tag: %q", expectedTag))
		}
	}

	for actualTag, expected := range actualTags {
		if !expected {
			errs = append(errs, fmt.Errorf("+ found unexpected tag: %q", actualTag))
		}
	}

	if len(errs) > 1 {
		return errors.Join(errs...)
	}

	return nil
}
