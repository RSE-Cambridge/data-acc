package main

import (
	"testing"
	"bytes"
	"github.com/stretchr/testify/assert"
)

func TestPools(t *testing.T) {
	var buf bytes.Buffer
	printPools(&buf)
	actual := buf.String()
	expected := `{"pools":[{"id":"fake","units":"bytes","granularity":214748364800,"quantity":40,"free":3}]}`
	assert.EqualValues(t,expected, actual[:len(actual)-1])
	assert.EqualValues(t,"\n", actual[len(actual)-1:])
}
