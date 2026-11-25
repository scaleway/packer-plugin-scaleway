# End-to-end tests

This folder contains scripts and a makefile to help run end to end tests.

Tests run the real `packer` binary (it must be in your `PATH`). Before running them, the plugin is built
from the current code and installed in a temporary `PACKER_PLUGIN_PATH`, so an installed version of
the plugin is never used.

## Write a test

Create a new test in `../internal/tests` using `tester.Test`.

A test is made of steps. Each step runs `packer build` on its config, then runs its checks:

```go
tester.Test(t, &tester.TestConfig{
	Steps: []tester.TestStep{
		{
			Config: config,
			Checks: []tester.PackerCheck{checks.Image(zone, imageName)},
		},
		{
			Config:      config,
			ExpectError: regexp.MustCompile("is used by existing image"), // the build must fail
		},
		{
			Config: config,
			Args:   []string{"-force"}, // extra `packer build` arguments
			Checks: []tester.PackerCheck{checks.Image(zone, imageName)},
		},
	},
	Cleanup: []tester.PackerCleanup{cleanup.Image(zone, imageName)},
})
```

For a single build, `Config` and `Checks` can be set directly on `TestConfig` instead of `Steps`.
Cleanup functions always run, even if a step failed.

## Cassettes

To run easily in a CI, tests can be run while recording http requests. This allows pipelines to test without token by using recorded requests.

- To record cassettes, you must set `PACKER_UPDATE_CASSETTES=true`.
- To use recorded cassettes, you must set `PACKER_UPDATE_CASSETTES=false`

Each test has its own folder in `../internal/tests/testdata`, named after the test:

```text
testdata/<test-name>/
  step-1.cassette.yaml   # requests of the plugin during the build of step 1
  step-2.cassette.yaml
  tester.cassette.yaml   # requests of the checks and cleanup functions
```

When replaying cassettes, the plugin does not wait between requests while polling resources.

## Running tests

`PACKER_UPDATE_CASSETTES=true make test`

Test script will create a new project for you then run all tests before cleaning up the project

To record a single test:

`PACKER_UPDATE_CASSETTES=true go test ./internal/tests -run TestSimple -v`

## Test environment

Tests should run in an empty project. Tests will check for non deleted resources and a non-empty project will interfere with the checks.
