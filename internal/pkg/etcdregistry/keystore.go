package etcdregistry

import (
	"context"
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"log"
	"os"
	"strings"
)

func getEndpoints() []string {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		endpoints = os.Getenv("ETCD_ENDPOINTS")
	}
	if endpoints == "" {
		log.Fatalf("Must set ETCDCTL_ENDPOINTS environemnt variable, e.g. export ETCDCTL_ENDPOINTS=127.0.0.1:2379")
	}
	return strings.Split(endpoints, ",")
}

func newEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: getEndpoints(),
	})
	if err != nil {
		fmt.Println("failed to create client")
		log.Fatal(err)
	}
	return cli
}

func NewKeystore() keystoreregistry.Keystore {
	cli := newEtcdClient()
	return &etcKeystore{cli}
}

type etcKeystore struct {
	*clientv3.Client
}

func handleError(err error) {
	if err != nil {
		switch err {
		case context.Canceled:
			log.Printf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Printf("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Printf("client-side error: %v", err)
		default:
			log.Printf("bad cluster endpoints, which are not etcd servers: %v", err)
		}
		log.Fatal(err)
	}
}

func (client *etcKeystore) runTransaction(ifOps []clientv3.Cmp, thenOps []clientv3.Op) error {
	kvc := clientv3.NewKV(client.Client)
	kvc.Txn(context.Background())
	response, err := kvc.Txn(context.Background()).If(ifOps...).Then(thenOps...).Commit()
	handleError(err)

	if !response.Succeeded {
		return fmt.Errorf("unable to add all the key values")
	}
	return nil
}

func (client *etcKeystore) Add(keyValues []keystoreregistry.KeyValue) error {
	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op
	for _, keyValue := range keyValues {
		ifOps = append(ifOps, clientv3util.KeyMissing(keyValue.Key))
		thenOps = append(thenOps, clientv3.OpPut(keyValue.Key, keyValue.Value))
	}
	return client.runTransaction(ifOps, thenOps)
}

func (client *etcKeystore) Update(keyValues []keystoreregistry.KeyValueVersion) error {
	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op
	for _, keyValue := range keyValues {
		if keyValue.ModRevision > 0 {
			ifOps = append(ifOps, clientv3util.KeyExists(keyValue.Key)) // only add new keys if ModRevision == 0
			checkModRev := clientv3.Compare(clientv3.ModRevision(keyValue.Key), "=", keyValue.ModRevision)
			ifOps = append(ifOps, checkModRev)
		}
		thenOps = append(thenOps, clientv3.OpPut(keyValue.Key, keyValue.Value))
	}
	return client.runTransaction(ifOps, thenOps)
}

func (client *etcKeystore) DeleteAll(keyValues []keystoreregistry.KeyValueVersion) error {
	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op
	for _, keyValue := range keyValues {
		ifOps = append(ifOps, clientv3util.KeyExists(keyValue.Key))
		if keyValue.ModRevision > 0 {
			checkModRev := clientv3.Compare(clientv3.ModRevision(keyValue.Key), "=", keyValue.ModRevision)
			ifOps = append(ifOps, checkModRev)
		}
		thenOps = append(thenOps, clientv3.OpDelete(keyValue.Key))
	}
	return client.runTransaction(ifOps, thenOps)
}

func getKeyValueVersion(rawKeyValue *mvccpb.KeyValue) *keystoreregistry.KeyValueVersion {
	if rawKeyValue == nil {
		return nil
	}
	return &keystoreregistry.KeyValueVersion{
		Key:            string(rawKeyValue.Key),
		Value:          string(rawKeyValue.Value),
		ModRevision:    rawKeyValue.ModRevision,
		CreateRevision: rawKeyValue.CreateRevision,
	}
}

func (client *etcKeystore) GetAll(prefix string) ([]keystoreregistry.KeyValueVersion, error) {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Get(context.Background(), prefix, clientv3.WithPrefix())
	handleError(err)

	if response.Count == 0 {
		return []keystoreregistry.KeyValueVersion{},
			fmt.Errorf("unable to find any values for prefix: %s", prefix)
	}
	var values []keystoreregistry.KeyValueVersion
	for _, rawKeyValue := range response.Kvs {
		values = append(values, *getKeyValueVersion(rawKeyValue))
	}
	return values, nil
}

func (client *etcKeystore) Get(key string) (keystoreregistry.KeyValueVersion, error) {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Get(context.Background(), key)
	handleError(err)

	value := keystoreregistry.KeyValueVersion{}

	if response.Count == 0 {
		return value, fmt.Errorf("unable to find any values for key: %s", key)
	}
	if response.Count > 1 {
		panic(errors.New("should never get more than one value for get"))
	}

	return *getKeyValueVersion(response.Kvs[0]), nil
}

func (client *etcKeystore) WatchPrefix(prefix string,
	onUpdate func(old *keystoreregistry.KeyValueVersion, new *keystoreregistry.KeyValueVersion)) {
	rch := client.Watch(context.Background(), prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				new := getKeyValueVersion(ev.Kv)
				if new != nil && new.CreateRevision == 0 {
					// show deleted by returning nil
					new = nil
				}
				old := getKeyValueVersion(ev.PrevKv)
				onUpdate(old, new)
			}
		}
	}()
}

func (client *etcKeystore) WatchKey(ctxt context.Context, key string,
	onUpdate func(old *keystoreregistry.KeyValueVersion, new *keystoreregistry.KeyValueVersion)) {
	rch := client.Watch(ctxt, key, clientv3.WithPrevKV())
	go func() {
		for watchResponse := range rch {
			if watchResponse.Canceled {
				return
			}
			for _, ev := range watchResponse.Events {
				new := getKeyValueVersion(ev.Kv)
				if new != nil && new.CreateRevision == 0 {
					// show deleted by returning nil
					new = nil
				}
				old := getKeyValueVersion(ev.PrevKv)

				onUpdate(old, new)
			}
		}
	}()
}

func (client *etcKeystore) KeepAliveKey(key string) error {
	// TODO what about configure timeout and ttl?
	var ttl int64 = 10
	grantResponse, err := client.Grant(context.Background(), ttl)
	if err != nil {
		log.Fatal(err)
	}
	leaseID := grantResponse.ID

	kvc := clientv3.NewKV(client.Client)
	txnResponse, err := kvc.Txn(context.Background()).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, "keep-alive", clientv3.WithLease(leaseID), clientv3.WithPrevKV())).
		Commit()
	handleError(err)
	if !txnResponse.Succeeded {
		return fmt.Errorf("unable to create keep-alive key: %s", key)
	}

	ch, err := client.KeepAlive(context.Background(), leaseID)
	if err != nil {
		log.Fatal(err)
	}

	counter := 9
	go func() {
		for {
			ka := <-ch
			if ka == nil {
				log.Println("Refresh stoped for key: ", key)
				// TODO: optionally make this log a fatal error?
				break
			} else {
				if counter >= 9 {
					counter = 0
					log.Println("Still refreshing key:", key)
				} else {
					counter++
				}
			}
		}
	}()
	return nil
}

// TODO... old methods may need removing....

func (client *etcKeystore) CleanPrefix(prefix string) error {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Delete(context.Background(), prefix, clientv3.WithPrefix())
	handleError(err)

	if response.Deleted == 0 {
		return fmt.Errorf("no keys with prefix: %s", prefix)
	}

	log.Printf("Cleaned %d keys with prefix: '%s'.\n", response.Deleted, prefix)
	// TODO return deleted count
	return nil
}
