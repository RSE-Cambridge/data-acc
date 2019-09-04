package brick_manager_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
)

func getDevices(brickManagerConfig config.BrickManagerConfig) []string {
	// TODO: should check these devices exist
	var bricks []string
	for i := 0; i < int(brickManagerConfig.DeviceCount); i++ {
		device := fmt.Sprintf(brickManagerConfig.DeviceAddressPattern, i)
		bricks = append(bricks, device)
	}
	return bricks
}

func getBrickHost(brickManagerConfig config.BrickManagerConfig) datamodel.BrickHost {
	var bricks []datamodel.Brick
	for _, device := range getDevices(brickManagerConfig) {
		bricks = append(bricks, datamodel.Brick{
			Device:        device,
			BrickHostName: brickManagerConfig.BrickHostName,
			PoolName:      brickManagerConfig.PoolName,
			CapacityGiB:   brickManagerConfig.DeviceCapacityGiB,
		})
	}

	return datamodel.BrickHost{
		Name:    brickManagerConfig.BrickHostName,
		Bricks:  bricks,
		Enabled: brickManagerConfig.HostEnabled,
	}
}
