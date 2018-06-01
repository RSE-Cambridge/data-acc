package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShowInstances(t *testing.T) {
	actual := getInstances().String()
	expected := `{"instances":[{"id":"fakebuffer","capacity":{"bytes":3,"nodes":40},"links":{"session":"fakebuffer"}}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}

func TestShowSessions(t *testing.T) {
	actual := getSessions().String()
	expected := `{"sessions":[{"id":"fakebuffer","created":12345678,"owner":1001,"token":"fakebuffer"}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}
