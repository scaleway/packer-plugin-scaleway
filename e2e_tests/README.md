# End-to-end tests

This folder contains scripts and a makefile to help run end to end tests.

## Write a test

Create a new test in `../internal/tests`.

## Cassettes

To run easily in a CI, tests can be run while recording http requests. This allows pipelines to test without token by using recorded requests.

- To record cassettes, you must set `PACKER_UPDATE_CASSETTES=true`.
- To use recorded cassettes, you must set `PACKER_UPDATE_CASSETTES=false`

## Running tests

`PACKER_UPDATE_CASSETTES=true make test`

Test script will create a new project for you then run all tests before cleaning up the project

## Test environment

Tests should run in an empty project. Tests will check for non deleted resources and a non-empty project will interfere with the checks.
