package fakewarp

import (
	"fmt"
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
	if _, error := CreatePersistentBuffer(&c); error != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(error))
	} else {
		t.Fatalf("Expected an error")
	}
}
