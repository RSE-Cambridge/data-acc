package registry

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
)

type SessionActions interface {
	// Updates session, then requests action
	//
	// Error if current revision of session doesn't match
	// Error if context is cancelled or timed-out
	SendSessionAction(
		ctxt context.Context, actionType datamodel.SessionActionType,
		session datamodel.Session) (<-chan datamodel.SessionAction, error)

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
