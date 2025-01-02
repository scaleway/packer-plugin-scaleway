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

for TEST in "${TESTS[@]}"; do
  packer build ./$TEST/build_scaleway.pkr.hcl
  go test ./$TEST/
  test_status=$?
  if [ $test_status -ge 1 ]; then
    TEST_RESULT=1
  fi
done

exit $TEST_RESULT
