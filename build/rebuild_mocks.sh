#!/bin/bash

set -eux

echo "Regenerate mocks:"

mkdir -p internal/pkg/mocks

items="session session_action_handler"
for i in $items; do
    mockgen -source=internal/pkg/facade/${i}.go \
        >internal/pkg/mock_facade/${i}.go
done

items="disk"
for i in $items; do
    mockgen -source=internal/pkg/fileio/${i}.go \
        >internal/pkg/mock_fileio/${i}_mock.go
done

items="provider ansible"
for i in $items; do
    mockgen -source=internal/pkg/filesystem/${i}.go \
        >internal/pkg/mock_filesystem/${i}.go
done

items="brick_allocation brick_host session session_actions"
for i in $items; do
    mockgen -source=internal/pkg/registry/${i}.go \
        >internal/pkg/mock_registry/${i}.go
done

items="keystore"
for i in $items; do
    mockgen -source=internal/pkg/store/${i}.go \
        >internal/pkg/mock_store/${i}.go
done
