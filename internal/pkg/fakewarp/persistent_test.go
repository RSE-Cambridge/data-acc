package fakewarp

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCliContext struct{}

func (c *mockCliContext) String(name string) string {
	switch name {
	case "capacity":
		return "pool1:42GiB"
	default:
		return ""
	}
}

func (c *mockCliContext) Int(name string) int {
	return 42 + len(name)
}

func TestCreatePersistentBufferReturnsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockVolumeRegistry(mockCtrl)
	mockObj.EXPECT().AddVolume(gomock.Any()) // TODO
	mockObj.EXPECT().AddJob(gomock.Any())
	mockPool := mocks.NewMockPoolRegistry(mockCtrl)
	mockCtxt := mockCliContext{}

	if actual, err := CreatePersistentBuffer(&mockCtxt, mockObj, mockPool); err != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(err))
	} else {
		assert.EqualValues(t, "", actual) // TODO
	}
}
