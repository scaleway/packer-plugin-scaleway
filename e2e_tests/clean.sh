#!/usr/bin/env bash

## This script will clean a project and its resources.

if [ "$#" -ne 1 ]
then
  echo "Usage: $0 [project-id]"
  exit 1
fi

export SCW_DEFAULT_PROJECT_ID=$1

# Clean images
scw instance image list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance image delete with-snapshots=true
# Clean volumes
scw instance volume list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance volume delete

# A security group will be created alongside the server during packer execution.
# We need to delete this security group before deleting the project
scw instance security-group list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance security-group delete

scw account project delete project-id="$SCW_DEFAULT_PROJECT_ID"
