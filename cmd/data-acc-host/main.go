package main

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/coreos/etcd/clientv3"
	"log"
	"math/rand"
	"os"
	"strings"
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

func getAvailableBricks(cli *clientv3.Client) []registry.Brick {
	var availableBricks []registry.Brick
	getResponse, err := cli.Get(context.Background(), "/slices/present", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, keyValue := range getResponse.Kvs {
		rawKey := fmt.Sprintf("%s", keyValue.Key)
		key := strings.Split(rawKey, "/")
		brick := registry.Brick{Name: key[3], Hostname: key[2]}
		availableBricks = append(availableBricks, brick)
	}
	// TODO exclude inuse bricks...
	return availableBricks
}

func addFakeBufferAndSlices(keystore keystoreregistry.Keystore, cli *clientv3.Client) {
	log.Println("Add fakebuffer and match to slices")
	bufferRegistry := keystoreregistry.NewBufferRegistry(keystore)
	availableBricks := getAvailableBricks(cli)

	log.Println("all bricks:")
	log.Println(availableBricks)
	var chosenBricks []registry.Brick

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	requestedBricks := 3
	for retries := 0; len(chosenBricks) < requestedBricks && retries <= (requestedBricks*30); retries++ {
		candidateBrick := availableBricks[r.Intn(len(availableBricks))]
		goodCandidate := true
		for _, brick := range chosenBricks {
			if brick == candidateBrick {
				goodCandidate = false
				break
			}
			if brick.Hostname == candidateBrick.Hostname {
				goodCandidate = false
				break
			}
			// TODO: check host is alive, etc, etc,...
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}
	}
	// TODO: check we have enough bricks?

	bufferName, _ := os.Hostname()
	log.Printf("For buffer %s selected following bricks: %s\n", bufferName, chosenBricks)

	// TODO: should be done in a single transaction, and retry if clash
	for i, brick := range chosenBricks {
		chosenKey := fmt.Sprintf("/slices/inuse/%s/%s", brick.Hostname, brick.Name)
		keystore.AtomicAdd(chosenKey, fmt.Sprintf("%s:%d", bufferName, i))
	}

	buffer := registry.Buffer{Name: bufferName, Bricks: chosenBricks}
	bufferRegistry.AddBuffer(buffer)
	defer bufferRegistry.RemoveBuffer(buffer)
}

func main() {
	fmt.Println("Hello from data-acc-host.")

	cli := etcdregistry.NewEtcdClient()
	keystore := etcdregistry.EtcKeystore{Client: cli}
	defer keystore.Close()

	hostname := getHostname()

	baseSliceKey := fmt.Sprintf("/slices/present/%s", hostname)
	go keystore.WatchPutPrefix(baseSliceKey, func(key string, value string) {
		log.Printf("Added slice: %s with value: %s\n", key, value)
	})

	baseSliceInUse := fmt.Sprintf("/slices/inuse/%s", hostname)
	go keystore.WatchPutPrefix(baseSliceInUse, func(key string, value string) {
		log.Printf("Added in use slice: %s with index: %s\n", key, value)
	})

	// TODO: should really just check if existing key needs an update
	cli.Delete(context.Background(), baseSliceKey, clientv3.WithPrefix())
	defer keystore.CleanPrefix(baseSliceKey)
	slices := getSlices(baseSliceKey)
	for _, sliceKey := range slices {
		keystore.AtomicAdd(sliceKey, FAKE_DEVICE_INFO)
	}

	// TODO: nasty testing hack
	cli.Delete(context.Background(), baseSliceInUse, clientv3.WithPrefix())
	defer keystore.CleanPrefix(baseSliceInUse)

	keepaliveKey := fmt.Sprintf("/bufferhost/alive/%s", hostname)
	log.Printf("Adding keepalive key: %s \n", keepaliveKey)

	addFakeBufferAndSlices(&keystore, cli)

	ch, err := startKeepAlive(cli, keepaliveKey)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ka := <-ch
		log.Println("Refreshed key. Current ttl:", ka.TTL)
	}
}
