package brick_host

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
)

type BrickHostRegistry interface {
	// Creates the pool if it doesn't exist
	// error if the granularity doesn't match and existing pool
	EnsurePoolCreated(poolName model.PoolName, granularityGB int) (model.PoolName, error)

	// BrickHost updates bricks on startup
	UpdateBrickHost(brickHostInfo model.BrickHostInfo) error

	// While the process is still running this notifies others the host is up
	//
	// When a host is dead non of its bricks will get new volumes assigned,
	// and no bricks will get cleaned up until the next service start.
	// Error will be returned if the host info has not yet been written.
	KeepAliveHost(hostname string) error

	// Get sessions where given hostname is the primary brick
	// Created by dacctl via CreateSessionAllocations
	GetNewSessionRequests(ctxt context.Context, hostname model.BrickHostName) (<-chan model.Session, error)

	// Gets all new actions for the given Session
	// This confirms that dacd is ready to service requests for this session
	// and updates the session with an appropriate SessionActionPrefix
	// The channel tracks all actions sent to that sub keys of the above prefix
	// This is also called when dacd is restarted and you want to resume sending
	// any actions that are not marked as complete
	// New tasks are only sent if the keepalive key is active
	GetNewSessionAction(ctxt context.Context, hostname model.BrickHostName) (<-chan model.SessionAction, error)

	// Update provided session
	// Ensures there are no other updates to Session since the revision contained in it
	UpdateSession(session model.Session) (model.Session, error)

	// This is called before confirming the Session delete request
	// after all bricks have been de-allocated
	DeleteSession(session model.Session) error

	// Mark the given SessionAction as completed
	// Optionally returns an error to caller of the action
	CompleteSessionAction(action model.SessionAction, error error) error

	// Check if a given action is complete
	IsSessionActionComplete(action model.SessionAction) (bool, error)

	// Check if given brick host is alive
	IsBrickHostAlive(hostname model.BrickHostName) (bool, error)
}
