package brick_manager_impl

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/filesystem"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/filesystem_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store"
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
	switch action.ActionType {
	case datamodel.SessionDelete:
		go s.handleDelete(action)
	case datamodel.SessionCreateFilesystem:
		go s.handleCreate(action)
	case datamodel.SessionCopyDataIn:
		go s.handleCopyIn(action)
	case datamodel.SessionMount:
		go s.handleMount(action)
	case datamodel.SessionUnmount:
		go s.handleUnmount(action)
	case datamodel.SessionCopyDataOut:
		go s.handleCopyOut(action)
	default:
		log.Panicf("not yet implemented action for %+v", action)
	}
}

func (s *sessionActionHandler) processWithMutex(action datamodel.SessionAction, process func() (datamodel.Session, error)) {

	sessionName := action.Session.Name
	sessionMutex, err := s.sessionRegistry.GetSessionMutex(sessionName)
	if err != nil {
		log.Printf("unable to get session mutex: %s due to: %s\n", sessionName, err)
		action.Error = err.Error()
		return
	}
	err = sessionMutex.Lock(context.TODO())
	if err != nil {
		log.Printf("unable to lock session mutex: %s due to: %s\n", sessionName, err)
		action.Error = err.Error()
		return
	}

	// Always complete action and drop mutex on function exit
	defer func() {
		if err := s.actions.CompleteSessionAction(action); err != nil {
			log.Printf("failed to complete action %+v due to: %s\n", action, err.Error())
		}
		if err := sessionMutex.Unlock(context.TODO()); err != nil {
			log.Printf("failed to drop mutex for: %s due to: %s\n", sessionName, err.Error())
		}
	}()

	log.Printf("starting action %+v\n", action)

	session, err := process()
	if err != nil {
		action.Error = err.Error()
		log.Printf("error during action %+v\n", action)
	} else {
		action.Session = session
		log.Printf("finished action %+v\n", action)
	}
}

func (s *sessionActionHandler) handleCreate(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		// Nothing to create, just complete the action
		// TODO: why do we send the action?
		if action.Session.ActualSizeBytes == 0 {
			return action.Session, nil
		}

		// Get latest session now we have the mutex
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}

		fsStatus, err := s.fsProvider.Create(action.Session)
		session.FilesystemStatus = fsStatus
		session.Status.FileSystemCreated = err == nil
		if err != nil {
			session.Status.Error = err.Error()
		}

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
	s.processWithMutex(action, func() (datamodel.Session, error) {
		if !action.Session.Status.UnmountComplete {
			if err := s.doAllUnmounts(action); err != nil {
				log.Println("failed unmount during delete", action.Session.Name)
			}
		}
		if !action.Session.Status.CopyDataOutComplete && !action.Session.Status.DeleteSkipCopyDataOut {
			if err := s.fsProvider.DataCopyOut(action.Session); err != nil {
				log.Println("failed DataCopyOut during delete", action.Session.Name)
			}
		}

		// Only try delete if we have bricks to delete
		if action.Session.ActualSizeBytes > 0 {
			if err := s.fsProvider.Delete(action.Session); err != nil {
				return action.Session, err
			}
		}

		return action.Session, s.sessionRegistry.DeleteSession(action.Session)
	})
}

func (s *sessionActionHandler) handleCopyIn(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		err := s.fsProvider.DataCopyIn(action.Session)
		if err != nil {
			return action.Session, err
		}

		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			session = action.Session
		}
		session.Status.CopyDataInComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) handleCopyOut(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		err := s.fsProvider.DataCopyOut(action.Session)
		if err != nil {
			return action.Session, err
		}

		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			session = action.Session
		}
		session.Status.CopyDataOutComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) doAllMounts(action datamodel.SessionAction) error {
	attachmentSession := datamodel.AttachmentSession{
		Hosts:       action.Session.RequestedAttachHosts,
		SessionName: action.Session.Name,
	}
	if action.Session.ActualSizeBytes > 0 {
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return err
		}
		jobAttachmentStatus := datamodel.AttachmentSessionStatus{
			AttachmentSession: attachmentSession,
			GlobalMount:       session.VolumeRequest.Access == datamodel.Striped || session.VolumeRequest.Access == datamodel.PrivateAndStriped,
			PrivateMount:      session.VolumeRequest.Access == datamodel.Private || session.VolumeRequest.Access == datamodel.PrivateAndStriped,
			SwapBytes:         session.VolumeRequest.SwapBytes,
		}
		if session.CurrentAttachments == nil {
			session.CurrentAttachments = map[datamodel.SessionName]datamodel.AttachmentSessionStatus{
				session.Name: jobAttachmentStatus,
			}
		} else {
			session.CurrentAttachments[session.Name] = jobAttachmentStatus
		}
		session, err = s.sessionRegistry.UpdateSession(session)
		if err != nil {
			return err
		}

		if err := s.fsProvider.Mount(action.Session, jobAttachmentStatus); err != nil {
			return err
		}
	}
	for _, sessionName := range action.Session.MultiJobAttachments {
		if err := s.doMutliJobMount(action, sessionName); err != nil {
			return nil
		}
	}
	return nil
}

