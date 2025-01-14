#!/usr/bin/env bash

export SCW_DEFAULT_PROJECT_ID=c276a890-0f62-4d02-a5c5-9f86d615a029

scw instance image list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance image delete with-snapshots=true

# A security group will be created alongside the server during packer execution.
# We need to delete this security group before deleting the project
scw instance security-group list zone=all project-id="$SCW_DEFAULT_PROJECT_ID" -otemplate="zone={{.Zone}} {{.ID}}" | xargs -L1 -P1 scw instance security-group delete

scw account project delete project-id="$SCW_DEFAULT_PROJECT_ID"
