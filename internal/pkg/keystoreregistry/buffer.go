package keystoreregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/oldregistry"
	"log"
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

func NewBufferRegistry(keystore Keystore) oldregistry.BufferRegistry {
	return &BufferRegistry{keystore}
}

func getBufferKey(buffer oldregistry.Buffer) string {
	return fmt.Sprintf("/buffers/%s", buffer.Name)
}

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	buffer := new(bytes.Buffer)
	buffer.Write(b)
	buffer.Write([]byte("\n"))
	return buffer.String()
}

func getBufferValue(buffer oldregistry.Buffer) string {
	return toJson(buffer)
}

func (r *BufferRegistry) AddBuffer(buffer oldregistry.Buffer) error {
	r.keystore.AtomicAdd(getBufferKey(buffer), getBufferValue(buffer))
	return nil
}

func (r *BufferRegistry) UpdateBuffer(buffer oldregistry.Buffer) (oldregistry.Buffer, error) {
	return oldregistry.Buffer{}, nil
}

func (r *BufferRegistry) RemoveBuffer(buffer oldregistry.Buffer) {
	r.keystore.CleanPrefix(getBufferKey(buffer))
}
