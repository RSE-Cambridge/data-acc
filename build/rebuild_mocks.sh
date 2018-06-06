set -eux

echo "Regenerate mocks:"                                                       
mockgen -source=internal/pkg/keystoreregistry/helpers.go \
    -package keystoreregistry >internal/pkg/keystoreregistry/mocks.go

