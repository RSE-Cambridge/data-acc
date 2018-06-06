package fakewarp

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCliContext struct{}

func (c *mockCliContext) String(name string) string {
	return "bob"
}

func (c *mockCliContext) Int(name string) int {
	return 42 + len(name)
}

func TestCreatePersistentBufferReturnsError(t *testing.T) {
	c := mockCliContext{}
	mockCtrl := gomock.NewController(t)
	mockObj := keystoreregistry.NewMockKeystore(mockCtrl)
	if _, error := CreatePersistentBuffer(&c, mockObj); error != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(error))
	} else {
		t.Fatalf("Expected an error")
	}
}
