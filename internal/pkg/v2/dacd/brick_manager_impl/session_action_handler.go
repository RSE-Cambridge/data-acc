package brick_manager_impl

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/filesystem"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"log"
)

func NewSessionActionHandler(actions registry.SessionActions) facade.SessionActionHandler {
	return &sessionActionHandler{actions: actions}
}

type sessionActionHandler struct {
	registry     registry.SessionRegistry
	actions      registry.SessionActions
	fsProvider   filesystem.Provider
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
	sessionName := action.Session.Name
	sessionMutex, err := s.registry.GetSessionMutex(sessionName)
	if err != nil {
		log.Printf("unable to get session mutex: %s due to: %s\n", sessionName, err)
		s.actions.CompleteSessionAction(action, err)
		return
	}
	err = sessionMutex.Lock(context.TODO())
	if err != nil {
		log.Printf("unable to lock session mutex: %s due to: %s\n", sessionName, err)
		s.actions.CompleteSessionAction(action, err)
		return
	}
	defer func() {
		if err := sessionMutex.Unlock(context.TODO()); err != nil {
			log.Println("failed to drop mutex for:", sessionName)
		}
	}()
	log.Printf("starting create for %+v\n", sessionName)

	// Get latest session now we have the mutex
	session, err := s.registry.GetSession(sessionName)

	fsStatus, err := s.fsProvider.Create(action.Session)
	session.FilesystemStatus = fsStatus
	session.Status.FileSystemCreated = err == nil
	session.Status.Error = err

	session, err = s.registry.UpdateSession(session)
	if err != nil {
		log.Printf("Failed to update session: %+v", session)
	} else {
		action.Session = session
	}

	if err := s.actions.CompleteSessionAction(action, action.Session.Status.Error); err != nil {
		log.Printf("Failed to complete action: %+v", action)
	}
	if action.Session.Status.Error != nil {
		log.Println("error during create for", sessionName, err)
	} else {
		log.Printf("completed create for %+v\n", sessionName)
	}
}

func (s *sessionActionHandler) handleDelete(action datamodel.SessionAction) {
	log.Println("delete")
	s.actions.CompleteSessionAction(action, nil)
}
