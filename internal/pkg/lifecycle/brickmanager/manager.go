package brickmanager

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider/ansible"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"os"
	"strconv"
)

type BrickManager interface {
	Start() error
	Hostname() string
}

func NewBrickManager(poolRegistry registry.PoolRegistry, volumeRegistry registry.VolumeRegistry) BrickManager {
	return &brickManager{poolRegistry, volumeRegistry, getHostname()}
}

func getHostname() string {
	// TODO: make this configurable?
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
	updateBricks(bm.poolRegistry, bm.hostname)

	// TODO, on startup see what existing allocations there are, and watch those volumes
	setupBrickEventHandlers(bm.poolRegistry, bm.volumeRegistry, bm.hostname)

	// Do this after registering all bricks and their handlers, and any required tidy up
	notifyStarted(bm.poolRegistry, bm.hostname)

	// Check after the processes have started up
	outputDebugLogs(bm.poolRegistry, bm.hostname)

	return nil
}

const DefaultDeviceAddress = "nvme%dn1"
const DefaultDeviceCapacityGB = 1400
const DefaultPoolName = "default"

func getDevices(devicesStr string) []string {
	// TODO: check for real devices!
	count, err := strconv.Atoi(devicesStr)
	if err != nil {
		count = 12
	}

	var bricks []string
	for i := 0; i < count; i++ {
		if i == 0 && FSType == ansible.Lustre {
			// TODO: we should use another disk for MGS
			continue
		}
		device := fmt.Sprintf(DefaultDeviceAddress, i)
		bricks = append(bricks, device)
	}
	return bricks
}

func getBricks(devices []string, hostname string, capacityStr string, poolName string) []registry.BrickInfo {
	capacity, ok := strconv.Atoi(capacityStr)
	if ok != nil || capacityStr == "" || capacity <= 0 {
		capacity = DefaultDeviceCapacityGB
	}
	if poolName == "" {
		poolName = DefaultPoolName
	}

	var bricks []registry.BrickInfo
	for _, device := range devices {
		bricks = append(bricks, registry.BrickInfo{
			Device:     device,
			Hostname:   hostname,
			CapacityGB: uint(capacity),
			PoolName:   poolName,
		})
	}
	return bricks
}

func updateBricks(poolRegistry registry.PoolRegistry, hostname string) {
	devicesStr := os.Getenv("DEVICE_COUNT")
	devices := getDevices(devicesStr)

	capacityStr := os.Getenv("DAC_DEVICE_COUNT")
	poolName := os.Getenv("DAC_POOL_NAME")
	bricks := getBricks(devices, hostname, capacityStr, poolName)

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
