package brick_manager_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionActionHandler_ProcessSessionAction_Unknown(t *testing.T) {
	action := datamodel.SessionAction{}
	handler := NewSessionActionHandler(nil)

	assert.PanicsWithValue(t,
		fmt.Sprintf("not yet implemented action for %+v", action),
		func() { handler.ProcessSessionAction(action) })
}
