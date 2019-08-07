package session_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/session"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/store"
	"github.com/google/uuid"
	"log"
	"time"
)

// TODO: this is the client side, need server side too
func NewSessionActions(keystore store.Keystore) session.Actions {
	return &sessionActions{keystore: keystore}
}

type sessionActions struct {
	keystore store.Keystore
}

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	return string(b)
}

func (s *sessionActions) sendAction(session model.Session, action string) error {
	// TODO: update session?

	actionId := uuid.New().String()
	sessionPrefix := fmt.Sprintf("/Session/Actions/%s", session.Name)
	actionPrefix := fmt.Sprintf("%s/%s", sessionPrefix, actionId)

	mutex, err := s.keystore.NewMutex(sessionPrefix)
	if err != nil {
		return fmt.Errorf("unable to start action due to: %s", err.Error())
	}
	mctxt, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()
	err = mutex.Lock(mctxt)
	if err != nil {
		return fmt.Errorf("unable to start action due to: %s", err.Error())
	}
	defer mutex.Unlock(context.Background())

	ctxt, cancel := context.WithTimeout(context.Background(), time.Minute*20)
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

func (*sessionActions) CreateSessionVolume(session model.Session) error {
	panic("implement me")
}

func (*sessionActions) DeleteSession(session model.Session) error {
	panic("implement me")
}

func (*sessionActions) DataIn(session model.Session) error {
	panic("implement me")
}

func (*sessionActions) AttachVolumes(session model.Session) error {
	panic("implement me")
}

func (*sessionActions) DetachVolumes(session model.Session) error {
	panic("implement me")
}

func (*sessionActions) DataOut(session model.Session) error {
	panic("implement me")
}
