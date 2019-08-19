package registry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
)

type AllocationRegistry interface {
	// Get all registered pools
	GetPool(name datamodel.PoolName) (datamodel.Pool, error)

	// Creates the pool if it doesn't exist
	// error if the granularity doesn't match and existing pool
	// TODO package method called by brick registry?
	//EnsurePoolCreated(poolName datamodel.PoolName, granularityGB int) (datamodel.Pool, error)

	// Get brick availability by pool
	GetAllPoolInfos() ([]datamodel.PoolInfo, error)

	// Get brick availability for one pool
	// bricks are only available if corresponding host currently alive
	GetPoolInfo(poolName datamodel.PoolName) (datamodel.PoolInfo, error)

	// Caller should acquire this mutex before calling GetAllPools then CreateAllocations
	GetAllocationMutex() (store.Mutex, error)

	// Allocations written (by the client), while holding above mutex
	//
	// Error if any bricks already have an allocation
	CreateAllocations(sessionName datamodel.SessionName, allocations []datamodel.Brick) ([]datamodel.BrickAllocation, error)

	// Allocations deleted by server when bricks no being used
	//
	// Does not error if given allocation has already been dropped
	DeleteAllocations(allocations []datamodel.BrickAllocation) error
}
