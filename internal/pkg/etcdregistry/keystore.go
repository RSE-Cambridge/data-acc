package etcdregistry

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
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
		log.Fatalf("Must set ETCD_ENDPOINTS environemnt variable, e.g. export ETCD_ENDPOINTS=127.0.0.1:2379")
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

func (client *EtcKeystore) Add(keyValues []keystoreregistry.KeyValue) error {
	panic("implement me")
}

func (client *EtcKeystore) Update(keyValues []keystoreregistry.KeyValueVersion) error {
	panic("implement me")
}

func (client *EtcKeystore) GetAll(prefix string) ([]keystoreregistry.KeyValueVersion, error) {
	panic("implement me")
}

func (client *EtcKeystore) Get(key string) (keystoreregistry.KeyValueVersion, error) {
	panic("implement me")
}

func (client *EtcKeystore) WatchPrefix(prefix string, onUpdate func(old keystoreregistry.KeyValueVersion, new keystoreregistry.KeyValueVersion)) (int64, error) {
	panic("implement me")
}

func (client *EtcKeystore) CleanPrefix(prefix string) error {
	kvc := clientv3.NewKV(client.Client)
	fmt.Println(kvc.Get(context.Background(), prefix, clientv3.WithPrefix()))
	response, error := kvc.Delete(context.Background(), prefix, clientv3.WithPrefix())
	if error != nil {
		panic(error)
	}
	if response.Deleted == 0 {
		panic(fmt.Errorf("oh dear, nothing to delete for prefix: %s", prefix))
	}
	// TODO: should return errors rather than panic here
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
