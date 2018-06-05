package keystoreregistry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
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

func (*PoolRegistry) UpdateHost(bricks []registry.BrickInfo) error {
	panic("implement me")
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
