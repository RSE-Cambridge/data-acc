package brick_manager_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/workflow"
	"log"
)

func NewSessionActionHandler(actions registry.SessionActions) workflow.SessionActionHandler {
	return &sessionActionHandler{actions: actions}
}

type sessionActionHandler struct {
	actions registry.SessionActions
}

func (s *sessionActionHandler) ProcessSessionAction(action datamodel.SessionAction) {
	log.Println("Started to process:", action)
	err := s.actions.CompleteSessionAction(action, nil)
	if err != nil {
		log.Println("Failed to complete Action:", err)
		return
	}
	log.Println("Stopped processing action:", action)
}
