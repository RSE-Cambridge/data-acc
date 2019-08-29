package store_impl

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/pkg/transport"
	"log"
	"os"
	"strings"
	"time"
)

func getTLSConfig() *tls.Config {
	certFile := os.Getenv("ETCDCTL_CERT_FILE")
	keyFile := os.Getenv("ETCDCTL_KEY_FILE")
	caFile := os.Getenv("ETCDCTL_CA_FILE")

	if certFile == "" || keyFile == "" || caFile == "" {
		return nil
	}

	tlsInfo := transport.TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: caFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	return tlsConfig
}

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
		Endpoints:   getEndpoints(),
		DialTimeout: 10 * time.Second,
		TLS:         getTLSConfig(),
	})
	if err != nil {
		fmt.Println("failed to create client")
		log.Fatal(err)
	}
	return cli
}

func NewKeystore() store.Keystore {
	cli := newEtcdClient()
	return &etcKeystore{
		Watcher: cli.Watcher,
		KV:      cli.KV,
		Lease:   cli.Lease,
		Client:  cli,
	}
}

type etcKeystore struct {
	Watcher clientv3.Watcher
	KV      clientv3.KV
	Lease   clientv3.Lease
	Client  *clientv3.Client
}

func (client *etcKeystore) NewMutex(lockKey string) (store.Mutex, error) {
	session, err := concurrency.NewSession(client.Client)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("/locks/%s", lockKey)
	return concurrency.NewMutex(session, key), nil
}

func handleError(err error) {
	if err != nil {
		switch err {
		case context.Canceled:
			log.Fatalf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Fatalf("client-side error: %v", err)
		default:
			log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
		}
	}
}

func (client *etcKeystore) Close() error {
	return client.Client.Close()
}

func (client *etcKeystore) runTransaction(ifOps []clientv3.Cmp, thenOps []clientv3.Op) (int64, error) {
	response, err := client.Client.Txn(context.Background()).If(ifOps...).Then(thenOps...).Commit()
	handleError(err)

	if !response.Succeeded {
		log.Println(ifOps)
		return 0, fmt.Errorf("transaction failed, as condition not met")
	}
	return response.Header.Revision, nil
}

func (client *etcKeystore) Create(key string, value []byte) (int64, error) {
	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op
	ifOps = append(ifOps, clientv3util.KeyMissing(key))
	thenOps = append(thenOps, clientv3.OpPut(key, string(value)))
	revision, err := client.runTransaction(ifOps, thenOps)
	if err != nil {
		return 0, fmt.Errorf("unable to create key: %s due to: %s", key, err)
	}
	return revision, nil
}

func (client *etcKeystore) Update(key string, value []byte, modRevision int64) (int64, error) {

	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op

	if modRevision > 0 {
		ifOps = append(ifOps, clientv3util.KeyExists(key))
		checkModRev := clientv3.Compare(clientv3.ModRevision(key), "=", modRevision)
		ifOps = append(ifOps, checkModRev)
	}
	thenOps = append(thenOps, clientv3.OpPut(key, string(value)))

	newRevision, err := client.runTransaction(ifOps, thenOps)
	if err != nil {
		return 0, fmt.Errorf("unable to update ke: %s", err)
	}
	return newRevision, nil
}

func (client *etcKeystore) Delete(key string, modRevision int64) error {
	var ifOps []clientv3.Cmp
	var thenOps []clientv3.Op

	ifOps = append(ifOps, clientv3util.KeyExists(key))
	if modRevision > 0 {
		checkModRev := clientv3.Compare(clientv3.ModRevision(key), "=", modRevision)
		ifOps = append(ifOps, checkModRev)
	}
	thenOps = append(thenOps, clientv3.OpDelete(key))

	_, err := client.runTransaction(ifOps, thenOps)
	return err
}

func getKeyValueVersion(rawKeyValue *mvccpb.KeyValue) *store.KeyValueVersion {
	if rawKeyValue == nil {
		return nil
	}
	return &store.KeyValueVersion{
		Key:            string(rawKeyValue.Key),
		Value:          rawKeyValue.Value,
		ModRevision:    rawKeyValue.ModRevision,
		CreateRevision: rawKeyValue.CreateRevision,
	}
}

func (client *etcKeystore) IsExist(key string) (bool, error) {
	response, err := client.Client.Get(context.Background(), key)
	handleError(err)
	return response.Count == 1, nil
}

func (client *etcKeystore) GetAll(prefix string) ([]store.KeyValueVersion, error) {
	response, err := client.Client.Get(context.Background(), prefix, clientv3.WithPrefix())
	handleError(err)

	var values []store.KeyValueVersion
	for _, rawKeyValue := range response.Kvs {
		values = append(values, *getKeyValueVersion(rawKeyValue))
	}
	return values, nil
}

func (client *etcKeystore) Get(key string) (store.KeyValueVersion, error) {
	response, err := client.Client.Get(context.Background(), key)
	handleError(err)

	value := store.KeyValueVersion{}

	if response.Count == 0 {
		return value, fmt.Errorf("unable to find any values for key: %s", key)
	}
	if response.Count > 1 {
		panic(errors.New("should never get more than one value for get"))
	}

	return *getKeyValueVersion(response.Kvs[0]), nil
}

func (client *etcKeystore) KeepAliveKey(ctxt context.Context, key string) error {

	getResponse, err := client.Client.Get(context.Background(), key)
	if getResponse.Count == 1 {
		// if another host seems to exist, back off for 10 seconds incase we just did a quick restart
		time.Sleep(time.Second * 10)
	}

	// TODO what about configure timeout and ttl?
	var ttl int64 = 10
	grantResponse, err := client.Client.Grant(ctxt, ttl)
	if err != nil {
		log.Fatal(err)
	}
	leaseID := grantResponse.ID

	txnResponse, err := client.Client.Txn(ctxt).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, "keep-alive", clientv3.WithLease(leaseID), clientv3.WithPrevKV())).
		Commit()
	handleError(err)
	if !txnResponse.Succeeded {
		return fmt.Errorf("unable to create keep-alive key %s due to: %+v", key, txnResponse.Responses)
	}

	ch, err := client.Client.KeepAlive(ctxt, leaseID)
	if err != nil {
		log.Fatal(err)
	}

	counter := 9
	go func() {
		for range ch {
			if counter >= 9 {
				counter = 0
				log.Println("Still refreshing key:", key)
			} else {
				counter++
			}
		}
		// TODO: should allow context to be cancelled
		log.Panicf("Unable to refresh key: %s", key)
	}()
	return nil
}

func (client *etcKeystore) DeleteAllKeysWithPrefix(prefix string) (int64, error) {
	response, err := client.Client.Delete(context.Background(), prefix, clientv3.WithPrefix())
	handleError(err)
	return response.Deleted, nil
}

func (client *etcKeystore) Watch(ctxt context.Context, key string, withPrefix bool) store.KeyValueUpdateChan {
	options := []clientv3.OpOption{clientv3.WithPrevKV()}
	if withPrefix {
		options = append(options, clientv3.WithPrefix())
	}
	rch := client.Watcher.Watch(ctxt, key, options...)

	c := make(chan store.KeyValueUpdate)

	go processWatchEvents(rch, c)

	return c
}

func processWatchEvents(watchChan clientv3.WatchChan, c chan store.KeyValueUpdate) {
	for watchResponse := range watchChan {
		// if error, send empty update with an error
		err := watchResponse.Err()
		if err != nil {
			c <- store.KeyValueUpdate{Err: err}
		}

		// send all events in this watch response
		for _, ev := range watchResponse.Events {
			update := store.KeyValueUpdate{
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
