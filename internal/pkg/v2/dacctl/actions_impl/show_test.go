package actions_impl

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_fileio"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDacctlActions_RealSize(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	session.EXPECT().GetSession(datamodel.SessionName("bar")).Return(datamodel.Session{
		Name:            datamodel.SessionName("bar"),
		ActualSizeBytes: 123,
	}, nil)

	actions := dacctlActions{session: session}
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
	session := mock_facade.NewMockSession(mockCtrl)
	disk := mock_fileio.NewMockDisk(mockCtrl)

	session.EXPECT().GetSession(datamodel.SessionName("bar")).Return(datamodel.Session{
		Name: datamodel.SessionName("bar"),
		Paths: map[string]string{
			"foo1": "bar1",
		},
	}, nil)
	disk.EXPECT().Write("paths", []string{"foo1=bar1"})

	actions := dacctlActions{session: session, disk: disk}
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
	session := mock_facade.NewMockSession(mockCtrl)
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
	actions := dacctlActions{session: session}

	output, err := actions.ShowInstances()

	assert.Nil(t, err)
	expected := `{"instances":[{"id":"foo","capacity":{"bytes":123,"nodes":0},"links":{"session":"foo"}},{"id":"bar","capacity":{"bytes":456,"nodes":0},"links":{"session":"bar"}}]}`
	assert.Equal(t, expected, output)

	fakeErr := errors.New("fake")
	session.EXPECT().GetAllSessions().Return(nil, fakeErr)
	output, err = actions.ShowInstances()
	assert.Equal(t, "", output)
	assert.Equal(t, fakeErr, err)

	session.EXPECT().GetAllSessions().Return(nil, nil)
	output, err = actions.ShowInstances()
	assert.Nil(t, err)
	expected = `{"instances":[]}`
	assert.Equal(t, expected, output)
}

func TestDacctlActions_ShowSessions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	session.EXPECT().GetAllSessions().Return([]datamodel.Session{
		{
			Name:      datamodel.SessionName("foo"),
			Owner:     42,
			CreatedAt: 1234,
		},
		{
			Name:      datamodel.SessionName("bar"),
			Owner:     43,
			CreatedAt: 5678,
		},
	}, nil)
	actions := dacctlActions{session: session}

	output, err := actions.ShowSessions()

	assert.Nil(t, err)
	expected := `{"sessions":[{"id":"foo","created":1234,"owner":42,"token":"foo"},{"id":"bar","created":5678,"owner":43,"token":"bar"}]}`
	assert.Equal(t, expected, output)

	fakeErr := errors.New("fake")
	session.EXPECT().GetAllSessions().Return(nil, fakeErr)
	output, err = actions.ShowSessions()
	assert.Equal(t, "", output)
	assert.Equal(t, fakeErr, err)

	session.EXPECT().GetAllSessions().Return(nil, nil)
	output, err = actions.ShowSessions()
	assert.Nil(t, err)
	expected = `{"sessions":[]}`
	assert.Equal(t, expected, output)
}

func TestDacctlActions_ListPools(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	session.EXPECT().GetPools().Return([]datamodel.PoolInfo{
		{
			Pool: datamodel.Pool{
				Name:             "default",
				GranularityBytes: 1024,
			},
			AllocatedBricks: []datamodel.BrickAllocation{
				{
					Brick: datamodel.Brick{Device: "sda"},
				},
			},
			AvailableBricks: []datamodel.Brick{
				{Device: "sdb"},
				{Device: "sdc"},
			},
		},
	}, nil)
	actions := dacctlActions{session: session}

	output, err := actions.ListPools()
	assert.Nil(t, err)
	expexted := `{"pools":[{"id":"default","units":"bytes","granularity":1024,"quantity":3,"free":2}]}`
	assert.Equal(t, expexted, output)

	session.EXPECT().GetPools().Return(nil, nil)
	output, err = actions.ListPools()
	assert.Nil(t, err)
	assert.Equal(t, `{"pools":[]}`, output)

	fakeErr := errors.New("fake")
	session.EXPECT().GetPools().Return(nil, fakeErr)
	output, err = actions.ListPools()
	assert.Equal(t, fakeErr, err)
	assert.Equal(t, "", output)
}

func TestDacctlActions_ShowConfigurations(t *testing.T) {
	actions := dacctlActions{}
	output, err := actions.ShowConfigurations()
	assert.Nil(t, err)
	assert.Equal(t, `{"configurations":[]}`, output)
}
