package brick_manager_impl

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_filesystem"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_store"
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

func TestSessionActionHandler_handleCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	fsProvider := mock_filesystem.NewMockProvider(mockCtrl)
	handler := sessionActionHandler{
		sessionRegistry: registry, actions: actions, fsProvider: fsProvider,
	}
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionCreateFilesystem,
		Session:    datamodel.Session{Name: "test", ActualSizeBytes: 42},
	}
	sessionMutex := mock_store.NewMockMutex(mockCtrl)
	registry.EXPECT().GetSessionMutex(action.Session.Name).Return(sessionMutex, nil)
	sessionMutex.EXPECT().Lock(context.TODO())
	sessionMutex.EXPECT().Unlock(context.TODO())
	registry.EXPECT().GetSession(action.Session.Name).Return(action.Session, nil)
	fsProvider.EXPECT().Create(action.Session)
	updatedSession := datamodel.Session{
		Name:            action.Session.Name,
		Status:          datamodel.SessionStatus{FileSystemCreated: true},
		ActualSizeBytes: 42,
	}
	registry.EXPECT().UpdateSession(updatedSession).Return(updatedSession, nil)
	updatedAction := datamodel.SessionAction{
		ActionType: datamodel.SessionCreateFilesystem,
		Session:    updatedSession,
	}
	actions.EXPECT().CompleteSessionAction(updatedAction)

	handler.handleCreate(action)
}
