package actions_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_workflow"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDacctlActions_CreatePersistentBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_workflow.NewMockSession(mockCtrl)

	session.EXPECT().CreateSessionVolume(datamodel.Session{
		Name:      "token",
		Owner:     1001,
		Group:     1002,
		CreatedAt: 123,
		VolumeRequest: datamodel.VolumeRequest{
			MultiJob:           true,
			Caller:             "caller",
			PoolName:           "pool1",
			TotalCapacityBytes: 2147483648,
		},
	}).Return(nil)
	fakeTime = 123

	actions := NewDacctlActions(session, nil)
	err := actions.CreatePersistentBuffer(getMockCliContext(2))

	assert.Nil(t, err)
}
