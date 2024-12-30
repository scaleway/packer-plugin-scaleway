#!/usr/bin/env bash

## This script will:
# - create a temporary project, configure it in environment
# - Run packer e2e tests
# - Tries to delete project

PROJECT_ID=$(scw account project create -otemplate="{{.ID}}")
export SCW_DEFAULT_PROJECT_ID=$PROJECT_ID

echo Running tests with new project $SCW_DEFAULT_PROJECT_ID

TESTS=(
  simple
  complete
)

TEST_RESULT=0

for TEST in "${TESTS[@]}"; do
  packer build ./$TEST/build_scaleway.pkr.hcl
  go run ./$TEST/
  test_status=$?
  if [ $test_status -ge 1 ]; then
    TEST_RESULT=1
  fi
done

scw instance image list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance image delete with-snapshots=true

# A security group will be created alongside the server during packer execution.
# We need to delete this security group before deleting the project
scw instance security-group list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance security-group delete

scw account project delete project-id="$SCW_DEFAULT_PROJECT_ID"

exit $TEST_RESULT
