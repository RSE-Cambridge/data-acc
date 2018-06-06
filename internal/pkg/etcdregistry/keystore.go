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

// TODO: this should be private
func NewEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: getEndpoints(),
	})
	if err != nil {
		log.Fatal(err)
		fmt.Println("Oh dear failed to create client...")
		panic(err)
	}
	return cli
}

func NewKeystore() keystoreregistry.Keystore {
	cli := NewEtcdClient()
	return &EtcKeystore{cli}
}

// TODO: this should be private, once abstraction finished
type EtcKeystore struct {
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

func (client *EtcKeystore) Add(keyValues []keystoreregistry.KeyValue) error {
	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op
	for _, keyValue := range keyValues {
		ifOps = append(ifOps, clientv3util.KeyMissing(keyValue.Key))
		thenOps = append(thenOps, clientv3.OpPut(keyValue.Key, keyValue.Value))
	}

	kvc := clientv3.NewKV(client.Client)
	kvc.Txn(context.Background())
	response, err := kvc.Txn(context.Background()).If(ifOps...).Then(thenOps...).Commit()
	handleError(err)

	if !response.Succeeded {
		return fmt.Errorf("unable to add all the key values")
	}
	return nil
}

func (client *EtcKeystore) Update(keyValues []keystoreregistry.KeyValueVersion) error {
	panic("implement me")
}

func getKeyValueVersion(rawKeyValue *mvccpb.KeyValue) *keystoreregistry.KeyValueVersion {
	return &keystoreregistry.KeyValueVersion{
		Key:            string(rawKeyValue.Key),
		Value:          string(rawKeyValue.Value),
		ModRevision:    rawKeyValue.ModRevision,
		CreateRevision: rawKeyValue.CreateRevision,
	}
}
func (client *EtcKeystore) GetAll(prefix string) ([]keystoreregistry.KeyValueVersion, error) {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Get(context.Background(), prefix, clientv3.WithPrefix())
	handleError(err)

	if response.Count == 0 {
		return []keystoreregistry.KeyValueVersion{},
			fmt.Errorf("Unable to find any values for prefix: %s", prefix)
	}
	var values []keystoreregistry.KeyValueVersion
	for _, rawKeyValue := range response.Kvs {
		values = append(values, *getKeyValueVersion(rawKeyValue))
	}
	return values, nil
}

func (client *EtcKeystore) Get(key string) (keystoreregistry.KeyValueVersion, error) {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Get(context.Background(), key)
	handleError(err)

	value := keystoreregistry.KeyValueVersion{}

	if response.Count == 0 {
		return value, fmt.Errorf("Unable to find any values for key: %s", key)
	}
	if response.Count > 1 {
		panic(errors.New("should never get more than one value for get"))
	}

	return *getKeyValueVersion(response.Kvs[0]), nil
}

func (client *EtcKeystore) WatchPrefix(prefix string, onUpdate func(old keystoreregistry.KeyValueVersion, new keystoreregistry.KeyValueVersion)) (int64, error) {
	panic("implement me")
}

func (client *EtcKeystore) CleanPrefix(prefix string) error {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Delete(context.Background(), prefix, clientv3.WithPrefix())
	handleError(err)

	if response.Deleted == 0 {
		return fmt.Errorf("No keys with prefix: %s", prefix)
	}

	log.Printf("Cleaned %d keys with prefix: '%s'.\n", response.Deleted, prefix)
	// TODO return deleted count
	return nil
}

func (client *EtcKeystore) AtomicAdd(key string, value string) {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Txn(context.Background()).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, value)).
		Commit()
	if err != nil {
		panic(err)
	}
	if !response.Succeeded {
		panic(fmt.Errorf("oh dear someone has added the key already: %s", key))
	}
}

func (client *EtcKeystore) WatchPutPrefix(prefix string, onPut func(key string, value string)) {
	rch := client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type.String() == "PUT" {
				onPut(string(ev.Kv.Key), string(ev.Kv.Value))
			} else {
				fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		}
	}
}
