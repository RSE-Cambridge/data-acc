set -eux

echo "Regenerate mocks:"

items="keystore"
for i in $items; do
    mockgen -source=internal/pkg/keystoreregistry/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done

items="pool volume"
for i in $items; do
    mockgen -source=internal/pkg/registry/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done

items="job"
for i in $items; do
    mockgen -source=internal/pkg/fakewarp/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done

items="reader"
for i in $items; do
    mockgen -source=internal/pkg/fileio/${i}.go \
        -package mocks >internal/pkg/mocks/${i}_mock.go
done
