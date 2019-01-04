package brickmanager

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDevices(t *testing.T) {
	devices := getDevices("5", "")
	assert.Equal(t, 5, len(devices))
	assert.Equal(t, "nvme0n1", devices[0])
	assert.Equal(t, "nvme4n1", devices[4])

	devices = getDevices("asdf", "loop%d")
	assert.Equal(t, 12, len(devices))
	assert.Equal(t, "loop0", devices[0])
	assert.Equal(t, "loop11", devices[11])
}

func TestGetBricks(t *testing.T) {
	devices := []string{"a", "b"}
	bricks := getBricks(devices, "host", "-1", "")

	assert.Equal(t, 2, len(bricks))
	assert.Equal(t, registry.BrickInfo{
		Device: "a", Hostname: "host", PoolName: "default", CapacityGB: 1400,
	}, bricks[0])
	assert.Equal(t, registry.BrickInfo{
		Device: "b", Hostname: "host", PoolName: "default", CapacityGB: 1400,
	}, bricks[1])

	bricks = getBricks(devices, "host", "20", "foo")
	assert.Equal(t, registry.BrickInfo{
		Device: "b", Hostname: "host", PoolName: "foo", CapacityGB: 20,
	}, bricks[1])
}
