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
const FakeDeviceCapacityGB = 1600
const FakePoolName = "default"

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

func updateBricks(poolRegistry registry.PoolRegistry, hostname string, devices []string) {
	var bricks []registry.BrickInfo
	for _, device := range devices {
		bricks = append(bricks, registry.BrickInfo{
			Device:     device,
			Hostname:   hostname,
			CapacityGB: FakeDeviceCapacityGB,
			PoolName:   FakePoolName,
		})
	}
	err := poolRegistry.UpdateHost(bricks)
	if err != nil {
		log.Fatalln(err)
	}
}

func setupBrickEventHandlers(poolRegistry registry.PoolRegistry, hostname string) {
	poolRegistry.WatchHostBrickAllocations(hostname,
		func(old *registry.BrickAllocation, new *registry.BrickAllocation) {
			log.Println("Noticed brick allocation update. Old:", old, "New:", new)
			if new != nil {
				if new.AllocatedIndex == 0 {
					log.Println("Dectected we host primary brick for:",
						new.AllocatedVolume, "Must check for action.")
				}
			}
		})
}

func outputDebugLogs(poolRegistry registry.PoolRegistry, hostname string) {
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
}

func notifyStarted(poolRegistry registry.PoolRegistry, hostname string) {
	// TODO: if we restart quickly this fails as key is already present, maybe don't check that key doesn't exist?
	time.Sleep(time.Second * 10)

	err := poolRegistry.KeepAliveHost(hostname)
	if err != nil {
		log.Fatalln(err)
	}
}

func waitForShutdown() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	log.Println("I have been asked to shutdown, doing tidy up...")
	os.Exit(1)
}

func main() {
	hostname := getHostname()
	log.Println("Starting data-accelerator host service for:", hostname)

	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()
	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)

	devices := getDevices()
	updateBricks(poolRegistry, hostname, devices)

	setupBrickEventHandlers(poolRegistry, hostname)

	log.Println("Notify others we have started:", hostname)
	notifyStarted(poolRegistry, hostname)

	// Check after the processes have started up
	outputDebugLogs(poolRegistry, hostname)

	waitForShutdown()
}
