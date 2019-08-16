package brick_manager_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"log"
)

func NewSessionActionHandler(actions registry.SessionActions) facade.SessionActionHandler {
	return &sessionActionHandler{actions: actions}
}

type sessionActionHandler struct {
	actions      registry.SessionActions
	skipActions  bool
	actionCalled datamodel.SessionActionType
}

func (s *sessionActionHandler) ProcessSessionAction(action datamodel.SessionAction) {
	log.Printf("Started to process: %+v\n", action)
	switch action.ActionType {
	case datamodel.SessionDelete:
		// TODO... must test this better!
		s.actionCalled = datamodel.SessionDelete
		if !s.skipActions {
			go s.handleDelete(action)
		}
	case datamodel.SessionCreate:
		s.actionCalled = datamodel.SessionCreate
		if !s.skipActions {
			go s.handleCreate(action)
		}
	default:
		log.Panicf("not yet implemented action for %+v", action)
	}
}

func (s *sessionActionHandler) handleCreate(action datamodel.SessionAction) {
	log.Println("create")
	s.reportComplete(action)
}

func (s *sessionActionHandler) handleDelete(action datamodel.SessionAction) {
	log.Println("delete")
	s.reportComplete(action)
}

func (s *sessionActionHandler) reportComplete(action datamodel.SessionAction) {
	err := s.actions.CompleteSessionAction(action, nil)
	if err != nil {
		log.Println("Failed to complete ActionType:", err)
		// TODO: put session into error state?
		return
	}
	log.Println("Stopped processing action:", action)
}
