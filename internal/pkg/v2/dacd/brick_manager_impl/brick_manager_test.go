package brick_manager_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBrickManager_Hostname(t *testing.T) {
	brickManager := NewBrickManager(nil)
	assert.Equal(t, getHostname(), brickManager.Hostname())
}

func TestBrickManager_Startup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	brickRegistry := mock_registry.NewMockBrickRegistry(mockCtrl)
	brickManager := NewBrickManager(brickRegistry)

	err := brickManager.Startup(false)

	assert.Nil(t, err)
}