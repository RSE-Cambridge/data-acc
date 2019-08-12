package pool

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/utils"
)

// Used by implementor of session to access brick_host things
type Registry interface {
	// Get this mutex before calling GetAllPools
	// choosing the required bricks, then calling
	// CreateSession that creates the allocation records
	GetAllocationMutex() (utils.Mutex, error)

	// Get all bricks listed by pools
	GetAllPools() ([]model.PoolInfo, error)

	// Dacd is waiting for new allocations
	// This call blocks until Dacd creates a key to what queue it is waiting for session actions on
	CreateAllocations(name model.SessionName, allocations []model.BrickAllocation)
}
