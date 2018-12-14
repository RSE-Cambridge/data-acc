package lifecycle

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"math/rand"
	"time"
)

type VolumeLifecycleManager interface {
	ProvisionBricks(pool registry.Pool) error
	DataIn() error
	Mount(hosts []string, jobName string) error
	Unmount(hosts []string, jobName string) error
	DataOut() error
	Delete() error // TODO allow context for timeout and cancel?
}

func NewVolumeLifecycleManager(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	volume registry.Volume) VolumeLifecycleManager {
	return &volumeLifecycleManager{volumeRegistry, poolRegistry, volume}
}

type volumeLifecycleManager struct {
	volumeRegistry registry.VolumeRegistry
	poolRegistry   registry.PoolRegistry
	volume         registry.Volume
}

func (vlm *volumeLifecycleManager) ProvisionBricks(pool registry.Pool) error {
	err := getBricksForBuffer(vlm.poolRegistry, pool, vlm.volume)
	if err != nil {
		return err
	}

	// if there are no bricks requested, don't wait for a provision that will never happen
	if vlm.volume.SizeBricks != 0 {
		err = vlm.volumeRegistry.WaitForState(vlm.volume.Name, registry.BricksProvisioned)
	}
	return err
}

// TODO: delete me?
func getBricksForBufferOld(poolRegistry registry.PoolRegistry,
	pool registry.Pool, volume registry.Volume) error {

	if volume.SizeBricks == 0 {
		// No bricks requested, so return right away
		return nil
	}

	availableBricks := pool.AvailableBricks
	availableBricksByHost := make(map[string][]registry.BrickInfo)
	for _, brick := range availableBricks {
		hostBricks := availableBricksByHost[brick.Hostname]
		availableBricksByHost[brick.Hostname] = append(hostBricks, brick)
	}

	var chosenBricks []registry.BrickInfo

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	var hosts []string
	for key := range availableBricksByHost {
		hosts = append(hosts, key)
	}

	randomWalk := rand.Perm(len(availableBricksByHost))
	for _, i := range randomWalk {
		hostBricks := availableBricksByHost[hosts[i]]
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
		if uint(len(chosenBricks)) >= volume.SizeBricks {
			break
		}
	}

	if uint(len(chosenBricks)) != volume.SizeBricks {
		return fmt.Errorf("unable to get number of requested bricks (%d) for given pool (%s)",
			volume.SizeBricks, pool.Name)
	}

	var allocations []registry.BrickAllocation
	for _, brick := range chosenBricks {
		allocations = append(allocations, registry.BrickAllocation{
			Device:              brick.Device,
			Hostname:            brick.Hostname,
			AllocatedVolume:     volume.Name,
			DeallocateRequested: false,
		})
	}
	err := poolRegistry.AllocateBricks(allocations)
	if err != nil {
		return err
	}
	_, err = poolRegistry.GetAllocationsForVolume(volume.Name) // TODO return result, wait for updates
	return err
}

func getBricksForBuffer(poolRegistry registry.PoolRegistry,
	pool registry.Pool, volume registry.Volume) error {

	if volume.SizeBricks == 0 {
		// No bricks requested, so return right away
		return nil
	}

	availableBricks := pool.AvailableBricks
	var chosenBricks []registry.BrickInfo

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	randomWalk := r.Perm(len(availableBricks))
	for _, i := range randomWalk {
		candidateBrick := availableBricks[i]

		// TODO: should not the random walk mean this isn't needed!
		goodCandidate := true
		for _, brick := range chosenBricks {
			if brick == candidateBrick {
				goodCandidate = false
				break
			}
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}
		if uint(len(chosenBricks)) >= volume.SizeBricks {
			break
		}
	}

	if uint(len(chosenBricks)) != volume.SizeBricks {
		return fmt.Errorf("unable to get number of requested bricks (%d) for given pool (%s)",
			volume.SizeBricks, pool.Name)
	}

	var allocations []registry.BrickAllocation
	for _, brick := range chosenBricks {
		allocations = append(allocations, registry.BrickAllocation{
			Device:              brick.Device,
			Hostname:            brick.Hostname,
			AllocatedVolume:     volume.Name,
			DeallocateRequested: false,
		})
	}
	err := poolRegistry.AllocateBricks(allocations)
	if err != nil {
		return err
	}
	_, err = poolRegistry.GetAllocationsForVolume(volume.Name) // TODO return result, wait for updates
	return err
}

func (vlm *volumeLifecycleManager) Delete() error {
	// TODO convert errors into volume related errors, somewhere?
	log.Println("Deleting volume:", vlm.volume.Name, vlm.volume)

	if vlm.volume.SizeBricks == 0 {
		log.Println("No bricks to delete, skipping request delete bricks for:", vlm.volume.Name)

	} else {
		log.Printf("Requested delete of %d bricks for %s", vlm.volume.SizeBricks, vlm.volume.Name)
		err := vlm.volumeRegistry.UpdateState(vlm.volume.Name, registry.DeleteRequested)
		if err != nil {
			return err
		}
		err = vlm.volumeRegistry.WaitForState(vlm.volume.Name, registry.BricksDeleted)
		if err != nil {
			return err
		}
		log.Println("Bricks deleted by brick manager for:", vlm.volume.Name)

		// TODO should we error out here when one of these steps fail?
		err = vlm.poolRegistry.DeallocateBricks(vlm.volume.Name)
		if err != nil {
			return err
		}
		allocations, err := vlm.poolRegistry.GetAllocationsForVolume(vlm.volume.Name)
		if err != nil {
			return err
		}
		// TODO we should really wait for the brick manager to call this API
		err = vlm.poolRegistry.HardDeleteAllocations(allocations)
		if err != nil {
			return err
		}
		log.Println("Allocations all deleted, count:", len(allocations))
	}

	log.Println("Deleting volume record in registry for:", vlm.volume.Name)
	return vlm.volumeRegistry.DeleteVolume(vlm.volume.Name)
}

