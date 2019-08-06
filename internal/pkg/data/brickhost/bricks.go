package brickhost

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

type BrickRegistry interface {
	// Returns a summary of the current state of all pools, including the bricks in each pool
	Pools() ([]datamodel.Pool, error)

	// BrickHosts need to check they match a pool
	// which may involve creating the default pool
	EnsureDefaultPoolCreated(granularityGB int) (datamodel.PoolName, error)

	// BrickHost updates bricks on startup
	UpdateBrickHost(brickHostInfo datamodel.BrickHostInfo) error

	// Get information on the BrickHost
	GetBrickHostInfo(name datamodel.BrickHostName) (datamodel.BrickHostInfo, error)

	// While the process is still running this notifies others the host is up
	//
	// When a host is dead non of its bricks will get new volumes assigned,
	// and no bricks will get cleaned up until the next service start.
	// Error will be returned if the host info has not yet been written.
	KeepAliveHost(hostname string) error

	// Update a brick with allocation information.
	//
	// No update is made and an error is returned if:
	// any brick already has an allocation,
	// or any volume a brick is being assigned to already has an allocation,
	// or if any of the volumes do not exist
	// or if there is not exactly one primary brick.
	//
	// Note: you may assign multiple volumes in a single call, but all bricks
	// for a particular volume must be set in a single call
	AllocateBricksForVolume(volume Volume) ([]BrickAllocation, error)

	// Deallocate all bricks associated with the given volume
	//
	// No update is made and an error is returned if any of brick allocations don't match the current state.
	// If any host associated with one of the bricks is down, an error is returned and the deallocate is
	// recorded as requested and not executed.
	// Note: this returns as soon as deallocate is requested, doesn't wait for cleanup completion
	DeallocateBricks(volume VolumeName) error

	// This is called after DeallocateBricks has been processed
	HardDeleteAllocations(allocations []BrickAllocation) error

	// Get all the allocations for bricks associated with the specified hostname
	GetAllocationsForHost(hostname string) ([]BrickAllocation, error)

	// Get all the allocations for bricks associated with the specific volume
	GetAllocationsForVolume(volume VolumeName) ([]BrickAllocation, error)

	// Get information on a specific brick
	GetBrickInfo(hostname string, device string) (BrickInfo, error)

	// Returns a channel that reports all new brick allocations for given hostname
	//
	// The channel is closed when the context is cancelled or timeout.
	// Any errors in the watching log the issue and panic
	GetNewHostBrickAllocations(ctxt context.Context, hostname string) <-chan BrickAllocation
}
