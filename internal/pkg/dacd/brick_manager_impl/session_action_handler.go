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
		session := action.Session
		// Nothing to create, just complete the action
		// TODO: why do we send the action?
		if session.ActualSizeBytes == 0 && len(session.MultiJobAttachments) == 0 {
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

		// Only call create if we have a per job buffer to create
		// Note: we always need to do the mount to enable copy-in/out
		if session.ActualSizeBytes != 0 {
			fsStatus, err := s.fsProvider.Create(session)
			session.FilesystemStatus = fsStatus
			if err != nil {
				session.Status.Error = err.Error()
			}

			var updateErr error
			session, updateErr = s.sessionRegistry.UpdateSession(session)
			if updateErr != nil {
				log.Println("Failed to update session:", updateErr)
				if err == nil {
					err = updateErr
				}
			}
			if err != nil {
				return session, err
			}
			log.Println("Filesystem created, now mount on primary brick host")
		}

		session, err = s.doAllMounts(session, true)
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
		session, err := s.sessionRegistry.GetSession(action.Session.Name)
		if err != nil {
			// TODO: deal with already being deleted? add check if exists call?
			return action.Session, fmt.Errorf("error getting session: %s", err)
		}

		if err := s.doAllUnmounts(session, getAttachmentKey(session.Name, true)); err != nil {
			return session, fmt.Errorf("failed primary brick host unmount, due to: %s", err.Error())
		}
		log.Println("did umount primary brick host during delete")

		if !session.Status.UnmountComplete {
			if err := s.doAllUnmounts(session, getAttachmentKey(session.Name, false)); err != nil {
				return session, fmt.Errorf("failed retry unmount during delete, due to: %s", err.Error())
			}
			log.Println("did unmount during delete")
		}
		if !session.Status.CopyDataOutComplete && !session.Status.DeleteSkipCopyDataOut {
			if err := s.fsProvider.DataCopyOut(action.Session); err != nil {
				return session, fmt.Errorf("failed DataCopyOut during delete, due to: %s", err.Error())
			}
			log.Println("did data copy out during delete")
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

func addHostsFromSession(attachment *datamodel.AttachmentSession, actionSession datamodel.Session, forPrimaryBrickHost bool) {
	if forPrimaryBrickHost {
		attachment.Hosts = []string{string(actionSession.PrimaryBrickHost)}
	} else {
		attachment.Hosts = actionSession.RequestedAttachHosts
	}
}

func (s *sessionActionHandler) doAllMounts(actionSession datamodel.Session, forPrimaryBrickHost bool) (datamodel.Session, error) {
	if actionSession.ActualSizeBytes > 0 {
		jobAttachment := datamodel.AttachmentSession{
			SessionName:  actionSession.Name,
			GlobalMount:  actionSession.VolumeRequest.Access == datamodel.Striped || actionSession.VolumeRequest.Access == datamodel.PrivateAndStriped,
			PrivateMount: actionSession.VolumeRequest.Access == datamodel.Private || actionSession.VolumeRequest.Access == datamodel.PrivateAndStriped,
			SwapBytes:    actionSession.VolumeRequest.SwapBytes,
		}
		if forPrimaryBrickHost {
			// Never deal with private mount, as make no sense for copy in to private dir
			jobAttachment.PrivateMount = false
		}
		addHostsFromSession(&jobAttachment, actionSession, forPrimaryBrickHost)
		if err := updateAttachments(&actionSession, jobAttachment, forPrimaryBrickHost); err != nil {
			return actionSession, err
		}
		session, err := s.sessionRegistry.UpdateSession(actionSession)
		if err != nil {
			return actionSession, err
		}
		actionSession = session

		if err := s.fsProvider.Mount(actionSession, jobAttachment, forPrimaryBrickHost); err != nil {
			return actionSession, err
		}
		// TODO: should we track success of each attachment session?
	}
	for _, sessionName := range actionSession.MultiJobAttachments {
		if err := s.doMultiJobMount(actionSession, sessionName, forPrimaryBrickHost); err != nil {
			return actionSession, nil
		}
	}
	return actionSession, nil
}

func getAttachmentKey(sessionName datamodel.SessionName, forPrimaryBrickHost bool) datamodel.SessionName {
	if forPrimaryBrickHost {
		return datamodel.SessionName(fmt.Sprintf("Primary_%s", sessionName))
	} else {
		return sessionName
	}
}

func updateAttachments(session *datamodel.Session, attachment datamodel.AttachmentSession, forPrimaryBrickHost bool) error {
	attachmentKey := getAttachmentKey(attachment.SessionName, forPrimaryBrickHost)
	if session.CurrentAttachments == nil {
		session.CurrentAttachments = map[datamodel.SessionName]datamodel.AttachmentSession{
			attachmentKey: attachment,
		}
	} else {
		if _, ok := session.CurrentAttachments[attachmentKey]; ok {
			return fmt.Errorf("already attached for session %s and target-volume %s",
				attachment.SessionName, session.Name)
		}
		session.CurrentAttachments[attachmentKey] = attachment
	}
	return nil
}

func (s *sessionActionHandler) doMultiJobMount(actionSession datamodel.Session, sessionName datamodel.SessionName, forPrimaryBrickHost bool) error {
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

	multiJobAttachment := datamodel.AttachmentSession{
		SessionName: actionSession.Name,
		GlobalMount: true,
	}
	addHostsFromSession(&multiJobAttachment, actionSession, forPrimaryBrickHost)
	if err := updateAttachments(&multiJobSession, multiJobAttachment, forPrimaryBrickHost); err != nil {
		return err
	}

	multiJobSession, err = s.sessionRegistry.UpdateSession(multiJobSession)
	if err != nil {
		return err
	}
	return s.fsProvider.Mount(multiJobSession, multiJobAttachment, false)
}

func (s *sessionActionHandler) doMultiJobUnmount(actionSession datamodel.Session, sessionName datamodel.SessionName, attachmentKey datamodel.SessionName) error {
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

	attachments, ok := multiJobSession.CurrentAttachments[attachmentKey]
	if !ok {
		log.Println("skip multi-job detach, already seems to be detached")
		return nil
	}
	if err := s.fsProvider.Unmount(multiJobSession, attachments); err != nil {
		return err
	}

	// update multi job session to note our attachments have now gone
	delete(multiJobSession.CurrentAttachments, attachmentKey)
	_, err = s.sessionRegistry.UpdateSession(multiJobSession)
	return err
}

func (s *sessionActionHandler) doAllUnmounts(actionSession datamodel.Session, attachmentKey datamodel.SessionName) error {
	if actionSession.ActualSizeBytes > 0 {
		if err := s.fsProvider.Unmount(actionSession, actionSession.CurrentAttachments[attachmentKey]); err != nil {
			return err
		}
	}
	for _, sessionName := range actionSession.MultiJobAttachments {
		if err := s.doMultiJobUnmount(actionSession, sessionName, attachmentKey); err != nil {
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

		session, err = s.doAllMounts(session, false)
		if err != nil {
			if err := s.doAllUnmounts(session, getAttachmentKey(session.Name, false)); err != nil {
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

		if err := s.doAllUnmounts(session, getAttachmentKey(session.Name, false)); err != nil {
			return action.Session, err
		}

		session.Status.UnmountComplete = true
		return s.sessionRegistry.UpdateSession(session)
	})
}

func (s *sessionActionHandler) RestoreSession(session datamodel.Session) {
	if session.ActualSizeBytes == 0 {
		// Nothing to do
		return
	}

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

	err = s.fsProvider.Restore(session)

	if err != nil {
		log.Printf("unable to restore session: %+v\n", session)
		session.Status.Error = err.Error()
		if _, err := s.sessionRegistry.UpdateSession(session); err != nil {
			log.Panicf("unable to report that session restore failed for session: %s", session.Name)
		}
	}

	// TODO: do we just assume any pending mounts will resume in their own time? or should we retry mounts too?
}
