package registry_impl

import (
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl/actions_impl/parsers"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"log"
)

func NewSessionRegistry(store store.Keystore) registry.SessionRegistry {
	return &sessionRegistry{store}
}

type sessionRegistry struct {
	store store.Keystore
}

func (s *sessionRegistry) GetSessionMutex(sessionName datamodel.SessionName) (store.Mutex, error) {
	sessionKey := getSessionKey(sessionName)
	lockKey := fmt.Sprintf("/lock%s", sessionKey)
	return s.store.NewMutex(lockKey)
}

const sessionPrefix = "/session/"

func getSessionKey(sessionName datamodel.SessionName) string {
	if !parsers.IsValidName(string(sessionName)) {
		log.Panicf("invalid session name: '%s'", sessionName)
	}
	return fmt.Sprintf("%s%s", sessionPrefix, sessionName)
}

func (s *sessionRegistry) CreateSession(session datamodel.Session) (datamodel.Session, error) {
	sessionKey := getSessionKey(session.Name)

	// TODO: more validation?
	if session.ActualSizeBytes > 0 {
		if len(session.Allocations) == 0 {
			return session, fmt.Errorf("session must have allocations before being created")
		}
		if session.PrimaryBrickHost == "" {
			return session, fmt.Errorf("session must have a primary brick host set")
		}
	} else {
		if len(session.Allocations) != 0 {
			return session, fmt.Errorf("allocations out of sync with ActualSizeBytes")
		}
		if session.PrimaryBrickHost != "" {
			return session, fmt.Errorf("PrimaryBrickHost should be empty if no bricks assigned")
		}
	}

	sessionAsString, err := json.Marshal(session)
	if err != nil {
		return session, fmt.Errorf("unable to convert session to json due to: %s", err)
	}

	keyValueVersion, err := s.store.Create(sessionKey, sessionAsString)
	if err != nil {
		return session, fmt.Errorf("unable to create session due to: %s", err)
	}

	// Return the last modification revision
	session.Revision = keyValueVersion.ModRevision
	return session, nil
}

func (s *sessionRegistry) GetSession(sessionName datamodel.SessionName) (datamodel.Session, error) {
	sessionKey := getSessionKey(sessionName)

	keyValueVersion, err := s.store.Get(sessionKey)
	if err != nil {
		return datamodel.Session{}, fmt.Errorf("unable to get session due to: %s", err)
	}

	session := datamodel.Session{}
	err = json.Unmarshal(keyValueVersion.Value, &session)
	if err != nil {
		return datamodel.Session{}, fmt.Errorf("unable parse session from store due to: %s", err)
	}

	session.Revision = keyValueVersion.ModRevision
	return session, nil
}

func (s *sessionRegistry) GetAllSessions() ([]datamodel.Session, error) {
	results, err := s.store.GetAll(sessionPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to get all sessions due to: %s", err.Error())
	}

	var sessions []datamodel.Session
	for _, keyValueVersion := range results {
		session := datamodel.Session{}
		err = json.Unmarshal(keyValueVersion.Value, &session)
		if err != nil {
			log.Panicf("unable parse session from store due to: %s", err)
		}
		session.Revision = keyValueVersion.ModRevision
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *sessionRegistry) UpdateSession(session datamodel.Session) (datamodel.Session, error) {
	sessionKey := getSessionKey(session.Name)

	keyValueVersion, err := s.store.Update(sessionKey, sessionToRaw(session), session.Revision)
	if err != nil {
		return session, fmt.Errorf("unable to update session due to: %s", err.Error())
	}

	newSession := sessionFromRaw(keyValueVersion.Value)
	newSession.Revision = keyValueVersion.ModRevision
	return newSession, nil
}

func (s *sessionRegistry) DeleteSession(session datamodel.Session) error {
	sessionKey := getSessionKey(session.Name)
	return s.store.Delete(sessionKey, session.Revision)
}

func sessionToRaw(session datamodel.Session) []byte {
	rawSession, err := json.Marshal(session)
	if err != nil {
		log.Panicf("unable to convert session to json due to: %s", err.Error())
	}
	return rawSession
}

func sessionFromRaw(raw []byte) datamodel.Session {
	session := datamodel.Session{}
	err := json.Unmarshal(raw, &session)
	if err != nil {
		log.Panicf("unable parse session from store due to: %s", err)
	}
	return session
}