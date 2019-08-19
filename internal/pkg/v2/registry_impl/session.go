package registry_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
)

func NewSessionRegistry(store store.Keystore) registry.SessionRegistry {
	return &sessionRegistry{store}
}

type sessionRegistry struct {
	store store.Keystore
}

func (s *sessionRegistry) GetSessionMutex(sessionName datamodel.SessionName) (store.Mutex, error) {
	return s.store.NewMutex(fmt.Sprintf("/session_lock/%s", sessionName))
}

func (s *sessionRegistry) CreateSession(session datamodel.Session) (datamodel.Session, error) {
	panic("implement me")
}

func (s *sessionRegistry) GetSession(sessionName datamodel.SessionName) (datamodel.Session, error) {
	panic("implement me")
}

func (s *sessionRegistry) GetAllSessions() ([]datamodel.Session, error) {
	panic("implement me")
}

func (s *sessionRegistry) UpdateSession(session datamodel.Session) (datamodel.Session, error) {
	panic("implement me")
}

func (s *sessionRegistry) DeleteSession(session datamodel.Session) error {
	panic("implement me")
}
