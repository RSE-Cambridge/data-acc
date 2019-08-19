package store

import (
	"context"
)

type Keystore interface {
	// Used to clean up any resources being used
	// ... such as a connection to etcd.
	Close() error

	// Atomically add all the key value pairs
	//
	// If an error occurs no keyvalues are written.
	// Error is returned if any key already exists.
	Create(key string, value []byte) (KeyValueVersion, error)

	// Update the specified key values, atomically
	//
	// If ModRevision is 0, it is ignored.
	// Otherwise if the revisions of any key doesn't
	// match the current revision of that key, the update fails.
	// When update fails an error is returned and no keyValues are updated
	Update(key string, value []byte, modRevision int64) (KeyValueVersion, error)

	// Delete the specified key values, atomically
	//
	// Similar to update, checks ModRevision matches current key,
	// ignores ModRevision if not zero.
	// If any keys are not currently present, the request fails.
	// Deletes no keys if an error is returned
	Delete(key string, modRevision int64) error

	// Removes all keys with given prefix
	DeleteAllKeysWithPrefix(keyPrefix string) error

	// Get all key values for a given prefix.
	GetAll(keyPrefix string) ([]KeyValueVersion, error)

	// Get given key
	Get(key string) (KeyValueVersion, error)

	// Get a channel containing all KeyValueUpdate events
	//
	// Use the context to control if you watch forever, or if you choose to cancel when a key
	// is deleted, or you stop watching after some timeout.
	Watch(ctxt context.Context, key string, withPrefix bool) KeyValueUpdateChan

	// Add a key, and remove it when calling process dies
	// Error is returned if the key already exists
	// can be cancelled via the context
	KeepAliveKey(ctxt context.Context, key string) error

	// Get a new mutex associated with the specified key
	NewMutex(lockKey string) (Mutex, error)
}

type KeyValueUpdateChan <-chan KeyValueUpdate

type KeyValueVersion struct {
	Key            string
	Value          []byte
	CreateRevision int64
	ModRevision    int64
}

type KeyValueUpdate struct {
	Old      *KeyValueVersion
	New      *KeyValueVersion
	IsCreate bool
	IsModify bool
	IsDelete bool
	Err      error
}

type Mutex interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}
