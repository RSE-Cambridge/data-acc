package oldregistry

import (
	"fmt"
)

type BufferRegistry interface {
	Close() error
	ClearAllData()
	AddBuffer(id int)
	AddSlice(id int, s string)
	AddMountpoint(id string, mountpoint string)
	WatchNewBuffer(callback func(string, string))
	WatchNewSlice(callback func(key string, value string))
	WatchNewReady(callback func(key string, value string))
}

type bufferRegistry struct {
	keystore Keystore
}

func NewBufferRegistry() BufferRegistry {
	keystore := NewKeystore()
	return &bufferRegistry{keystore}
}

func (registry *bufferRegistry) Close() error {
	return registry.keystore.Close()
}

func (registry *bufferRegistry) ClearAllData() {
	fmt.Println("Cleanup started")
	registry.keystore.CleanPrefix("/buffer")
	registry.keystore.CleanPrefix("/slice")
	registry.keystore.CleanPrefix("/ready")
	fmt.Println("Cleanup done")
}

func (registry *bufferRegistry) AddBuffer(id int) {
	var key = fmt.Sprintf("/buffer/%d", id)
	registry.keystore.AtomicAdd(key, "I am a new buffer")
}

func (registry *bufferRegistry) AddSlice(id int, value string) {
	var key = fmt.Sprintf("/slice/%d", id)
	registry.keystore.AtomicAdd(key, value)
}

// This is a big hack! id should be an int here
func (registry *bufferRegistry) AddMountpoint(id string, mountpoint string) {
	registry.keystore.AtomicAdd(fmt.Sprintf("/ready%s", id), mountpoint)
}

func (registry *bufferRegistry) WatchNewBuffer(callback func(string, string)) {
	registry.keystore.WatchPutPrefix("/buffer", callback)
}

func (registry *bufferRegistry) WatchNewSlice(callback func(key string, value string)) {
	registry.keystore.WatchPutPrefix("/slice", callback)
}

func (registry *bufferRegistry) WatchNewReady(callback func(key string, value string)) {
	registry.keystore.WatchPutPrefix("/ready", callback)
}
