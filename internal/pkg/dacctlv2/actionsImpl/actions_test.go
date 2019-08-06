package actionsImpl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/mock_session"
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
		return fmt.Sprintf("pool1:%dGB", c.capacity)
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
		return ""
	}
}

func (c *mockCliContext) Int(name string) int {
	switch name {
	case "user":
		return 1001
	case "group":
		return 1001
	default:
		return 42 + len(name)
	}
}

func TestDacctlActions_CreatePersistentBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_session.NewMockRegistry(mockCtrl)
	session := mock_session.NewMockActions(mockCtrl)

	actions := NewDacctlActions(registry, session, nil)
	err := actions.CreatePersistentBuffer(&mockCliContext{})

	assert.Nil(t, err)
}
