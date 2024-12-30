# End-to-end tests

This folder contains packer projects with each one go binaries that will check the behavior of the packer plugin.

## Test environment

- Tests should run in an empty project. Tests will check for non deleted resources and a non-empty project will interfere with the checks.

## Write a test

Copy `simple/` to your new test folder. Edit the packer file to test your feature then change asserts in `test.go`

## Running tests

`make test`

Test script will create a new project for you then run all tests before cleaning up the project
