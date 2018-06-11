package fakewarp

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type mockCliContext struct{}

func (c *mockCliContext) String(name string) string {
	switch name {
	case "capacity":
		return "pool1:42GiB"
	default:
		return ""
	}
}

func (c *mockCliContext) Int(name string) int {
	return 42 + len(name)
}

func TestCreatePersistentBufferReturnsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockVolumeRegistry(mockCtrl)
	mockObj.EXPECT().AddVolume(gomock.Any()) // TODO
	mockObj.EXPECT().AddJob(gomock.Any())
	mockCtxt := mockCliContext{}

	if actual, err := CreatePersistentBuffer(&mockCtxt, mockObj); err != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(err))
	} else {
		assert.EqualValues(t, "", actual) // TODO
	}
}

func TestParseJobRequest(t *testing.T) {
	jobRequest := []string{
		`#DW asdf`,
		`foo`,
		`#DW swap`,
	}
	if err := parseJobRequest(jobRequest); err != nil {
		log.Fatal(err)
	}

}
