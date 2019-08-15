package actions_impl

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_workflow"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDacctlActions_RealSize(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_workflow.NewMockSession(mockCtrl)
	session.EXPECT().GetSession(datamodel.SessionName("bar")).Return(datamodel.Session{
		Name:            datamodel.SessionName("bar"),
		ActualSizeBytes: 123,
	}, nil)

	actions := NewDacctlActions(session, nil)
	output, err := actions.RealSize(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Nil(t, err)
	assert.Equal(t, `{"token":"bar", "capacity":123, "units":"bytes"}`, output)

	_, err = actions.RealSize(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())
}

func TestDacctlActions_Paths(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_workflow.NewMockSession(mockCtrl)
	disk := mocks.NewMockDisk(mockCtrl)

	session.EXPECT().GetSession(datamodel.SessionName("bar")).Return(datamodel.Session{
		Name: datamodel.SessionName("bar"),
		Paths: map[string]string{
			"foo1": "bar1",
		},
	}, nil)
	disk.EXPECT().Write("paths", []string{"foo1=bar1"})

	actions := NewDacctlActions(session, disk)
	err := actions.Paths(&mockCliContext{
		strings: map[string]string{
			"token":    "bar",
			"pathfile": "paths",
		},
	})

	assert.Nil(t, err)

	err = actions.Paths(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token, pathfile", err.Error())

	fakeError := errors.New("fake")
	session.EXPECT().GetSession(datamodel.SessionName("bar")).Return(datamodel.Session{}, fakeError)
	err = actions.Paths(&mockCliContext{
		strings: map[string]string{
			"token":    "bar",
			"pathfile": "paths",
		},
	})
	assert.Equal(t, fakeError, err)
}

func TestDacctlActions_ShowInstances(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_workflow.NewMockSession(mockCtrl)
	session.EXPECT().GetAllSessions().Return([]datamodel.Session{
		{
			Name:            datamodel.SessionName("foo"),
			ActualSizeBytes: 123,
		},
		{
			Name:            datamodel.SessionName("bar"),
			ActualSizeBytes: 456,
		},
	}, nil)

	actions := NewDacctlActions(session, nil)
	output, err := actions.ShowInstances()
	assert.Nil(t, err)
	expected := `{"instances":[{"id":"foo","capacity":{"bytes":123,"nodes":0},"links":{"session":"foo"}},{"id":"bar","capacity":{"bytes":456,"nodes":0},"links":{"session":"bar"}}]}`
	assert.Equal(t, expected, output)

	fakeErr := errors.New("fake")
	session.EXPECT().GetAllSessions().Return(nil, fakeErr)
	output, err = actions.ShowInstances()
	assert.Equal(t, "", output)
	assert.Equal(t, fakeErr, err)
}
