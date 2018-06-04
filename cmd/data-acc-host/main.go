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

func getBricks(cli *clientv3.Client, prefix string) map[string]map[string]registry.Brick {
	allBricks := make(map[string]map[string]registry.Brick)
	getResponse, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, keyValue := range getResponse.Kvs {
		rawKey := fmt.Sprintf("%s", keyValue.Key) // e.g. /slices/present/1aff0f8468ee/nvme7n1
		key := strings.Split(rawKey, "/")
		brick := registry.Brick{Name: key[4], Hostname: key[3]}
		_, ok := allBricks[brick.Hostname]
		if !ok {
			allBricks[brick.Hostname] = make(map[string]registry.Brick)
		}
		allBricks[brick.Hostname][brick.Name] = brick
	}
	return allBricks
}

func getAvailableBricks(cli *clientv3.Client) map[string][]registry.Brick {
	allBricks := getBricks(cli, "/slices/present/")
	inUseBricks := getBricks(cli, "/slices/inuse/")

	aliveHosts := make(map[string]string)
	getHostsResponse, err := cli.Get(context.Background(), "/bufferhost/alive/", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, keyValue := range getHostsResponse.Kvs {
		rawKey := fmt.Sprintf("%s", keyValue.Key)
		key := strings.Split(rawKey, "/") // e.g. /bufferhost/alive/afe30ea9f27e
		host := key[3]
		aliveHosts[host] = rawKey
	}

	availableBricks := make(map[string][]registry.Brick)

	for host, allHostBricks := range allBricks {
		aliveHost, ok := aliveHosts[host]
		if !ok || aliveHost == "" {
			continue
		}
		inuseHostBricks := inUseBricks[host]

		availableBricks[host] = []registry.Brick{}

		for _, brick := range allHostBricks {
			inuse := false
			for _, inUseBrick := range inuseHostBricks {
				if inUseBrick.Name == brick.Name {
					inuse = true
					break
				}
			}
			if !inuse {
				availableBricks[host] = append(availableBricks[host], brick)
			}
		}
	}
	return availableBricks
}

func addFakeBufferAndSlices(keystore keystoreregistry.Keystore, cli *clientv3.Client) registry.Buffer {
	log.Println("Add fakebuffer and match to slices")
	bufferRegistry := keystoreregistry.NewBufferRegistry(keystore)
	availableBricks := getAvailableBricks(cli)
	var chosenBricks []registry.Brick

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	requestedBricks := 2

	var hosts []string
	for key := range availableBricks {
		hosts = append(hosts, key)
	}

	randomWalk := rand.Perm(len(availableBricks))
	for _, i := range randomWalk {
		hostBricks := availableBricks[hosts[i]]
		candidateBrick := hostBricks[r.Intn(len(hostBricks))]

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
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}

		if len(chosenBricks) >= requestedBricks {
			break
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
	return buffer
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
	ch, err := startKeepAlive(cli, keepaliveKey)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: hack, lets wait a bit for others to start
	time.Sleep(2)
	buffer := addFakeBufferAndSlices(&keystore, cli)
	bufferRegistry := keystoreregistry.NewBufferRegistry(&keystore)
	defer bufferRegistry.RemoveBuffer(buffer) // TODO remove in-use slice entries

	for {
		ka := <-ch
		log.Println("Refreshed key. Current ttl:", ka.TTL)
	}
}
