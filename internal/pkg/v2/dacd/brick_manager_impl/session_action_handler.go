package brick_manager_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/workflow"
)

func NewSessionActionHandler() workflow.SessionActionHandler {
	return &sessionActionHandler{}
}

type sessionActionHandler struct {}

func (*sessionActionHandler) ProcessSessionAction(action datamodel.SessionAction) {
	panic("implement me")
}
