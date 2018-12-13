package brickmanager

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
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

func TestGetBricks(t *testing.T) {
	devices := []string{"a", "b"}
	bricks := getBricks(devices, "host")

	assert.Equal(t, 2, len(bricks))
	assert.Equal(t, registry.BrickInfo{
		Device: "a", Hostname: "host", PoolName: "default", CapacityGB: 1400,
	}, bricks[0])
	assert.Equal(t, registry.BrickInfo{
		Device: "b", Hostname: "host", PoolName: "default", CapacityGB: 1400,
	}, bricks[1])
}