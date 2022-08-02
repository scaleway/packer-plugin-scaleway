package e2e

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/require"
)

func loadProfile() (*scw.Profile, error) {
	config, err := scw.LoadConfig()
	// If the config file do not exist, don't return an error as we may find config in ENV or flags.
	if _, isNotFoundError := err.(*scw.ConfigFileNotFoundError); isNotFoundError {
		config = &scw.Config{}
	} else if err != nil {
		return nil, err
	}

	// By default we set default zone and region to fr-par
	defaultRegion := scw.RegionFrPar
	defaultZone := scw.ZoneFrPar1
	defaultZoneProfile := &scw.Profile{
		DefaultRegion: scw.StringPtr(defaultRegion.String()),
		DefaultZone:   scw.StringPtr(defaultZone.String()),
	}

	activeProfile, err := config.GetActiveProfile()
	if err != nil {
		return nil, err
	}
	envProfile := scw.LoadEnvProfile()

	profile := scw.MergeProfiles(defaultZoneProfile, activeProfile, envProfile)

	// If profile have a defaultZone but no defaultRegion we set the defaultRegion
	// to the one of the defaultZone
	if profile.DefaultZone != nil && *profile.DefaultZone != "" &&
		(profile.DefaultRegion == nil || *profile.DefaultRegion == "") {
		zone := scw.Zone(*profile.DefaultZone)
		region, err := zone.Region()
		if err == nil {
			profile.DefaultRegion = scw.StringPtr(region.String())
		}
	}
	return profile, nil
}

func NewClient(t *testing.T) (*scw.Client, func(), error) {
	// Create an http client with recording capabilities
	httpClient, cleanup, err := getHTTPRecoder(t, *updateCassettes)
	require.NoError(t, err)

	profile, err := loadProfile()
	if err != nil {
		return nil, nil, err
	}

	c, err := scw.NewClient(
		scw.WithHTTPClient(httpClient),
		scw.WithProfile(profile),
		/*		scw.WithUserAgent("protoc-gen-e2e-testing"),
		 */)
	if err != nil {
		return nil, nil, err
	}
	return c, cleanup, err
}