func (s *sessionActionHandler) doMutliJobMount(action datamodel.SessionAction, sessionName datamodel.SessionName) error {
	sessionMutex, err := s.sessionRegistry.GetSessionMutex(sessionName)
	if err != nil {
		log.Printf("unable to get session mutex: %s due to: %s\n", sessionName, err)
		return err
	}
	if err = sessionMutex.Lock(context.TODO()); err != nil {
		log.Printf("unable to lock session mutex: %s due to: %s\n", sessionName, err)
		return err
	}
	defer func() {
		if err := sessionMutex.Unlock(context.TODO()); err != nil {
			log.Println("failed to drop mutex for:", sessionName)
		}
	}()

	multiJobSession, err := s.sessionRegistry.GetSession(sessionName)
	if err != nil {
		return err
	}
	if !multiJobSession.VolumeRequest.MultiJob {
		log.Panicf("trying multi-job attach to non-multi job session %s", multiJobSession.Name)
	}

	attachmentSession := datamodel.AttachmentSession{
		Hosts:       action.Session.RequestedAttachHosts,
		SessionName: action.Session.Name,
	}
	multiJobAttachmentStatus := datamodel.AttachmentSessionStatus{
		AttachmentSession: attachmentSession,
		GlobalMount:       true,
	}
	if multiJobSession.CurrentAttachments == nil {
		multiJobSession.CurrentAttachments = map[datamodel.SessionName]datamodel.AttachmentSessionStatus{
			attachmentSession.SessionName: multiJobAttachmentStatus,
		}
	} else {
		if _, ok := multiJobSession.CurrentAttachments[attachmentSession.SessionName]; ok {
			return fmt.Errorf("already attached for session %s and multi-job %s",
				attachmentSession.SessionName, sessionName)
		}
		multiJobSession.CurrentAttachments[attachmentSession.SessionName] = multiJobAttachmentStatus
	}

	multiJobSession, err = s.sessionRegistry.UpdateSession(multiJobSession)
	if err != nil {
		return err
	}
	return s.fsProvider.Mount(multiJobSession, multiJobAttachmentStatus)
}

func (s *sessionActionHandler) doAllUnmounts(action datamodel.SessionAction) error {
	if action.Session.ActualSizeBytes > 0 {
		if err := s.fsProvider.Unmount(action.Session, action.Session.CurrentAttachments[action.Session.Name]); err != nil {
			return err
		}
		// TODO: delete attachments?
	}
	for _, sessionName := range action.Session.MultiJobAttachments {
		multiJobSession, _ := s.sessionRegistry.GetSession(sessionName)
		attachments := multiJobSession.CurrentAttachments[action.Session.Name]
		if err := s.fsProvider.Unmount(multiJobSession, attachments); err != nil {
			return err
		}
		// TODO: delete attachments to prevent double unmount on teardown?
	}
	return nil
}

func (s *sessionActionHandler) handleMount(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		err := s.doAllMounts(action)
		if err != nil {
			if err := s.doAllUnmounts(action); err != nil {
				log.Println("error while rolling back possible partial mount", action.Session.Name, err)
			}
			return action.Session, err
		}

		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			session = action.Session
		}
		session.Status.MountComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) handleUnmount(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		err := s.doAllUnmounts(action)
		if err != nil {
			return action.Session, err
		}

		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			session = action.Session
		}
		session.Status.UnmountComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}
