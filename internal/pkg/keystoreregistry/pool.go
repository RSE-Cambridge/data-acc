package keystoreregistry

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"strings"
)

func NewPoolRegistry(keystore Keystore) registry.PoolRegistry {
	return &PoolRegistry{keystore}
}

type PoolRegistry struct {
	keystore Keystore
}

func (*PoolRegistry) Pools() ([]registry.Pool, error) {
	panic("implement me")
}

func getBrickInfoKey(hostname string, device string) string {
	return fmt.Sprintf("/bricks/%s/%s", hostname, device)
}

func (poolRegistry *PoolRegistry) UpdateHost(bricks []registry.BrickInfo) error {
	var values []KeyValueVersion
	var problems []string
	var hostname string
	for _, brickInfo := range bricks {
		if hostname == "" {
			hostname = brickInfo.Hostname
		}
		if hostname != brickInfo.Hostname {
			problems = append(problems, "Only one host to be updated at once")
		}
		// TODO: lots more error handing needed, like pool consistency
		values = append(values, KeyValueVersion{
			Key:   getBrickInfoKey(brickInfo.Hostname, brickInfo.Device),
			Value: toJson(brickInfo),
		})
	}
	if len(problems) > 0 {
		return fmt.Errorf("can't update host because: %s", strings.Join(problems, ", "))
	}
	return poolRegistry.keystore.Update(values)
}

func (*PoolRegistry) KeepAliveHost(hostname string) error {
	panic("implement me")
}

func (*PoolRegistry) AllocateBricks(allocations []registry.BrickAllocation) error {
	panic("implement me")
}

func (*PoolRegistry) DeallocateBrick(allocations []registry.BrickAllocation) error {
	panic("implement me")
}

func (*PoolRegistry) GetAllocationsForHost(hostname string) ([]registry.BrickAllocation, error) {
	panic("implement me")
}

func (*PoolRegistry) GetAllocationsForVolume(volume registry.VolumeName) ([]registry.BrickAllocation, error) {
	panic("implement me")
}

func (*PoolRegistry) GetBrickInfo(hostname string, device string) (registry.BrickInfo, error) {
	panic("implement me")
}

func (*PoolRegistry) WatchHostBrickAllocations(hostname string,
	callback func(old registry.BrickAllocation, new registry.BrickAllocation)) {
	panic("implement me")
}
