#!/bin/bash

set -eux

echo "Install gomock"

go get github.com/golang/mock/gomock || True
# note we use v1.1.1 tag of gomock to not conflict with etcd
pushd ../../golang/mock
git checkout v1.1.1
popd
go install github.com/golang/mock/mockgen || True

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

items="interface"
mockgen -source=internal/pkg/pfsprovider/interface.go \
    -package mocks >internal/pkg/mocks/pfsprovider_mock.go
