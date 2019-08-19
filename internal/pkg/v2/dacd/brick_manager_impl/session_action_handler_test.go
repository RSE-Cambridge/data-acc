package brick_manager_impl

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_filesystem"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionActionHandler_ProcessSessionAction_Unknown(t *testing.T) {
	action := datamodel.SessionAction{}
	handler := NewSessionActionHandler(nil)

	assert.PanicsWithValue(t,
		fmt.Sprintf("not yet implemented action for %+v", action),
		func() { handler.ProcessSessionAction(action) })
}

func TestSessionActionHandler_ProcessSessionAction_Create(t *testing.T) {
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionCreateFilesystem,
	}
	handler := sessionActionHandler{skipActions: true}

	handler.ProcessSessionAction(action)

	assert.Equal(t, datamodel.SessionCreateFilesystem, handler.actionCalled)
}

func TestSessionActionHandler_ProcessSessionAction_Delete(t *testing.T) {
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionDelete,
	}
	handler := sessionActionHandler{skipActions: true}

	handler.ProcessSessionAction(action)

	assert.Equal(t, datamodel.SessionDelete, handler.actionCalled)
}

func TestSessionActionHandler_handleCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	fsProvider := mock_filesystem.NewMockProvider(mockCtrl)
	handler := sessionActionHandler{
		registry: registry, actions: actions, fsProvider: fsProvider,
	}
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionCreateFilesystem,
		Session:    datamodel.Session{Name: "test"},
	}
	sessionMutex := mock_store.NewMockMutex(mockCtrl)
	registry.EXPECT().GetSessionMutex(action.Session.Name).Return(sessionMutex, nil)
	sessionMutex.EXPECT().Lock(context.TODO())
	sessionMutex.EXPECT().Unlock(context.TODO())
	registry.EXPECT().GetSession(action.Session.Name).Return(action.Session, nil)
	fsProvider.EXPECT().Create(action.Session)
	updatedSession := datamodel.Session{
		Name:   action.Session.Name,
		Status: datamodel.SessionStatus{FileSystemCreated: true},
	}
	registry.EXPECT().UpdateSession(updatedSession).Return(updatedSession, nil)
	updatedAction := datamodel.SessionAction{
		ActionType: datamodel.SessionCreateFilesystem,
		Session:    updatedSession,
	}
	actions.EXPECT().CompleteSessionAction(updatedAction, nil)

	handler.handleCreate(action)
}

func TestSessionActionHandler_handleDelete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	handler := sessionActionHandler{actions: actions, registry: registry}
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionDelete,
	}
	// TODO: need to pass session better? who deletes allocations?
	registry.EXPECT().DeleteSession(action.Session)
	actions.EXPECT().CompleteSessionAction(action, nil)

	handler.handleDelete(action)
}
