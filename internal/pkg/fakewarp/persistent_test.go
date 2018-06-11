package fakewarp

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
)

type mockCliContext struct{}

func (c *mockCliContext) String(name string) string {
	return "bob"
}

func (c *mockCliContext) Int(name string) int {
	return 42 + len(name)
}

func TestCreatePersistentBufferReturnsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockKeystore(mockCtrl)
	mockCtxt := mockCliContext{}

	if _, error := CreatePersistentBuffer(&mockCtxt, mockObj); error != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(error))
	} else {
		t.Fatalf("Expected an error")
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
