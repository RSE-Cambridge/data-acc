package fakewarp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPools(t *testing.T) {
	actual := GetPools().String()
	expected := `{"pools":[{"id":"fake","units":"bytes","granularity":214748364800,"quantity":400,"free":395}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}
