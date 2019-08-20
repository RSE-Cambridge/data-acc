package brick_manager_impl

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/filesystem"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/filesystem_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"log"
)

func NewSessionActionHandler(keystore store.Keystore) facade.SessionActionHandler {
	return &sessionActionHandler{
		registry_impl.NewSessionRegistry(keystore),
		registry_impl.NewSessionActionsRegistry(keystore),
		// TODO: fix up fsprovider!!
		filesystem_impl.NewFileSystemProvider(nil),
		false,
	}
}

type sessionActionHandler struct {
	sessionRegistry registry.SessionRegistry
	actions         registry.SessionActions
	fsProvider      filesystem.Provider
	skipActions     bool
}

func (s *sessionActionHandler) ProcessSessionAction(action datamodel.SessionAction) {
	log.Printf("Started to process: %+v\n", action)
	switch action.ActionType {
	case datamodel.SessionDelete:
		// TODO... must test this better!
		if !s.skipActions {
			s.handleDelete(action)
		}
	case datamodel.SessionCreateFilesystem:
		if !s.skipActions {
			// TODO: really should all happen in a goroutine
			s.handleCreate(action)
		}
	case datamodel.SessionCopyDataIn:
		s.handleCopyIn(action)
	case datamodel.SessionMount:
		s.handleMount(action)
	case datamodel.SessionUnmount:
		s.handleUnmount(action)
	case datamodel.SessionCopyDataOut:
		s.handleCopyOut(action)
	default:
		log.Panicf("not yet implemented action for %+v", action)
	}
}

func (s *sessionActionHandler) processWithMutex(action datamodel.SessionAction, process func() (datamodel.Session, error)) {

	sessionName := action.Session.Name
	sessionMutex, err := s.sessionRegistry.GetSessionMutex(sessionName)
	if err != nil {
		log.Printf("unable to get session mutex: %s due to: %s\n", sessionName, err)
		action.Error = err
		return
	}
	err = sessionMutex.Lock(context.TODO())
	if err != nil {
		log.Printf("unable to lock session mutex: %s due to: %s\n", sessionName, err)
		action.Error = err
		return
	}

	// Always complete action and drop mutex on function exit
	defer func() {
		if err := s.actions.CompleteSessionAction(action); err != nil {
			log.Printf("failed to complete action %+v\n", action)
		}
		if err := sessionMutex.Unlock(context.TODO()); err != nil {
			log.Println("failed to drop mutex for:", sessionName)
		}
	}()

	log.Printf("starting action %+v\n", action)

	session, err := process()
	if err != nil {
		action.Error = err
		log.Printf("error during action %+v\n", action)
	} else {
		action.Session = session
		log.Printf("finished action %+v\n", action)
	}
}

func (s *sessionActionHandler) handleCreate(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		// Get latest session now we have the mutex
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}

		fsStatus, err := s.fsProvider.Create(action.Session)
		session.FilesystemStatus = fsStatus
		session.Status.FileSystemCreated = err == nil
		session.Status.Error = err

		session, updateErr := s.sessionRegistry.UpdateSession(session)
		if updateErr != nil {
			log.Println("Failed to update session:", updateErr)
			if err == nil {
				err = updateErr
			}
		}
		return session, err
	})
}

func (s *sessionActionHandler) handleDelete(action datamodel.SessionAction) {
	// TODO... mutex, etc?
	s.doUnmount(action)
	if !action.Session.Status.DeleteSkipCopyDataOut && !action.Session.Status.CopyDataOutComplete {
		s.fsProvider.DataCopyOut(action.Session)
	}
	s.fsProvider.Delete(action.Session)
	s.sessionRegistry.DeleteSession(action.Session)
	s.actions.CompleteSessionAction(action)
}

func (s *sessionActionHandler) handleCopyIn(action datamodel.SessionAction) {
	s.fsProvider.DataCopyIn(action.Session)
	s.actions.CompleteSessionAction(action)
}

func (s *sessionActionHandler) handleMount(action datamodel.SessionAction) {
	if action.Session.ActualSizeBytes > 0 {
		s.fsProvider.Mount(action.Session,
			datamodel.AttachmentSession{Hosts: action.Session.RequestedAttachHosts})
	}
	for _, sessionName := range action.Session.MultiJobAttachments {
		session, _ := s.sessionRegistry.GetSession(sessionName)
		if session.VolumeRequest.MultiJob {
			s.fsProvider.Mount(session,
				datamodel.AttachmentSession{Hosts: action.Session.RequestedAttachHosts})
		}
	}
	s.actions.CompleteSessionAction(action)
}

func (s *sessionActionHandler) doUnmount(action datamodel.SessionAction) {
	if action.Session.ActualSizeBytes > 0 {
		s.fsProvider.Unmount(action.Session,
			datamodel.AttachmentSession{Hosts: action.Session.RequestedAttachHosts})
	}
	for _, sessionName := range action.Session.MultiJobAttachments {
		session, _ := s.sessionRegistry.GetSession(sessionName)
		if session.VolumeRequest.MultiJob {
			s.fsProvider.Mount(session,
				datamodel.AttachmentSession{Hosts: action.Session.RequestedAttachHosts})
		}
	}
}
func (s *sessionActionHandler) handleUnmount(action datamodel.SessionAction) {
	s.doUnmount(action)
	s.actions.CompleteSessionAction(action)
}

func (s *sessionActionHandler) handleCopyOut(action datamodel.SessionAction) {
	s.fsProvider.DataCopyOut(action.Session)
	// TODO: update session.CopyDataOutComplete, mutex, etc.
	s.actions.CompleteSessionAction(action)
}
