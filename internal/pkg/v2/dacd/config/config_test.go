package config

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetBrickManagerConfig(t *testing.T) {
	config := GetBrickManagerConfig(DefaultEnv)

	hostname, _ := os.Hostname()
	assert.Equal(t, datamodel.BrickHostName(hostname), config.BrickHostName)
	assert.Equal(t, uint(12), config.DeviceCount)
	assert.Equal(t, datamodel.PoolName("default"), config.PoolName)
	assert.Equal(t, true, config.HostEnabled)
	assert.Equal(t, "nvme%dn1", config.DeviceAddressPattern)
	assert.Equal(t, uint(1400), config.DeviceCapacityGiB)
}
