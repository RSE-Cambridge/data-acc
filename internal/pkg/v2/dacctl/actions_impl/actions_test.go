package actions_impl

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_workflow"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCliContext struct {
	strings map[string]string
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
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	session := mock_workflow.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().DeleteSession(datamodel.SessionName("bar"), true).Return(fakeError)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.DeleteBuffer(&mockCliContext{
		strings: map[string]string{"token": "bar"},
		booleans: map[string]bool{"hurry": true},
	})

	assert.Equal(t, fakeError, err)
}

func TestDacctlActions_DataIn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	session := mock_workflow.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().DataIn(datamodel.SessionName("bar")).Return(fakeError)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.DataIn(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)

	err = actions.DataIn(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: token", err.Error())
}

func TestDacctlActions_DataOut(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	session := mock_workflow.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().DataOut(datamodel.SessionName("bar")).Return(fakeError)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.DataOut(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)
}

func TestDacctlActions_PreRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	session := mock_workflow.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().AttachVolumes(datamodel.SessionName("bar"), nil, nil).Return(fakeError)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.PreRun(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)
}

func TestDacctlActions_PostRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockSessionRegistry(mockCtrl)
	session := mock_workflow.NewMockSession(mockCtrl)

	fakeError := errors.New("fake")
	session.EXPECT().DetachVolumes(datamodel.SessionName("bar")).Return(fakeError)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.PostRun(&mockCliContext{
		strings: map[string]string{"token": "bar"},
	})

	assert.Equal(t, fakeError, err)
}
