package keystoreregistry

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeKeystore struct {
	watchChan KeyValueUpdateChan
}

func (fakeKeystore) Close() error {
	panic("implement me")
}
func (fakeKeystore) CleanPrefix(prefix string) error {
	panic("implement me")
}
func (fakeKeystore) Add(keyValues []KeyValue) error {
	panic("implement me")
}
func (fakeKeystore) Update(keyValues []KeyValueVersion) error {
	panic("implement me")
}
func (fakeKeystore) DeleteAll(keyValues []KeyValueVersion) error {
	panic("implement me")
}
func (fakeKeystore) GetAll(prefix string) ([]KeyValueVersion, error) {
	panic("implement me")
}
func (fakeKeystore) Get(key string) (KeyValueVersion, error) {
	panic("implement me")
}
func (fakeKeystore) WatchPrefix(prefix string, onUpdate func(old *KeyValueVersion, new *KeyValueVersion)) {
	panic("implement me")
}
func (fakeKeystore) WatchKey(ctxt context.Context, key string, onUpdate func(old *KeyValueVersion, new *KeyValueVersion)) {
	panic("implement me")
}
func (fk fakeKeystore) Watch(ctxt context.Context, key string, withPrefix bool) KeyValueUpdateChan {
	return fk.watchChan
}
func (fakeKeystore) KeepAliveKey(key string) error {
	panic("implement me")
}
func (fakeKeystore) NewMutex(lockKey string) (Mutex, error) {
	panic("implement me")
}

func TestPoolRegistry_GetNewHostBrickAllocations(t *testing.T) {
	rawEvents := make(chan KeyValueUpdate)
	reg := poolRegistry{keystore: &fakeKeystore{watchChan: rawEvents}}

	events := reg.GetNewHostBrickAllocations(context.TODO(), "host1")

	go func() {
		rawEvents <- KeyValueUpdate{IsCreate: false}
		rawEvents <- KeyValueUpdate{
			IsCreate: true,
			New: &KeyValueVersion{Value: toJson(registry.BrickAllocation{
				Hostname: "host1", Device: "sdb",
			})},
		}
		rawEvents <- KeyValueUpdate{IsCreate: false}
		close(rawEvents)
	}()

	ev1 := <-events
	assert.Equal(t, "host1", ev1.Hostname)
	assert.Equal(t, "sdb", ev1.Device)

	_, ok := <-events
	assert.False(t, ok)
	_, ok = <-rawEvents
	assert.False(t, ok)
}

func TestPoolRegistry_GetNewHostBrickAllocations_nil(t *testing.T) {
	reg := poolRegistry{keystore: &fakeKeystore{}}

	events := reg.GetNewHostBrickAllocations(context.TODO(), "host1")

	_, ok := <-events
	assert.False(t, ok)
}
