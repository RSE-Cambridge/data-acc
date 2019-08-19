package registry_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl/actions_impl/parsers"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"github.com/google/uuid"
	"log"
)

func NewSessionActionsRegistry(store store.Keystore, brickHostRegistry registry.BrickHostRegistry) registry.SessionActions {
	// TODO: create brickHostRegistry
	return &sessionActions{store, brickHostRegistry}
}

type sessionActions struct {
	store             store.Keystore
	brickHostRegistry registry.BrickHostRegistry
}

const sessionActionRequestPrefix = "/session_action/request/"

func getSessionActionRequestHostPrefix(brickHost datamodel.BrickHostName) string {
	if !parsers.IsValidName(string(brickHost)) {
		log.Panicf("invalid session PrimaryBrickHost")
	}
	return fmt.Sprintf("%s%s/", sessionActionRequestPrefix, brickHost)
}

func getSessionActionRequestKey(action datamodel.SessionAction) string {
	hostPrefix := getSessionActionRequestHostPrefix(action.Session.PrimaryBrickHost)
	if !parsers.IsValidName(action.Uuid) {
		log.Panicf("invalid session action uuid")
	}
	return fmt.Sprintf("%s%s", hostPrefix, action.Uuid)
}

const sessionActionResponsePrefix = "/session_action/response/"

func getSessionActionResponseHostPrefix(brickHost datamodel.BrickHostName) string {
	if !parsers.IsValidName(string(brickHost)) {
		log.Panicf("invalid session PrimaryBrickHost")
	}
	return fmt.Sprintf("%s%s/", sessionActionResponsePrefix, brickHost)
}

func getSessionActionResponseKey(action datamodel.SessionAction) string {
	hostPrefix := getSessionActionResponseHostPrefix(action.Session.PrimaryBrickHost)
	if !parsers.IsValidName(action.Uuid) {
		log.Panicf("invalid session action uuid")
	}
	return fmt.Sprintf("%s%s", hostPrefix, action.Uuid)
}

func sessionActionToRaw(session datamodel.SessionAction) []byte {
	rawSession, err := json.Marshal(session)
	if err != nil {
		log.Panicf("unable to convert session action to json due to: %s", err.Error())
	}
	return rawSession
}

func sessionActionFromRaw(raw []byte) datamodel.SessionAction {
	session := datamodel.SessionAction{}
	err := json.Unmarshal(raw, &session)
	if err != nil {
		log.Panicf("unable parse session action from store due to: %s", err)
	}
	return session
}

func (s *sessionActions) SendSessionAction(
	ctxt context.Context, actionType datamodel.SessionActionType,
	session datamodel.Session) (<-chan datamodel.SessionAction, error) {

	if session.PrimaryBrickHost == "" {
		panic("sessions must have a primary brick host set")
	}
	sessionAction := datamodel.SessionAction{
		Session:    session,
		ActionType: actionType,
		Uuid:       uuid.New().String(),
	}

	isAlive, err := s.brickHostRegistry.IsBrickHostAlive(session.PrimaryBrickHost)
	if err != nil {
		return nil, fmt.Errorf("unable to check host status: %s", session.PrimaryBrickHost)
	}
	if !isAlive {
		return nil, fmt.Errorf("can't send as primary brick host not alive: %s", session.PrimaryBrickHost)
	}

	responseKey := getSessionActionResponseKey(sessionAction)
	callbackKeyUpdates := s.store.Watch(ctxt, responseKey, false)

	requestKey := getSessionActionRequestKey(sessionAction)
	if _, err := s.store.Create(requestKey, sessionActionToRaw(sessionAction)); err != nil {
		return nil, fmt.Errorf("unable to send session action due to: %s", err)
	}

	responseChan := make(chan datamodel.SessionAction)

	go func() {
		log.Printf("started waiting for action response %+v\n", sessionAction)
		for update := range callbackKeyUpdates {
			if !update.IsCreate || update.New.Value == nil {
				log.Panicf("only expected to see the action response key being created")
			}

			responseSessionAction := sessionActionFromRaw(update.New.Value)
			log.Printf("found action response %+v\n", responseSessionAction)

			responseChan <- responseSessionAction
			close(responseChan)

			log.Printf("completed waiting for action response %+v\n", sessionAction)
			return
		}
	}()
	return responseChan, nil
}

func (s *sessionActions) GetSessionActionRequests(ctxt context.Context,
	brickHostName datamodel.BrickHostName) (<-chan datamodel.SessionAction, error) {
	requestHostPrefix := getSessionActionRequestHostPrefix(brickHostName)

	// TODO: how do we check for any pending actions that exist before we start watching?
	//   or do we only care about pending deletes, and we let them just timeout?
	requestUpdates := s.store.Watch(ctxt, requestHostPrefix, true)

	sessionActionChan := make(chan datamodel.SessionAction)
	go func() {
		log.Printf("Starting watching for SessionActionRequests for %s\n", brickHostName)
		for update := range requestUpdates {
			if update.IsDelete {
				log.Printf("Seen SessionActionRequest deleted for %s\n", brickHostName)
				continue
			}
			if !update.IsCreate || update.New.Value == nil {
				log.Panicf("don't expect to see updates of session action request key")
			}
			log.Printf("Seen SessionActionRequest created for %s\n", brickHostName)

			sessionAction := sessionActionFromRaw(update.New.Value)
			sessionActionChan <- sessionAction
		}
		log.Printf("Stopped watching for SessionActionRequests for %s\n", brickHostName)
		close(sessionActionChan)
	}()
	return sessionActionChan, nil
}

func (s *sessionActions) CompleteSessionAction(sessionAction datamodel.SessionAction) error {
	// TODO: when you delete a session, you should delete all completion records?

	// Tell caller we are done by writing this key
	responseKey := getSessionActionResponseKey(sessionAction)
	_, err := s.store.Create(responseKey, sessionActionToRaw(sessionAction))
	if err != nil {
		return fmt.Errorf("unable to create response message due to: %s", err)
	}

	// Delete the request, not it is processed
	requestKey := getSessionActionRequestKey(sessionAction)
	err = s.store.Delete(requestKey, 0)
	if err != nil {
		return fmt.Errorf("unable to delete stale request message due to: %s", err)
	}
	return nil
}
