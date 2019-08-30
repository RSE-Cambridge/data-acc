package brick_manager_impl

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBrickManager_Hostname(t *testing.T) {
	brickManager := brickManager{config: config.BrickManagerConfig{BrickHostName: "host"}}
	assert.Equal(t, "host", brickManager.Hostname())
}

func TestBrickManager_Startup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	brickRegistry := mock_registry.NewMockBrickHostRegistry(mockCtrl)
	sessionActions := mock_registry.NewMockSessionActions(mockCtrl)
	handler := mock_facade.NewMockSessionActionHandler(mockCtrl)
	brickManager := brickManager{
		config:        config.GetBrickManagerConfig(config.DefaultEnv),
		brickRegistry: brickRegistry, sessionActions: sessionActions, sessionActionHandler: handler,
	}

	// TODO...
	brickRegistry.EXPECT().UpdateBrickHost(gomock.Any())
	sessionActions.EXPECT().GetSessionActionRequests(context.TODO(), gomock.Any())
	hostname, _ := os.Hostname()
	brickRegistry.EXPECT().KeepAliveHost(context.TODO(), datamodel.BrickHostName(hostname))

	brickManager.Startup()
}
