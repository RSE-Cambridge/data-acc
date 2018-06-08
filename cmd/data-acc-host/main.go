package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const FakeDeviceAddress = "nvme%dn1"

func getHostname() string {
	hostname, error := os.Hostname()
	if error != nil {
		log.Fatal(error)
	}
	return hostname
}

func getDevices() []string {
	// TODO: check for real devices!
	devices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var bricks []string
	for _, i := range devices {
		device := fmt.Sprintf(FakeDeviceAddress, i)
		bricks = append(bricks, device)
	}
	return bricks
}

func main() {
	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()

	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)

	hostname := getHostname()
	devices := getDevices()

	var bricks []registry.BrickInfo
	for _, device := range devices {
		bricks = append(bricks, registry.BrickInfo{
			Device:     device,
			Hostname:   hostname,
			CapacityGB: 42,
			PoolName:   "DefaultPool",
		})
	}
	err := poolRegistry.UpdateHost(bricks)
	if err != nil {
		log.Fatalln(err)
	}

	poolRegistry.WatchHostBrickAllocations(hostname,
		func(old *registry.BrickAllocation, new *registry.BrickAllocation) {
			log.Println("Noticed brick allocation update. Old:", old, "New:", new)
		})

	allocations, err := poolRegistry.GetAllocationsForHost(hostname)
	if err != nil {
		// Ignore errors, we may not have any results when there are no allocations
		// TODO: maybe stop returing an error for the empty case?
		log.Println(err)
	}
	log.Println("Current allocations:", allocations)

	pools, err := poolRegistry.Pools()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Current pools:", pools)

	// TODO: if we restart quickly this fails as key is already present, maybe don't check that key doesn't exist?
	time.Sleep(time.Second * 10)
	log.Println("Started, now notify others, for:", hostname)
	err = poolRegistry.KeepAliveHost(hostname)
	if err != nil {
		log.Fatalln(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	log.Println("I have been asked to shutdown, doing tidy up...")
	os.Exit(1)
}
