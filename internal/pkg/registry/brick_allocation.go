package registry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store"
)

type AllocationRegistry interface {
	// Caller should acquire this mutex before calling GetAllPools then CreateAllocations
	GetAllocationMutex() (store.Mutex, error)

	// Get all registered pools
	GetPool(name datamodel.PoolName) (datamodel.Pool, error)

	// Creates the pool if it doesn't exist
	// error if the granularity doesn't match and existing pool
	EnsurePoolCreated(poolName datamodel.PoolName, granularityBytes uint) (datamodel.Pool, error)

	// Get brick availability by pool
	GetAllPoolInfos() ([]datamodel.PoolInfo, error)

	// Get brick availability for one pool
	// bricks are only available if corresponding host currently alive
	GetPoolInfo(poolName datamodel.PoolName) (datamodel.PoolInfo, error)
}
