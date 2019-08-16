package facade

import "github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"

// Each volume has an associated primary brick
// that is responsible for responding to any actions
// All the calls block until they are complete, or an error occurs
type Session interface {
	// Allocates storage and
	CreateSession(session datamodel.Session) error

	// Deletes the requested volume and session allocation
	// If hurry, there is no stage-out attempted
	// Unmount is always attempted before deleting the buffer
	DeleteSession(sessionName datamodel.SessionName, hurry bool) error

	// Update the session and trigger requested data copy in
	CopyDataIn(sessionName datamodel.SessionName) error

	// Update session hosts and attach volumes as needed
	Mount(sessionName datamodel.SessionName, computeNodes []string, loginNodes []string) error

	// Attempt to detach volumes
	Unmount(sessionName datamodel.SessionName) error

	// Update the session and trigger requested data copy out
	CopyDataOut(sessionName datamodel.SessionName) error

	// Get brick availability by pool
	GetPools() ([]datamodel.PoolInfo, error)

	// Get requested session
	//
	// Error if session does not exist
	GetSession(sessionName datamodel.SessionName) (datamodel.Session, error)

	// Get all sessions
	GetAllSessions() ([]datamodel.Session, error)
}
