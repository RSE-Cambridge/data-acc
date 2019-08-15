package actions_impl

import (
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
}
