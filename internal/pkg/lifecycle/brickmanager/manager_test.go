package brickmanager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDevices(t *testing.T) {
	devices := getDevices()
	assert.Equal(t, 4, len(devices))
	assert.Equal(t, "nvme1n1", devices[0])
	assert.Equal(t, "nvme4n1", devices[3])
}
