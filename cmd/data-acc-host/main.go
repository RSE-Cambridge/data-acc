package main

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/coreos/etcd/clientv3"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	// TODO: check for real devices!
	devices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var bricks []string
	for _, i := range devices {
		device := fmt.Sprintf(FAKE_DEVICE_ADDRESS, i)
		bricks = append(bricks, fmt.Sprintf("%s/%s", baseBrickKey, device))
	}
	return bricks
}

func startKeepAlive(cli *clientv3.Client) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	// TODO: move general pattern into Keystore interface, somehow, or just stop keystore registry from being abstract
	hostname := getHostname()
	keepaliveKey := fmt.Sprintf("/bufferhost/alive/%s", hostname)
	log.Println("Adding keepalive key to notify that we have started up: ", keepaliveKey)

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

func addDebugWatches(keystore keystoreregistry.Keystore) {
	hostname := getHostname()

	baseBrickKey := fmt.Sprintf("/bricks/present/%s", hostname)
	go keystore.WatchPutPrefix(baseBrickKey, func(key string, value string) {
		log.Printf("Added brick: %s with value: %s\n", key, value)
	})

	baseBrickInUse := fmt.Sprintf("/bricks/inuse/%s", hostname)
	go keystore.WatchPutPrefix(baseBrickInUse, func(key string, value string) {
		log.Printf("Added in use brick: %s for: %s\n", key, value)
	})
}

func updateDevices(keystore keystoreregistry.Keystore, cli *clientv3.Client) {
	hostname := getHostname()
	baseBrickKey := fmt.Sprintf("/bricks/present/%s", hostname)

	// TODO: should do proper update of exiting entries
	cli.Delete(context.Background(), baseBrickKey, clientv3.WithPrefix()) // don't error if nothing deleted
	bricks := getDevices(baseBrickKey)
	for _, brickKey := range bricks {
		keystore.AtomicAdd(brickKey, FAKE_DEVICE_INFO)
	}
}

func main() {
	cli := etcdregistry.NewEtcdClient()
	keystore := etcdregistry.EtcKeystore{Client: cli} // TODO: fix this hack, shouldn't need cli and keystore
	defer keystore.Close()

	addDebugWatches(&keystore)
	updateDevices(&keystore, cli)

	// TODO: should restore system to expected state, before telling others we are alive

	// Tell system we are ready to configure slices
	ch, err := startKeepAlive(cli)
	if err != nil {
		log.Fatal(err)
	}

	/* time.Sleep(2) // TODO: hack, lets wait a bit for others to start
	buffer := fakewarp.AddFakeBufferAndBricks(&keystore, cli)
	bufferRegistry := keystoreregistry.NewBufferRegistry(&keystore)
	defer bufferRegistry.RemoveBuffer(buffer) // TODO remove in-use brick entries?
	*/

	go func() {
		for {
			ka := <-ch
			log.Println("Refreshed key. Current ttl:", ka.TTL)
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	log.Println("I have been asked to shutdown, doing tidy up...")
	os.Exit(1)
}
