package actions_impl

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_facade"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCliContext struct {
	strings  map[string]string
	integers map[string]int
	booleans map[string]bool
}

func (c *mockCliContext) String(name string) string {
	return c.strings[name]
}

func (c *mockCliContext) Int(name string) int {
	return c.integers[name]
}

func (c *mockCliContext) Bool(name string) bool {
	return c.booleans[name]
}

func getMockCliContext(capacity int) *mockCliContext {
	ctxt := mockCliContext{}
	ctxt.strings = map[string]string{
		"capacity":         fmt.Sprintf("pool1:%dGiB", capacity),
		"token":            "token",
		"caller":           "caller",
		"access":           "asdf",
		"type":             "type",
		"job":              "jobfile",
		"nodehostnamefile": "nodehostnamefile1",
		"pathfile":         "pathfile1",
	}
	ctxt.integers = map[string]int{
		"user":  1001,
		"group": 1002,
	}
	return &ctxt
}

func TestDacctlActions_DeleteBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().DeleteSession(datamodel.SessionName("bar"), true).Return(fakeError)

	actions := NewDacctlActions(session, nil)
	err := actions.DeleteBuffer(&mockCliContext{
		strings:  map[string]string{"token": "bar"},
		booleans: map[string]bool{"hurry": true},
	})

	assert.Equal(t, fakeError, err)

	err = actions.DeleteBuffer(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())
}

func TestDacctlActions_DataIn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().CopyDataIn(datamodel.SessionName("bar")).Return(fakeError)

	actions := NewDacctlActions(session, nil)
	err := actions.DataIn(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)

	err = actions.DataIn(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())

	err = actions.DataIn(&mockCliContext{
		strings: map[string]string{"token": "bad token"},
	})
	assert.Equal(t, "badly formatted session name: bad token", err.Error())
}

func TestDacctlActions_DataOut(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().CopyDataOut(datamodel.SessionName("bar")).Return(fakeError)

	actions := NewDacctlActions(session, nil)
	err := actions.DataOut(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)

	err = actions.DataOut(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())
}

func TestDacctlActions_PreRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	disk := mocks.NewMockDisk(mockCtrl)

	computeHosts := []string{"host1", "host2"}
	loginHosts := []string{"login"}
	disk.EXPECT().Lines("computehostfile").Return(computeHosts, nil)
	disk.EXPECT().Lines("loginhostfile").Return(loginHosts, nil)
	fakeError := errors.New("fake")
	session.EXPECT().Mount(datamodel.SessionName("bar"), computeHosts, loginHosts).Return(fakeError)

	actions := NewDacctlActions(session, disk)
	err := actions.PreRun(&mockCliContext{
		strings: map[string]string{
			"token":                "bar",
			"nodehostnamefile":     "computehostfile",
			"jobexecutionnodefile": "loginhostfile",
		},
	})

	assert.Equal(t, fakeError, err)

	err = actions.PreRun(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())

	err = actions.PreRun(&mockCliContext{strings: map[string]string{"token": "bar"}})
	assert.Equal(t, "Please provide these required parameters: nodehostnamefile", err.Error())
}

func TestDacctlActions_PreRun_NoLoginHosts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	disk := mocks.NewMockDisk(mockCtrl)

	computeHosts := []string{"host1", "host2"}
	disk.EXPECT().Lines("computehostfile").Return(computeHosts, nil)
	fakeError := errors.New("fake")
	session.EXPECT().Mount(datamodel.SessionName("bar"), computeHosts, nil).Return(fakeError)

	actions := NewDacctlActions(session, disk)
	err := actions.PreRun(&mockCliContext{
		strings: map[string]string{
			"token":            "bar",
			"nodehostnamefile": "computehostfile",
		},
	})

	assert.Equal(t, fakeError, err)
}

func TestDacctlActions_PreRun_BadHosts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	disk := mocks.NewMockDisk(mockCtrl)

	computeHosts := []string{"host1", "host/2"}
	disk.EXPECT().Lines("computehostfile").Return(computeHosts, nil)

	actions := NewDacctlActions(nil, disk)
	err := actions.PreRun(&mockCliContext{
		strings: map[string]string{
			"token":            "bar",
			"nodehostnamefile": "computehostfile",
		},
	})

	assert.Equal(t, "invalid hostname in: [host/2]", err.Error())
}

func TestDacctlActions_PreRun_BadLoginHosts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	disk := mocks.NewMockDisk(mockCtrl)

	computeHosts := []string{"host1", "host2"}
	loginHosts := []string{"login/asdf"}
	disk.EXPECT().Lines("computehostfile").Return(computeHosts, nil)
	disk.EXPECT().Lines("loginhostfile").Return(loginHosts, nil)

	actions := NewDacctlActions(session, disk)
	err := actions.PreRun(&mockCliContext{
		strings: map[string]string{
			"token":                "bar",
			"nodehostnamefile":     "computehostfile",
			"jobexecutionnodefile": "loginhostfile",
		},
	})

	assert.Equal(t, "invalid hostname in: [login/asdf]", err.Error())
}

func TestDacctlActions_PreRun_NoHosts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	disk := mocks.NewMockDisk(mockCtrl)

	disk.EXPECT().Lines("computehostfile").Return(nil, nil)

	actions := NewDacctlActions(nil, disk)
	err := actions.PreRun(&mockCliContext{
		strings: map[string]string{
			"token":            "bar",
			"nodehostnamefile": "computehostfile",
		},
	})

	assert.Equal(t, "unable to mount to zero compute hosts", err.Error())
}

func TestDacctlActions_PostRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().Unmount(datamodel.SessionName("bar")).Return(fakeError)

	actions := NewDacctlActions(session, nil)
	err := actions.PostRun(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)

	err = actions.PostRun(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())
}
