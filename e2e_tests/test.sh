#!/usr/bin/env bash

## This script will:
# - create a temporary project, configure it in environment
# - Run packer e2e tests
# - Tries to delete project

if [ "$PACKER_UPDATE_CASSETTES" == "true" ]
then
  PROJECT_ID=$(scw account project create -otemplate="{{.ID}}")
  export SCW_DEFAULT_PROJECT_ID=$PROJECT_ID

  echo Running tests with new project $SCW_DEFAULT_PROJECT_ID
else
  export SCW_ACCESS_KEY=SCWXXXXXXXXXXXXXFAKE
  export SCW_SECRET_KEY=11111111-1111-1111-1111-111111111111
  export SCW_DEFAULT_PROJECT_ID=11111111-1111-1111-1111-111111111111
  echo Using cassettes, no test project was created
fi


TESTS=(
#  simple
  complete
)

TEST_RESULT=0

rm ./tests.test
go test -c ../internal/tests
./tests.test -test.v
TEST_RESULT=$?

if [ "$PACKER_UPDATE_CASSETTES" == "true" ]
then
  ./clean.sh $SCW_DEFAULT_PROJECT_ID
fi

exit $TEST_RESULT
