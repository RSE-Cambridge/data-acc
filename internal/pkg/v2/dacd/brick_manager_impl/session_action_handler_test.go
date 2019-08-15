package brick_manager_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestSessionActionHandler_ProcessSessionAction(t *testing.T) {
	action := datamodel.SessionAction{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionActions(mockCtrl)
	registry.EXPECT().CompleteSessionAction(action, nil)
	handler := NewSessionActionHandler(registry)

	// TODO...
	handler.ProcessSessionAction(action)
}
