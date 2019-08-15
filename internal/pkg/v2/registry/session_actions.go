package registry

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
)

type SessionActions interface {
	// Client requests session volume is created
	//
	// Error if session does not have bricks allocated
	// Error if session volume has already been created
	// Error if primary brick host is not alive or not enabled
	// Error is context is cancelled or timed-out
	CreateSessionVolume(ctxt context.Context, sessionName datamodel.SessionName) (<-chan datamodel.SessionAction, error)

	// Updates session, then requests action
	//
	// Error if current revision of session doesn't match
	// Error if context is cancelled or timed-out
	SendSessionAction(
		ctxt context.Context, actionType datamodel.SessionActionType,
		session datamodel.Session) (<-chan datamodel.Session, error)

	// Get session volume create requests,
	// where given hostname is the primary brick host
	// Called very time brick host is started
	//
	// Error is context is cancelled or timed-out
	GetCreateSessionVolumeRequests(
		ctxt context.Context, brickHostName datamodel.BrickHostName) (<-chan datamodel.SessionAction, error)

	// Gets all new actions for the given Session
	//
	// Error if context is cancelled or timed-out
	GetSessionActions(
		ctxt context.Context, sessionName datamodel.SessionName) (<-chan datamodel.SessionAction, error)

	// Server reports given action is complete
	// Includes callbacks for Create Session Volume
	//
	// Error if action has already completed or doesn't exist
	CompleteSessionAction(action datamodel.SessionAction, err error) error
}
