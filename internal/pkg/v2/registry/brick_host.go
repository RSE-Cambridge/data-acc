package registry

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
)

type BrickHostRegistry interface {
	// BrickHost updates bricks on startup
	// This will error if we remove a brick that has an allocation
	// for a Session that isn't in an error state
	// This includes ensuring the pool exists and is consistent with the given brick host info
	UpdateBrickHost(brickHostInfo datamodel.BrickHost) error

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
