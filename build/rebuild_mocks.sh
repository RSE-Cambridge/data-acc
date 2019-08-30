#!/bin/bash

set -eux

echo "Regenerate mocks:"

mkdir -p internal/pkg/mocks

items="disk"
for i in $items; do
    mockgen -source=internal/pkg/fileio/${i}.go \
        >internal/pkg/mock_fileio/${i}_mock.go
done

items="session session_action_handler"
for i in $items; do
    mockgen -source=internal/pkg/v2/facade/${i}.go \
        >internal/pkg/v2/mock_facade/${i}.go
done

items="keystore"
for i in $items; do
    mockgen -source=internal/pkg/v2/store/${i}.go \
        >internal/pkg/v2/mock_store/${i}.go
done

items="provider ansible"
for i in $items; do
    mockgen -source=internal/pkg/v2/filesystem/${i}.go \
        >internal/pkg/v2/mock_filesystem/${i}.go
done
