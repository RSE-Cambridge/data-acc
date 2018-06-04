package main

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/coreos/etcd/clientv3"
	"log"
	"os"
	"time"
)

const FAKE_DEVICE_ADDRESS = "nvme%dn1"
const FAKE_DEVICE_INFO = "TODO"

func getHostname() string {
	hostname, error := os.Hostname()
	if error != nil {
		log.Fatal(error)
	}
	return hostname
}

func getDevices(baseBrickKey string) []string {

	devices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var bricks []string
	for _, i := range devices {
		device := fmt.Sprintf(FAKE_DEVICE_ADDRESS, i)
		bricks = append(bricks, fmt.Sprintf("%s/%s", baseBrickKey, device))
	}
	return bricks
}

func startKeepAlive(cli *clientv3.Client, keepaliveKey string) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	grantResponse, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
	}

	leaseID := grantResponse.ID

	_, err = cli.Put(context.TODO(), keepaliveKey, "", clientv3.WithLease(leaseID))
	if err != nil {
		log.Fatal(err)
	}

	return cli.KeepAlive(context.TODO(), leaseID)
}

func main() {
	fmt.Println("Hello from data-acc-host.")

	cli := etcdregistry.NewEtcdClient()
	keystore := etcdregistry.EtcKeystore{Client: cli}
	defer keystore.Close()

	hostname := getHostname()

	baseBrickKey := fmt.Sprintf("/bricks/present/%s", hostname)
	go keystore.WatchPutPrefix(baseBrickKey, func(key string, value string) {
		log.Printf("Added brick: %s with value: %s\n", key, value)
	})

	baseBrickInUse := fmt.Sprintf("/bricks/inuse/%s", hostname)
	go keystore.WatchPutPrefix(baseBrickInUse, func(key string, value string) {
		log.Printf("Added in use brick: %s for: %s\n", key, value)
	})

	// TODO: should really just check if existing key needs an update
	cli.Delete(context.Background(), baseBrickKey, clientv3.WithPrefix())
	defer keystore.CleanPrefix(baseBrickKey)
	bricks := getDevices(baseBrickKey)
	for _, brickKey := range bricks {
		keystore.AtomicAdd(brickKey, FAKE_DEVICE_INFO)
	}

	// TODO: nasty testing hack
	cli.Delete(context.Background(), baseBrickInUse, clientv3.WithPrefix())
	defer keystore.CleanPrefix(baseBrickInUse)

	keepaliveKey := fmt.Sprintf("/bufferhost/alive/%s", hostname)
	log.Printf("Adding keepalive key: %s \n", keepaliveKey)
	ch, err := startKeepAlive(cli, keepaliveKey)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: hack, lets wait a bit for others to start
	time.Sleep(2)
	buffer := fakewarp.AddFakeBufferAndBricks(&keystore, cli)
	bufferRegistry := keystoreregistry.NewBufferRegistry(&keystore)
	defer bufferRegistry.RemoveBuffer(buffer) // TODO remove in-use brick entries

	for {
		ka := <-ch
		log.Println("Refreshed key. Current ttl:", ka.TTL)
	}
}
