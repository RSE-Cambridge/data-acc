package registry_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"github.com/google/uuid"
	"log"
	"time"
)

// TODO: this is the client side, need server side too
func NewSessionActions(keystore store.Keystore) registry.SessionActions {
	return &sessionActions{keystore: keystore, defaultTimeout: time.Minute * 20}
}

type sessionActions struct {
	keystore       store.Keystore
	defaultTimeout time.Duration
}

func (s *sessionActions) CreateSessionVolume(ctxt context.Context, sessionName datamodel.SessionName) (<-chan datamodel.Session, error) {
	panic("implement me")
}

func (s *sessionActions) SendSessionAction(
	ctxt context.Context, actionType datamodel.SessionActionType,
	session datamodel.Session) (<-chan datamodel.Session, error) {
	panic("implement me")
}

func (s *sessionActions) GetCreateSessionVolumeRequests(
	ctxt context.Context, brickHostName datamodel.BrickHostName) (<-chan datamodel.SessionAction, error) {
	panic("implement me")
}

func (s *sessionActions) GetSessionActions(
	ctxt context.Context, sessionName datamodel.SessionName) (<-chan datamodel.SessionAction, error) {
	panic("implement me")
}

func (s *sessionActions) CompleteSessionAction(action datamodel.SessionAction, err error) error {
	panic("implement me")
}

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	return string(b)
}

func (s *sessionActions) sendAction(session datamodel.Session, action string) error {
	// TODO: update session?

	// TODO: check primary session host is alive before sending event
	// TODO: If we timeout, cancel the event only if the host is now not alive

	actionId := uuid.New().String()
	sessionPrefix := fmt.Sprintf("/Session/Actions/%s", session.Name)
	actionPrefix := fmt.Sprintf("%s/%s", sessionPrefix, actionId)

	mutex, err := s.keystore.NewMutex(sessionPrefix)
	if err != nil {
		return fmt.Errorf("unable to start action due to: %s", err.Error())
	}
	mctxt, cancel := context.WithTimeout(context.Background(), s.defaultTimeout)
	defer cancel()
	err = mutex.Lock(mctxt)
	if err != nil {
		return fmt.Errorf("unable to start action due to: %s", err.Error())
	}
	defer mutex.Unlock(context.Background())

	ctxt, cancel := context.WithTimeout(context.Background(), s.defaultTimeout)
	defer cancel()
	responses := s.keystore.Watch(ctxt, fmt.Sprintf("%s/%s", actionPrefix, "output"), false)

	err = s.keystore.Add(store.KeyValue{
		Key: fmt.Sprintf("%s/%s", actionPrefix, "input"),
		// TODO: need format request object
		Value: action,
	})
	if err != nil {
		return err
	}

	// TODO: how do we detect a timeout here?
	for response := range responses {
		if response.Err != nil {
			return response.Err
		}

		if response.IsCreate {
			// TODO: need formal response object?
			if response.New.Value != "" {
				return fmt.Errorf("error while sending action")
			}
			// TODO: who deletes the action key, if anyone?
		}
		return nil
	}
	return nil
}
