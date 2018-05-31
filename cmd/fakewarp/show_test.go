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
	assert.EqualValues(t,expected, actual[:len(actual)-1])
	assert.EqualValues(t,"\n", actual[len(actual)-1:])
}
