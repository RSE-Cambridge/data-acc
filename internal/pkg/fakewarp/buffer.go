package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
)

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func DeleteBuffer(c CliContext) error {
	error := processDeleteBuffer(c.String("token"), etcdregistry.NewKeystore())
	return error
}

func processDeleteBuffer(bufferName string, keystore keystoreregistry.Keystore) error {
	r := keystoreregistry.NewBufferRegistry(keystore)
	// TODO: should do a get buffer before doing a delete
	buf := registry.Buffer{Name: bufferName}
	r.AddBuffer(buf)
	return nil
}
