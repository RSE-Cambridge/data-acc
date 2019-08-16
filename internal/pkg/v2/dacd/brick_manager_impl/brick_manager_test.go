package brick_manager_impl

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacd/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBrickManager_Hostname(t *testing.T) {
	brickManager := brickManager{config: config.BrickManagerConfig{BrickHostName: "host"}}
	assert.Equal(t, "host", brickManager.Hostname())
}

func TestBrickManager_Startup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	brickRegistry := mock_registry.NewMockBrickRegistry(mockCtrl)
	handler := mock_facade.NewMockSessionActionHandler(mockCtrl)
	brickManager := NewBrickManager(brickRegistry, handler)

	// TODO...
	brickRegistry.EXPECT().UpdateBrickHost(gomock.Any())
	brickRegistry.EXPECT().GetSessionActions(context.TODO(), gomock.Any())
	brickRegistry.EXPECT().KeepAliveHost(context.TODO(), datamodel.BrickHostName("hostname"))

	err := brickManager.Startup(false)

	assert.Nil(t, err)
}
