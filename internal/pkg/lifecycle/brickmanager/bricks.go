package brickmanager

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider/fake"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

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
					log.Printf("Requested clean of brick: %d:%s", new.AllocatedIndex, new.Device)
				}
			}
		})
}

func processNewPrimaryBlock(volumeRegistry registry.VolumeRegistry, new *registry.BrickAllocation) {
	volume, err := volumeRegistry.Volume(new.AllocatedVolume)
	if err != nil {
		log.Printf("Could not file volume: %s because: %s\n", new.AllocatedVolume, err)
		return
	}
	log.Println("Found new volume to watch:", volume.Name)
	log.Println(volume)

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
					// Ignore the state changes we triggered
					log.Println(". ingore volume:", volume.Name, "state move:", old.State, "->", new.State)
				}
			}
			// TODO spot data in or data out requested
			// TODO spot attachments added or removed
		}
	})

	// Move to new state, ignored by above watch
	provisionNewVolume(volumeRegistry, volume)
}

func handleError(volumeRegistry registry.VolumeRegistry, volume registry.Volume, err error) {
	if err != nil {
		log.Println("Error provisioning", volume.Name, err)
		err = volumeRegistry.UpdateState(volume.Name, registry.Error) // TODO record an error string?
		if err != nil {
			log.Println("Unable to move volume", volume.Name, "to Error state")
		}
	}
}

// TODO: should not be hardcoded here
var plugin = fake.GetPlugin()

func provisionNewVolume(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	if volume.State != registry.Registered {
		log.Println("Volume in bad initial state:", volume.Name)
		return
	}

	err := plugin.VolumeProvider().SetupVolume(volume)
	if err != nil {
		handleError(volumeRegistry, volume, err)
		return
	}

	err = volumeRegistry.UpdateState(volume.Name, registry.BricksProvisioned)
	handleError(volumeRegistry, volume, err)
}

func processDataIn(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	err := plugin.VolumeProvider().CopyDataIn(volume)
	if err != nil {
		handleError(volumeRegistry, volume, err)
		return
	}

	err = volumeRegistry.UpdateState(volume.Name, registry.DataInComplete)
	handleError(volumeRegistry, volume, err)
}

// TODO: well this doesn't work for jobs that have no new bicks, i.e. just attach to persistent buffers
func processMount(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	err := plugin.Mounter().Mount(volume)
	if err != nil {
		handleError(volumeRegistry, volume, err)
		return
	}

	// TODO update volume.Attachments?

	err = volumeRegistry.UpdateState(volume.Name, registry.MountComplete)
	handleError(volumeRegistry, volume, err)
}

func processUnmount(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	err := plugin.Mounter().Unmount(volume)
	if err != nil {
		handleError(volumeRegistry, volume, err)
	}

	// TODO update volume.Attachments?

	err = volumeRegistry.UpdateState(volume.Name, registry.UnmountComplete)
	handleError(volumeRegistry, volume, err)
}

func processDataOut(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	err := plugin.VolumeProvider().CopyDataOut(volume)
	if err != nil {
		handleError(volumeRegistry, volume, err)
	}

	err = volumeRegistry.UpdateState(volume.Name, registry.DataOutComplete)
	handleError(volumeRegistry, volume, err)
}

func processDelete(volumeRegistry registry.VolumeRegistry, volume registry.Volume) {
	err := plugin.VolumeProvider().TeardownVolume(volume)
	if err != nil {
		handleError(volumeRegistry, volume, err)
	}

	err = volumeRegistry.UpdateState(volume.Name, registry.BricksDeleted)
	handleError(volumeRegistry, volume, err)
}
