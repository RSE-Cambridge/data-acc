package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShowInstances(t *testing.T) {
	var buf bytes.Buffer
	printInstances(&buf)
	actual := buf.String()
	expected := `{"instances":[{"id":"fakebuffer","capacity":{"bytes":3,"nodes":40},"links":{"session":"fakebuffer"}}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}

func TestShowSessions(t *testing.T) {
	var buf bytes.Buffer
	printSessions(&buf)
	actual := buf.String()
	expected := `{"sessions":[{"id":"fakebuffer","created":12345678,"owner":1001,"token":"fakebuffer"}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}
