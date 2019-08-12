package session

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/utils"
)

type Registry interface {
	// Gets a session and its current allocations
	// Returns an error if the session is not found
	GetSession(token string) (model.Session, error)

	// Any required allocations are created for the given session
	// such that actions can now be sent to the given session
	// Returns an error if the session already exists
	// Note that deleting a session and its allocation is an action, as is any update
	CreateSession(s model.Session) (model.Session, error)

	// Checks it would be a valid call to CreateAllocations
	// Error will describe any validation issues
	ValidateSessionRequest(token string) (model.Session, error)

	// Used for show instances and show sessions
	GetAllSessions() ([]model.Session, error)

	// Get all bricks listed by pools
	GetAllPools() ([]model.PoolInfo, error)

	// Get this mutex before calling GetAllPools
	// choosing the required bricks, then calling
	// CreateSession that creates the allocation records
	GetAllocationMutex() (utils.Mutex, error)
}
