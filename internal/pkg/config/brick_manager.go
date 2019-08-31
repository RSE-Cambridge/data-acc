package config

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"log"
)

type BrickManagerConfig struct {
	BrickHostName        datamodel.BrickHostName
	PoolName             datamodel.PoolName
	DeviceCapacityGiB    uint
	DeviceCount          uint
	DeviceAddressPattern string
	HostEnabled          bool
}

// TODO: need additional validation here
func GetBrickManagerConfig(env ReadEnvironemnt) BrickManagerConfig {
	config := BrickManagerConfig{
		datamodel.BrickHostName(getHostname(env)),
		datamodel.PoolName(getString(env, "DAC_POOL_NAME", "default")),
		getUint(env, "DAC_BRICK_CAPACITY_GB",
			getUint(env, "DAC_DEVICE_CAPACITY_GB", 1400)),
		getUint(env, "DAC_BRICK_COUNT",
			getUint(env, "DEVICE_COUNT", 12)),
		getString(env, "DAC_BRICK_ADDRESS_PATTERN",
			getString(env, "DEVICE_TYPE", "nvme%dn1")),
		// Disabled means don't accept new Sessions, but allow Actions on existing Sessions
		getBool(env, "DAC_HOST_ENABLED", true),
	}
	log.Println("Got brick manager config:", config)
	return config
}
