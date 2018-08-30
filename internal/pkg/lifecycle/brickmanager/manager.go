package brickmanager

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider/ansible"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"os"
)

type BrickManager interface {
	Start() error
	Hostname() string
}

func NewBrickManager(poolRegistry registry.PoolRegistry, volumeRegistry registry.VolumeRegistry) BrickManager {
	return &brickManager{poolRegistry, volumeRegistry, getHostname()}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}

type brickManager struct {
	poolRegistry   registry.PoolRegistry
	volumeRegistry registry.VolumeRegistry
	hostname       string
}

func (bm *brickManager) Hostname() string {
	return bm.hostname
}

func (bm *brickManager) Start() error {
	devices := getDevices()
	updateBricks(bm.poolRegistry, bm.hostname, devices)

	// TODO, on startup see what existing allocations there are, and watch those volumes
	setupBrickEventHandlers(bm.poolRegistry, bm.volumeRegistry, bm.hostname)

	// Do this after registering all bricks and their handlers, and any required tidy up
	notifyStarted(bm.poolRegistry, bm.hostname)

	// Check after the processes have started up
	outputDebugLogs(bm.poolRegistry, bm.hostname)

	return nil
}

const FakeDeviceAddress = "nvme%dn1"
const FakeDeviceCapacityGB = 1600
const FakePoolName = "default"

func getDevices() []string {
	// TODO: check for real devices!
	//devices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	devices := []int{1, 2, 3, 4}
	if FSType == ansible.BeegFS {
		devices = []int{0, 1, 2, 3, 4}
	}
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
	err := poolRegistry.KeepAliveHost(hostname)
	if err != nil {
		log.Fatalln(err)
	}
}
