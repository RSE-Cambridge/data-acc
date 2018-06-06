set -eux

echo "Regenerate mocks:"                                                       
mockgen -source=internal/pkg/keystoreregistry/helpers.go >internal/pkg/mock_keystoregistry/keystore.go

