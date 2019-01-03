package etcdregistry

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeWatcher struct {
	t    *testing.T
	ch   clientv3.WatchChan
	opts []clientv3.OpOption
}

func (fw fakeWatcher) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	assert.Equal(fw.t, "key1", key)
	assert.EqualValues(fw.t, len(fw.opts), len(opts)) // TODO: how to assert this properly?
	return fw.ch
}
func (fakeWatcher) Close() error {
	panic("implement me")
}

func TestEtcKeystore_Watch_Nil(t *testing.T) {
	keystore := etcKeystore{
		Watcher: fakeWatcher{
			t: t, ch: nil,
			opts: []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithPrevKV()},
		},
	}

	response := keystore.Watch(context.TODO(), "key1", true)

	assert.Empty(t, response)
}
