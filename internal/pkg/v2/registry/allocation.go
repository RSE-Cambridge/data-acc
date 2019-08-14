package registry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
)

type AllocationRegistry interface {
	// Get brick availability by pool
	GetBricksByPool() ([]datamodel.PoolInfo, error)

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
