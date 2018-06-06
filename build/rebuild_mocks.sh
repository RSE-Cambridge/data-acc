set -eux

echo "Regenerate mocks:"                                                       
mockgen -source=internal/pkg/keystoreregistry/keystore.go \
    -package keystoreregistry >internal/pkg/keystoreregistry/mocks.go

