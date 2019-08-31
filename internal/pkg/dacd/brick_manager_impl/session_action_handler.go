package brick_manager_impl

import (
	"context"
	"errors"
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
		s.handleDelete(action)
	case datamodel.SessionCreateFilesystem:
		s.handleCreate(action)
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

func (s *sessionActionHandler) RestoreSession(session datamodel.Session) {
	// Get session lock before attempting the restore
	sessionMutex, err := s.sessionRegistry.GetSessionMutex(session.Name)
	if err != nil {
		log.Printf("unable to get session mutex: %s due to: %s\n", session.Name, err)
		return
	}
	err = sessionMutex.Lock(context.TODO())
	if err != nil {
		log.Printf("unable to lock session mutex: %s due to: %s\n", session.Name, err)
		return
	}
	// Always drop mutex on function exit
	defer func() {
		if err := sessionMutex.Unlock(context.TODO()); err != nil {
			log.Printf("failed to drop mutex for: %s due to: %s\n", session.Name, err.Error())
		}
	}()

	// TODO: need a way that doesn't try to do format!!
	_, err = s.doCreate(session)
	if err != nil {
		log.Printf("unable to restore session: %+v\n", session)
		session.Status.Error = err.Error()
		if _, err := s.sessionRegistry.UpdateSession(session); err != nil {
			log.Panicf("unable to report that session restore failed for session: %s", session.Name)
		}
	}

	// TODO: do we just assume any pending mounts will resume in their own time? or should we retry mounts too?
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
		return s.doCreate(action.Session)
	})
}

func (s *sessionActionHandler) doCreate(session datamodel.Session) (datamodel.Session, error) {
	// Nothing to create, just complete the action
	// TODO: why do we send the action?
	if session.ActualSizeBytes == 0 {
		return session, nil
	}

	// Get latest session now we have the mutex
	session, err := s.sessionRegistry.GetSession(session.Name)
	if err != nil {
		return session, fmt.Errorf("error getting session: %s", err)
	}
	if session.Status.DeleteRequested {
		return session, fmt.Errorf("can't do action once delete has been requested for")
	}

	fsStatus, err := s.fsProvider.Create(session)
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
}

func (s *sessionActionHandler) handleDelete(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			// TODO: deal with already being deleted? add check if exists call?
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}

		if !session.Status.UnmountComplete {
			if err := s.doAllUnmounts(session); err != nil {
				log.Println("failed unmount during delete", session.Name)
			}
		}
		if !session.Status.CopyDataOutComplete && !session.Status.DeleteSkipCopyDataOut {
			if err := s.fsProvider.DataCopyOut(action.Session); err != nil {
				log.Println("failed DataCopyOut during delete", action.Session.Name)
			}
		}

		// Only try delete if we have bricks to delete
		if session.ActualSizeBytes > 0 {
			if err := s.fsProvider.Delete(session); err != nil {
				return session, err
			}
		}

		return session, s.sessionRegistry.DeleteSession(session)
	})
}

func (s *sessionActionHandler) handleCopyIn(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		// Get latest session now we have the mutex
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}
		if session.Status.DeleteRequested {
			return session, fmt.Errorf("can't do action once delete has been requested for")
		}

		if err := s.fsProvider.DataCopyIn(session); err != nil {
			return session, err
		}

		session.Status.CopyDataInComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) handleCopyOut(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		// Get latest session now we have the mutex
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}
		if session.Status.DeleteRequested {
			return session, fmt.Errorf("can't do action once delete has been requested for")
		}

		if err := s.fsProvider.DataCopyOut(session); err != nil {
			return session, err
		}

		session.Status.CopyDataOutComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) doAllMounts(actionSession datamodel.Session) error {
	attachmentSession := datamodel.AttachmentSession{
		Hosts:       actionSession.RequestedAttachHosts,
		SessionName: actionSession.Name,
	}
	if actionSession.ActualSizeBytes > 0 {
		jobAttachmentStatus := datamodel.AttachmentSessionStatus{
			AttachmentSession: attachmentSession,
			GlobalMount:       actionSession.VolumeRequest.Access == datamodel.Striped || actionSession.VolumeRequest.Access == datamodel.PrivateAndStriped,
			PrivateMount:      actionSession.VolumeRequest.Access == datamodel.Private || actionSession.VolumeRequest.Access == datamodel.PrivateAndStriped,
			SwapBytes:         actionSession.VolumeRequest.SwapBytes,
		}
		//if actionSession.CurrentAttachments == nil {
		//	session.CurrentAttachments = map[datamodel.SessionName]datamodel.AttachmentSessionStatus{
		//		session.Name: jobAttachmentStatus,
		//	}
		//} else {
		//	session.CurrentAttachments[session.Name] = jobAttachmentStatus
		//}
		//session, err = s.sessionRegistry.UpdateSession(session)
		//if err != nil {
		//	return err
		//}

		if err := s.fsProvider.Mount(actionSession, jobAttachmentStatus); err != nil {
			return err
		}
		// TODO: should we update the session? and delete attachments later?
	}
	for _, sessionName := range actionSession.MultiJobAttachments {
		if err := s.doMultiJobMount(actionSession, sessionName); err != nil {
			return nil
		}
	}
	return nil
}

func (s *sessionActionHandler) doMultiJobMount(actionSession datamodel.Session, sessionName datamodel.SessionName) error {
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
		Hosts:       actionSession.RequestedAttachHosts,
		SessionName: actionSession.Name,
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

func (s *sessionActionHandler) doMultiJobUnmount(actionSession datamodel.Session, sessionName datamodel.SessionName) error {
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

	attachments, ok := multiJobSession.CurrentAttachments[actionSession.Name]
	if !ok {
		log.Println("skip detach, already seems to be detached")
		return nil
	}
	if err := s.fsProvider.Unmount(multiJobSession, attachments); err != nil {
		return err
	}

	// update multi job session to note our attachments have now gone
	delete(multiJobSession.CurrentAttachments, actionSession.Name)
	_, err = s.sessionRegistry.UpdateSession(multiJobSession)
	return err
}

func (s *sessionActionHandler) doAllUnmounts(actionSession datamodel.Session) error {
	if actionSession.ActualSizeBytes > 0 {
		if err := s.fsProvider.Unmount(actionSession, actionSession.CurrentAttachments[actionSession.Name]); err != nil {
			return err
		}
	}
	for _, sessionName := range actionSession.MultiJobAttachments {
		if err := s.doMultiJobUnmount(actionSession, sessionName); err != nil {
			return err
		}
	}
	return nil
}

func (s *sessionActionHandler) handleMount(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}
		if session.Status.DeleteRequested {
			return session, fmt.Errorf("can't do action once delete has been requested for")
		}
		if session.Status.MountComplete {
			return session, errors.New("already mounted, can't mount again")
		}

		if err := s.doAllMounts(session); err != nil {
			if err := s.doAllUnmounts(session); err != nil {
				log.Println("error while rolling back possible partial mount", action.Session.Name, err)
			}
			return action.Session, err
		}

		session.Status.MountComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) handleUnmount(action datamodel.SessionAction) {
	s.processWithMutex(action, func() (datamodel.Session, error) {
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}
		if session.Status.DeleteRequested {
			return session, fmt.Errorf("can't do action once delete has been requested for")
		}
		if session.Status.UnmountComplete {
			return session, errors.New("already unmounted, can't umount again")
		}

		if err := s.doAllUnmounts(session); err != nil {
			return action.Session, err
		}

		session.Status.UnmountComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}
