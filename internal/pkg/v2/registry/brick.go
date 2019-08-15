package registry

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
)

type BrickRegistry interface {
	// BrickHost updates bricks on startup
	// This will error if we remove a brick that has an allocation
	// for a Session that isn't in an error state
	UpdateBrickHost(brickHostInfo datamodel.BrickHost) error

	// Gets all new actions for the given Session
	// This confirms that dacd is ready to service requests for this session
	// and updates the session with an appropriate SessionActionPrefix
	// The channel tracks all actions sent to that sub keys of the above prefix
	// This is also called when dacd is restarted and you want to resume sending
	// any actions that are not marked as complete
	// New tasks are only sent if the keepalive key is active
	GetSessionActions(ctxt context.Context, brickHostName datamodel.BrickHostName) (<-chan datamodel.SessionAction, error)

	// While the process is still running this notifies others the host is up
	//
	// When a host is dead non of its bricks will get new volumes assigned,
	// and no bricks will get cleaned up until the next service start.
	// Error will be returned if the host info has not yet been written.
	KeepAliveHost(ctxt context.Context, brickHostName datamodel.BrickHostName) error

	// Check if given brick host is alive
	//
	// Error if brick host doesn't exist
	IsBrickHostAlive(brickHostName datamodel.BrickHostName) (bool, error)
}
