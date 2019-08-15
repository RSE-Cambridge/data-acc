package workflow_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/workflow"
)

func NewSessionWorkflow() workflow.Session {
	return sessionWorkflow{}
}

type sessionWorkflow struct {
	registry registry.SessionRegistry
	actions  registry.SessionActions
}

func (s sessionWorkflow) CreateSessionVolume(session datamodel.Session) error {
	// TODO needs to get the allocation mutex, create the session, then create the allocations
	//   failing if the pool isn't known, or doesn't have enough space
	panic("implement me")
}

func (s sessionWorkflow) DeleteSession(sessionName datamodel.SessionName, hurry bool) error {
	// TODO get the session mutex, then update the session, then send the event
	//   note the actions registry will error out if the host is not up
	//   release the mutex and wait for the server to complete its work or we timeout
	panic("implement me")
}

func (s sessionWorkflow) DataIn(sessionName datamodel.SessionName) error {
	panic("implement me")
}

func (s sessionWorkflow) AttachVolumes(sessionName datamodel.SessionName, computeNodes []string, loginNodes []string) error {
	panic("implement me")
}

func (s sessionWorkflow) DetachVolumes(sessionName datamodel.SessionName) error {
	panic("implement me")
}

func (s sessionWorkflow) DataOut(sessionName datamodel.SessionName) error {
	panic("implement me")
}

func (s sessionWorkflow) GetPools() ([]datamodel.PoolInfo, error) {
	panic("implement me")
}

func (s sessionWorkflow) GetSession(sessionName datamodel.SessionName) (datamodel.Session, error) {
	panic("implement me")
}

func (s sessionWorkflow) GetAllSessions() ([]datamodel.Session, error) {
	panic("implement me")
}
