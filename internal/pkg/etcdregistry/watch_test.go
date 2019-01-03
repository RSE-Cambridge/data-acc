package etcdregistry

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeWatcher struct {
	t    *testing.T
	ch   clientv3.WatchChan
	opts []clientv3.OpOption
}

func (fw fakeWatcher) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	assert.Equal(fw.t, "key", key)
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
			opts: []clientv3.OpOption{clientv3.WithPrevKV()},
		},
	}

	response := keystore.Watch(context.TODO(), "key", false)

	assert.Empty(t, response)
}

func TestEtcKeystore_Watch(t *testing.T) {
	ch := make(chan clientv3.WatchResponse)

	keystore := etcKeystore{
		Watcher: fakeWatcher{
			t: t, ch: ch,
			opts: []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithPrevKV()},
		},
	}

	go func(){
		ch <- clientv3.WatchResponse{
			Events: []*clientv3.Event{
				{Type: clientv3.EventTypePut, Kv: &mvccpb.KeyValue{Key:[]byte("key1")}},
				{Type: clientv3.EventTypePut, Kv: &mvccpb.KeyValue{Key:[]byte("key2")}},
		}}
		ch <- clientv3.WatchResponse{
			Events: []*clientv3.Event{
				{Type: clientv3.EventTypePut, Kv: &mvccpb.KeyValue{CreateRevision:1}},
			}}
		ch <- clientv3.WatchResponse{
			Events: []*clientv3.Event{
				{Type: clientv3.EventTypeDelete, Kv: &mvccpb.KeyValue{}},
				{Type: clientv3.EventTypeDelete, Kv: &mvccpb.KeyValue{}},
			}}
		close(ch)
	}()

	response := keystore.Watch(context.TODO(), "key", true)

	ev1 := <- response
	assert.True(t, ev1.IsCreate)
	assert.False(t, ev1.IsModify)
	assert.False(t, ev1.IsDelete)

	ev2 := <- response
	assert.True(t, ev2.IsCreate)
	assert.False(t, ev2.IsModify)
	assert.False(t, ev2.IsDelete)

	ev3 := <- response
	assert.False(t, ev3.IsCreate)
	assert.True(t, ev3.IsModify)
	assert.False(t, ev3.IsDelete)

	ev4 := <- response
	assert.False(t, ev4.IsCreate)
	assert.False(t, ev4.IsModify)
	assert.True(t, ev4.IsDelete)

	ev5 := <- response
	assert.False(t, ev5.IsCreate)
	assert.False(t, ev5.IsModify)
	assert.True(t, ev5.IsDelete)

	// Check chan is closed
	_, ok := <- response
	assert.False(t, ok)
}
