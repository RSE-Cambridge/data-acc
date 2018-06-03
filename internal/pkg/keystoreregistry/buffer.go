package keystoreregistry

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
)

type Keystore interface {
	Close() error
	CleanPrefix(prefix string)
	AtomicAdd(key string, value string)
	WatchPutPrefix(prefix string, onPut func(string, string))
}

type BufferRegistry struct {
	keystore Keystore
}

func NewBufferRegistry(keystore Keystore) registry.BufferRegistry {
	return &BufferRegistry{keystore}
}

func getBufferKey(buffer registry.Buffer) string {
	return fmt.Sprintf("/buffer/%s", buffer.Name)
}

func getBufferValue(buffer registry.Buffer) string {
	return buffer.Owner
}

func (r *BufferRegistry) AddBuffer(buffer registry.Buffer) error {
	r.keystore.AtomicAdd(getBufferKey(buffer), getBufferValue(buffer))
	return nil
}

func (r *BufferRegistry) UpdateBuffer(buffer registry.Buffer) (registry.Buffer, error) {
	return registry.Buffer{}, nil
}

func (r *BufferRegistry) RemoveBuffer(buffer registry.Buffer) {
	r.keystore.CleanPrefix(getBufferKey(buffer))
}
