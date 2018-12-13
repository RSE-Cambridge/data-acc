package brickmanager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDevices(t *testing.T) {
	devices := getDevices("5")
	assert.Equal(t, 4, len(devices))
	assert.Equal(t, "nvme1n1", devices[0])
	assert.Equal(t, "nvme4n1", devices[3])

	devices = getDevices("asdf")
	assert.Equal(t, 11, len(devices))
	assert.Equal(t, "nvme1n1", devices[0])
	assert.Equal(t, "nvme11n1", devices[10])
}