func (vlm *volumeLifecycleManager) DataIn() error {
	if vlm.volume.SizeBricks == 0 {
		log.Println("skipping datain for:", vlm.volume.Name)
		return nil
	}

	err := vlm.volumeRegistry.UpdateState(vlm.volume.Name, registry.DataInRequested)
	if err != nil {
		return err
	}
	return vlm.volumeRegistry.WaitForState(vlm.volume.Name, registry.DataInComplete)
}

func (vlm *volumeLifecycleManager) Mount(hosts []string, jobName string) error {
	if vlm.volume.SizeBricks == 0 {
		log.Println("skipping mount for:", vlm.volume.Name) // TODO: should never happen now?
		return nil
	}

	if vlm.volume.State != registry.BricksProvisioned && vlm.volume.State != registry.DataInComplete {
		return fmt.Errorf("unable to mount volume: %s in state: %s", vlm.volume.Name, vlm.volume.State)
	}

	var attachments []registry.Attachment
	for _, host := range hosts {
		attachments = append(attachments, registry.Attachment{
			Hostname: host, State: registry.RequestAttach, Job: jobName,
		})
	}

	if err := vlm.volumeRegistry.UpdateVolumeAttachments(vlm.volume.Name, attachments); err != nil {
		return err
	}

	// TODO: should share code with Unmount!!
	var volumeInErrorState bool
	err := vlm.volumeRegistry.WaitForCondition(vlm.volume.Name, func(old *registry.Volume, new *registry.Volume) bool {
		if new.State == registry.Error {
			volumeInErrorState = true
			return true
		}
		allAttached := false
		for _, host := range hosts {

			var isAttached bool
			for _, attachment := range new.Attachments {
				if attachment.Job == jobName && attachment.Hostname == host {
					if attachment.State == registry.Attached {
						isAttached = true
					} else if attachment.State == registry.AttachmentError {
						// found an error bail out early
						volumeInErrorState = true
						return true // Return true to stop the waiting
					} else {
						isAttached = false
					}
					break
				}
			}

			if isAttached {
				allAttached = true
			} else {
				allAttached = false
				break
			}
		}
		return allAttached
	})
	if volumeInErrorState {
		return fmt.Errorf("unable to mount volume: %s", vlm.volume.Name)
	}
	return err
}

func (vlm *volumeLifecycleManager) Unmount(hosts []string, jobName string) error {
	if vlm.volume.SizeBricks == 0 {
		log.Println("skipping postrun for:", vlm.volume.Name) // TODO return error type and handle outside?
		return nil
	}

	if vlm.volume.State != registry.BricksProvisioned && vlm.volume.State != registry.DataInComplete {
		return fmt.Errorf("unable to unmount volume: %s in state: %s", vlm.volume.Name, vlm.volume.State)
	}

	var updates []registry.Attachment
	for _, host := range hosts {
		attachment, ok := vlm.volume.FindAttachment(host, jobName)
		if !ok {
			return fmt.Errorf(
				"can't find attachment for volume: %s host: %s job: %s",
				vlm.volume, host, jobName)
		}

		if attachment.State != registry.Attached {
			return fmt.Errorf("attachment must be attached to do unmount for volume: %s", vlm.volume.Name)
		}
		attachment.State = registry.RequestDetach
		updates = append(updates, *attachment)
	}
	if err := vlm.volumeRegistry.UpdateVolumeAttachments(vlm.volume.Name, updates); err != nil {
		return err
	}

	// TODO: must share way more code and do more tests on this logic!!
	volumeInErrorState := false
	err := vlm.volumeRegistry.WaitForCondition(vlm.volume.Name, func(old *registry.Volume, new *registry.Volume) bool {
		if new.State == registry.Error {
			volumeInErrorState = true
			return true
		}
		allDettached := false
		for _, host := range hosts {
			newAttachment, ok := new.FindAttachment(host, jobName)
			if !ok {
				// TODO: debug log or something?
				volumeInErrorState = true
				return true
			}

			if newAttachment.State == registry.AttachmentError {
				// found an error bail out early
				volumeInErrorState = true
				return true
			}

			if newAttachment.State == registry.Detached {
				allDettached = true
			} else {
				allDettached = false
				break
			}
		}
		return allDettached
	})
	if volumeInErrorState {
		return fmt.Errorf("unable to mount volume: %s", vlm.volume.Name)
	}
	if err != nil {
		return err
	}
	return vlm.volumeRegistry.DeleteVolumeAttachments(vlm.volume.Name, hosts, jobName)
}

func (vlm *volumeLifecycleManager) DataOut() error {
	if vlm.volume.SizeBricks == 0 {
		log.Println("skipping data_out for:", vlm.volume.Name)
		return nil
	}

	err := vlm.volumeRegistry.UpdateState(vlm.volume.Name, registry.DataOutRequested)
	if err != nil {
		return err
	}
	return vlm.volumeRegistry.WaitForState(vlm.volume.Name, registry.DataOutComplete)
}
