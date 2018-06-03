package etcdregistry

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"log"
	"os"
	"strings"
)

type Keystore interface {
	Close() error
	CleanPrefix(prefix string)
	AtomicAdd(key string, value string)
	WatchPutPrefix(prefix string, onPut func(string, string))
}

type etcKeystore struct {
	*clientv3.Client
}

func getEndpoints() []string {
	endpoints := os.Getenv("ETCD_ENDPOINTS")
	if endpoints == "" {
		log.Fatalf("Must set ETCD_ENDPOINTS environemnt variable, e.g. export ETCD_ENDPOINTS=127.0.0.1:2379")
	}
	return strings.Split(endpoints, ",")
}

func NewKeystore() Keystore {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: getEndpoints(),
	})
	if err != nil {
		log.Fatal(err)
		fmt.Println("Oh dear failed to create client...")
		panic(err)
	}
	return &etcKeystore{cli}
}

func (client *etcKeystore) CleanPrefix(prefix string) {
	kvc := clientv3.NewKV(client.Client)
	fmt.Println(kvc.Get(context.Background(), prefix, clientv3.WithPrefix()))
	kvc.Delete(context.Background(), prefix, clientv3.WithPrefix())
}

func (client *etcKeystore) AtomicAdd(key string, value string) {
	kvc := clientv3.NewKV(client.Client)
	response, err := kvc.Txn(context.Background()).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, value)).
		Commit()
	if err != nil {
		panic(err)
	}
	if !response.Succeeded {
		panic("oh dear someone has added the key already")
	}
}

func (client *etcKeystore) WatchPutPrefix(prefix string, onPut func(key string, value string)) {
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
