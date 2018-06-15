package brickmanager

import (
	"fmt"
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

type brickManager struct {
	poolRegistry   registry.PoolRegistry
	volumeRegistry registry.VolumeRegistry
	hostname       string
}

func (bm *brickManager) Start() error {
	devices := getDevices()
	updateBricks(bm.poolRegistry, bm.hostname, devices)

	// TODO, on startup see what existing allocations there are, and watch those volumes
	setupBrickEventHandlers(bm.poolRegistry, bm.volumeRegistry, bm.hostname)

	notifyStarted(bm.poolRegistry, bm.hostname)

	// Check after the processes have started up
	outputDebugLogs(bm.poolRegistry, bm.hostname)

	return nil
}

func (bm *brickManager) Hostname() string {
	return bm.hostname
}

const FakeDeviceAddress = "nvme%dn1"
const FakeDeviceCapacityGB = 1600
const FakePoolName = "default"

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
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

func handleError(volumeRegistry registry.VolumeRegistry, volume registry.Volume, err error) {
	if err != nil {
		log.Println("Error provisioning", volume.Name, err)
		err = volumeRegistry.UpdateState(volume.Name, registry.Error)
		if err != nil {
			log.Println("Unable to move volume", volume.Name, "to Error state")
		}
	}
}
func provisionNewVolume(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	if volume.State != registry.Registered {
		log.Println("Volume in bad initial state:", volume.Name)
		return
	}

	log.Println("FAKE provision volume:", volume.Name)
	err := volumeRegistry.UpdateState(volume.Name, registry.BricksProvisioned)
	handleError(volumeRegistry, volume, err)
}

func processDataIn(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	log.Println("FAKE datain volume:", volume.Name)
	err := volumeRegistry.UpdateState(volume.Name, registry.DataInComplete)
	handleError(volumeRegistry, volume, err)
}

// TODO: well this doesn't work for jobs that have no new bicks, i.e. just attach to persistent buffers
func processMount(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	log.Println("FAKE mount volume:", volume.Name)
	err := volumeRegistry.UpdateState(volume.Name, registry.MountComplete)
	handleError(volumeRegistry, volume, err)
}

func processUnmount(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	log.Println("FAKE unmount volume:", volume.Name)
	err := volumeRegistry.UpdateState(volume.Name, registry.UnmountComplete)
	handleError(volumeRegistry, volume, err)
}

func processDataOut(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	log.Println("FAKE data_out volume:", volume.Name)
	err := volumeRegistry.UpdateState(volume.Name, registry.DataOutComplete)
	handleError(volumeRegistry, volume, err)
}

func processDelete(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	log.Println("FAKE decomission all bricks for:", volume.Name)
	err := volumeRegistry.UpdateState(volume.Name, registry.BricksDeleted)
	handleError(volumeRegistry, volume, err)
}

func processNewPrimaryBlock(volumeRegistry registry.VolumeRegistry, new *registry.BrickAllocation) {
	volume, err := volumeRegistry.Volume(new.AllocatedVolume)
	if err != nil {
		log.Printf("Could not file volume: %s because: %s\n", new.AllocatedVolume, err)
	}
	log.Println("Found new volume to watch:", volume.Name, "curent state is:", volume.State)

	// TODO: watch from version associated with above volume to avoid any missed events
	// TODO: leaking goroutines here, should cancel the watch when volume is deleted
	volumeRegistry.WatchVolumeChanges(string(volume.Name), func(old *registry.Volume, new *registry.Volume) {
		if old != nil && new != nil {
			if new.State != old.State {
				switch new.State {
				case registry.DataInRequested:
					processDataIn(volumeRegistry, *new)
				case registry.MountRequested:
					processMount(volumeRegistry, *new)
				case registry.UnmountRequested:
					processUnmount(volumeRegistry, *new)
				case registry.DataOutRequested:
					processDataOut(volumeRegistry, *new)
				case registry.DeleteRequested:
					processDelete(volumeRegistry, *new)
				default:
					log.Println("Ingore volume:", volume.Name, "move to state:", volume.State)
				}
			}
		}
	})

	// Move to new state, ignored by above watch
	provisionNewVolume(volumeRegistry, volume)
}

func setupBrickEventHandlers(poolRegistry registry.PoolRegistry, volumeRegistry registry.VolumeRegistry,
	hostname string) {

	poolRegistry.WatchHostBrickAllocations(hostname,
		func(old *registry.BrickAllocation, new *registry.BrickAllocation) {
			// log.Println("Noticed brick allocation update. Old:", old, "New:", new)
			if new.AllocatedVolume != "" && old.AllocatedVolume == "" && new.AllocatedIndex == 0 {
				//log.Println("Dectected we host primary brick for:",
				//	new.AllocatedVolume, "Must check for action.")
				processNewPrimaryBlock(volumeRegistry, new)
			}
			if old.AllocatedVolume != "" {
				if new.DeallocateRequested && !old.DeallocateRequested {
					log.Printf("requested clean of: %d:%s", new.AllocatedIndex, new.Device)
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
	err := poolRegistry.KeepAliveHost(hostname)
	if err != nil {
		log.Fatalln(err)
	}
}
