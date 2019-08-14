package registry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
)

type SessionRegistry interface {
	// Update provided session
	//
	// Error is session already exists
	CreateSession(session datamodel.Session) (datamodel.Session, error)

	// Get requested session
	//
	// Error if session does not exist
	GetSession(sessionName datamodel.SessionName) (datamodel.Session, error)

	// Get all sessions
	GetAllSessions() ([]datamodel.Session, error)

	// Update provided session
	//
	// Error if current revision does not match (i.e. caller has a stale copy of Session)
	// Error if session does not exist
	UpdateSession(session datamodel.Session) (datamodel.Session, error)

	// This is called before confirming the Session delete request,
	// after all bricks have been de-allocated
	//
	// Error if session has any allocations
	// Error if session doesn't match current revision
	// No error if session has already been deleted
	DeleteSession(session datamodel.Session) error

	// This mutex should be held before doing any operations on given session
	//
	// No error if the session doesn't exist, as this is also used when creating a session
	GetSessionMutex(sessionName datamodel.SessionName) (store.Mutex, error)
}
