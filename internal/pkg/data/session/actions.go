package session

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

// Each volume has an associated primary brick
// that is responsible for responding to any actions
// All the calls block until they are complete, or an error occurs
type Actions interface {
	// Creates the requested volumes
	// Error if Session has not had its bricks allocated
	CreateSessionVolume(session datamodel.Session) error

	// Deletes the requested volume and session allocation
	DeleteSession(session datamodel.Session) error

	// Update the session and trigger requested data copy in
	DataIn(session datamodel.Session) error

	// Update session hosts and attach volumes as needed
	AttachVolumes(session datamodel.Session) error

	// Attempt to detach volumes
	DetachVolumes(session datamodel.Session) error

	// Update the session and trigger requested data copy out
	DataOut(session datamodel.Session) error
}