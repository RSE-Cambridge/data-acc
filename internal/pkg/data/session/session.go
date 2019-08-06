package session

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

type Registry interface {
	// Gets a session and its current allocations
	// Returns an error if the session is not found
	GetSession(token string) (datamodel.Session, error)

	// Any required allocations are created for the given session
	// such that actions can now be sent to the given session
	// Returns an error if the session already exists
	// Note that deleting a session and its allocation is an action, as is any update
	CreateSessionAllocations(s datamodel.Session) (datamodel.Session, error)

	// Checks it would be a valid call to CreateAllocations
	// Error will describe any validation issues
	ValidateSessionRequest(token string) (datamodel.Session, error)

	// Used for show instances and show sessions
	GetAllSessions() ([]datamodel.Session, error)

	// Get all bricks listed by pools
	GetAllPools() ([]datamodel.Pool, error)
}