package workflow

import "github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"

// Each volume has an associated primary brick
// that is responsible for responding to any actions
// All the calls block until they are complete, or an error occurs
type Session interface {
	// Allocates storage and
	CreateSessionVolume(session datamodel.Session) error

	// Deletes the requested volume and session allocation
	DeleteSession(sessionName datamodel.SessionName) error

	// Update the session and trigger requested data copy in
	DataIn(sessionName datamodel.SessionName) error

	// Update session hosts and attach volumes as needed
	AttachVolumes(sessionName datamodel.SessionName, attachHosts []string) error

	// Attempt to detach volumes
	DetachVolumes(sessionName datamodel.SessionName) error

	// Update the session and trigger requested data copy out
	DataOut(sessionName datamodel.SessionName) error

	// Get brick availability by pool
	GetPools() ([]datamodel.PoolInfo, error)

	// Get requested session
	//
	// Error if session does not exist
	GetSession(sessionName datamodel.SessionName) (datamodel.Session, error)

	// Get all sessions
	GetAllSessions() ([]datamodel.Session, error)
}
