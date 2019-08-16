package brick_manager_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
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
		ActionType: datamodel.SessionCreate,
	}
	handler := sessionActionHandler{skipActions: true}

	handler.ProcessSessionAction(action)

	assert.Equal(t, datamodel.SessionCreate, handler.actionCalled)
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
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	handler := sessionActionHandler{actions: actions}
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionCreate,
	}

	actions.EXPECT().CompleteSessionAction(action, nil)

	handler.handleCreate(action)
}

func TestSessionActionHandler_handleDelete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	handler := sessionActionHandler{actions: actions}
	action := datamodel.SessionAction{
		ActionType: datamodel.SessionDelete,
	}

	actions.EXPECT().CompleteSessionAction(action, nil)

	handler.handleDelete(action)
}
