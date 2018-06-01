package fakewarp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func assertNewline(t *testing.T, actual string) {
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}

func TestGetInstances(t *testing.T) {
	actual := GetInstances().String()
	expected := `{"instances":[{"id":"fakebuffer","capacity":{"bytes":3,"nodes":40},"links":{"session":"fakebuffer"}}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assertNewline(t, actual)
}

func TestGetSessions(t *testing.T) {
	actual := GetSessions().String()
	expected := `{"sessions":[{"id":"fakebuffer","created":1234567890,"owner":1001,"token":"fakebuffer"}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assertNewline(t, actual)
}

func TestGetConfigurations(t *testing.T) {
	actual := GetConfigurations().String()
	expected := `{"configurations":[]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assertNewline(t, actual)
}
