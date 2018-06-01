package fakewarp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetInstances(t *testing.T) {
	actual := GetInstances().String()
	expected := `{"instances":[{"id":"fakebuffer","capacity":{"bytes":3,"nodes":40},"links":{"session":"fakebuffer"}}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}

func TestGetSessions(t *testing.T) {
	actual := GetSessions().String()
	expected := `{"sessions":[{"id":"fakebuffer","created":12345678,"owner":1001,"token":"fakebuffer"}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}
