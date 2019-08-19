package registry_impl

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
)

func NewSessionActionsRegistry(store store.Keystore) registry.SessionActions {
	return &sessionActions{store}
}

type sessionActions struct {
	store store.Keystore
}

func (s *sessionActions) SendSessionAction(
	ctxt context.Context, actionType datamodel.SessionActionType,
	session datamodel.Session) (<-chan datamodel.SessionAction, error) {

	if session.PrimaryBrickHost == "" {
		panic("TODO: implement actions without assigned primary brick host")
	}
	panic("implement me")
}

func (s *sessionActions) GetSessionActions(
	ctxt context.Context, sessionName datamodel.SessionName) (<-chan datamodel.SessionAction, error) {
	panic("implement me")
}

func (s *sessionActions) CompleteSessionAction(action datamodel.SessionAction, err error) error {
	panic("implement me")
}
