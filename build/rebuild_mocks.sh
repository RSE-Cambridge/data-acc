#!/bin/bash

set -eux

echo "Regenerate mocks:"

mkdir -p internal/pkg/mocks

items="pool volume"
for i in $items; do
    mockgen -source=internal/pkg/registry/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done

items="job"
for i in $items; do
    mockgen -source=internal/pkg/dacctl/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done

items="disk"
for i in $items; do
    mockgen -source=internal/pkg/fileio/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done

mockgen -source=internal/pkg/pfsprovider/interface.go \
    -package mocks >internal/pkg/mocks/pfsprovider_mock.go

items="allocation brick pool session session_actions"
for i in $items; do
    mockgen -source=internal/pkg/v2/registry/${i}.go \
        >internal/pkg/v2/mock_registry/${i}.go
done

items="session session_action_handler"
for i in $items; do
    mockgen -source=internal/pkg/v2/workflow/${i}.go \
        >internal/pkg/v2/mock_workflow/${i}.go
done

items="keystore"
for i in $items; do
    mockgen -source=internal/pkg/v2/store/${i}.go \
        >internal/pkg/v2/mock_store/${i}.go
done
