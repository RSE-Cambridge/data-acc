package etcdregistry

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/coreos/etcd/clientv3"
)

func (client *etcKeystore) Watch(ctxt context.Context, key string, withPrefix bool) keystoreregistry.KeyValueUpdateChan {
	options := []clientv3.OpOption{clientv3.WithPrevKV()}
	if withPrefix {
		options = append(options, clientv3.WithPrefix())
	}
	rch := client.Client.Watch(ctxt, key, options...)

	c := make(chan keystoreregistry.KeyValueUpdate)

	go processWatchEvents(rch, c)

	return c
}

func processWatchEvents(watchChan clientv3.WatchChan, c chan keystoreregistry.KeyValueUpdate) {
	for watchResponse := range watchChan {
		// if error, send empty update with an error
		err := watchResponse.Err()
		if err != nil {
			c <- keystoreregistry.KeyValueUpdate{Err: err}
		}

		// send all events in this watch response
		for _, ev := range watchResponse.Events {
			update := keystoreregistry.KeyValueUpdate{
				IsCreate: ev.IsCreate(),
				IsModify: ev.IsModify(),
				IsDelete: ev.Type == clientv3.EventTypeDelete,
			}
			if update.IsCreate || update.IsModify {
				update.New = getKeyValueVersion(ev.Kv)
			}
			if update.IsDelete || update.IsModify {
				update.Old = getKeyValueVersion(ev.PrevKv)
			}

			c <- update
		}
	}

	// Assuming we get here when the context is cancelled or hits its timeout
	// i.e. there are no more events, so we close the channel
	close(c)
}
