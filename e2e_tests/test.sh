#!/usr/bin/env bash

## This script will:
# - create a temporary project, configure it in environment
# - Run packer e2e tests
# - Tries to delete project

PROJECT_ID=$(scw account project create -otemplate="{{.ID}}")
export SCW_DEFAULT_PROJECT_ID=$PROJECT_ID

echo Running tests with new project $SCW_DEFAULT_PROJECT_ID

TESTS=(
#  simple
  complete
)

TEST_RESULT=0

go test -c ./tests
./tests.test -test.v
TEST_RESULT=$?

./clean.sh $SCW_DEFAULT_PROJECT_ID

exit $TEST_RESULT
