package actionsImpl

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/mock_session"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCliContext struct {
	capacity int
}

func (c *mockCliContext) String(name string) string {
	switch name {
	case "capacity":
		return fmt.Sprintf("pool1:%dGiB", c.capacity)
	case "token":
		return "token"
	case "caller":
		return "caller"
	case "user":
		return "user"
	case "access":
		return "access"
	case "type":
		return "type"
	case "job":
		return "jobfile"
	case "nodehostnamefile":
		return "nodehostnamefile1"
	case "pathfile":
		return "pathfile1"
	default:
		return "foobar1"
	}
}

func (c *mockCliContext) Int(name string) int {
	switch name {
	case "user":
		return 1001
	case "group":
		return 1002
	default:
		return 42 + len(name)
	}
}

func TestDacctlActions_CreatePersistentBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_session.NewMockRegistry(mockCtrl)
	session := mock_session.NewMockActions(mockCtrl)

	fakeSession := datamodel.Session{Name: "foo"}
	registry.EXPECT().CreateSessionAllocations(datamodel.Session{
		Name:      "token",
		Owner:     1001,
		Group:     1002,
		CreatedAt: 123,
		PersistentVolumeRequest: datamodel.PersistentVolumeRequest{
			Caller:        "caller",
			PoolName:      "pool1",
			CapacityBytes: 2147483648,
		},
	}).Return(fakeSession, nil)
	session.EXPECT().CreateSessionVolume(fakeSession)
	fakeTime = 123

	actions := NewDacctlActions(registry, session, nil)
	err := actions.CreatePersistentBuffer(&mockCliContext{capacity: 2})

	assert.Nil(t, err)
}

func TestDacctlActions_DeleteBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_session.NewMockRegistry(mockCtrl)
	session := mock_session.NewMockActions(mockCtrl)

	fakeSession := datamodel.Session{Name: "foo"}
	registry.EXPECT().GetSession("token").Return(fakeSession, nil)
	fakeError := errors.New("fake")
	session.EXPECT().DeleteSession(fakeSession).Return(fakeError)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.DeleteBuffer(&mockCliContext{})

	assert.Equal(t, fakeError, err)
}