package main

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/coreos/etcd/clientv3"
	"log"
	"os"
)

const FAKE_DEVICE_ADDRESS = "nvme%dn1"
const FAKE_DEVICE_INFO = "TODO"

func getBaseKey() string {
	hostname, error := os.Hostname()
	if error != nil {
		log.Fatal(error)
	}
	return fmt.Sprintf("/bufferhosts/%s", hostname)
}

func getSlices(baseSliceKey string) []string {

	devices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var slices []string
	for _, i := range devices {
		device := fmt.Sprintf(FAKE_DEVICE_ADDRESS, i)
		slices = append(slices, fmt.Sprintf("%s/%s", baseSliceKey, device))
	}
	return slices
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

	baseKey := getBaseKey()
	baseSliceKey := fmt.Sprintf("%s/slices", baseKey)
	slices := getSlices(baseSliceKey)

	// TODO: should really just check if existing key needs an update
	keystore.CleanPrefix(baseSliceKey)
	for _, sliceKey := range slices {
		keystore.AtomicAdd(sliceKey, FAKE_DEVICE_INFO)
	}

	keepaliveKey := fmt.Sprintf("%s/alive", baseSliceKey)
	ch, err := startKeepAlive(cli, keepaliveKey)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ka := <-ch
		log.Println("ttl:", ka.TTL)
	}
}
