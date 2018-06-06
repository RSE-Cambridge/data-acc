package keystoreregistry

import (
	"bytes"
	"encoding/json"
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
	return fmt.Sprintf("/bricks/registered/%s/%s", hostname, device)
}

func getBrickAllocationKeyHost(allocation registry.BrickAllocation) string {
	return fmt.Sprintf("/bricks/allocated/host/%s/%s", allocation.Hostname, allocation.Device)
}
func getBrickAllocationKeyVolume(allocation registry.BrickAllocation) string {
	return fmt.Sprintf("/bricks/allocated/volume/%d/%s/%s",
		allocation.AllocatedIndex, allocation.Hostname, allocation.Device)
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
		// TODO: lots more error handing needed, like pool consistency, valid keys for etcd
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

func (poolRegistry *PoolRegistry) AllocateBricks(allocations []registry.BrickAllocation) error {
	var bricks []registry.BrickInfo
	var hostname string
	var raw []KeyValue
	for i, allocation := range allocations {
		brick, err := poolRegistry.GetBrickInfo(allocation.Hostname, allocation.Device)
		if err != nil {
			return fmt.Errorf("unable to find brick for: %s", allocation)
		}
		bricks = append(bricks, brick)

		if allocation.DeallocateRequested {
			return fmt.Errorf("should not requeste deallocated: %s", allocation)
		}
		if allocation.AllocatedIndex != 0 {
			return fmt.Errorf("should not specify the allocated index")
		}
		if hostname == "" {
			hostname = allocation.Hostname
		}
		if hostname != allocation.Hostname {
			return fmt.Errorf("all allocations must be for same host")
		}
		// TODO: this error checking suggest we specify the wrong format here!

		allocation.AllocatedIndex = uint(i)
		raw = append(raw, KeyValue{
			Key:   getBrickAllocationKeyHost(allocation),
			Value: toJson(allocation),
		})
		raw = append(raw, KeyValue{
			Key:   getBrickAllocationKeyVolume(allocation),
			Value: toJson(allocation),
		})
	}
	return poolRegistry.keystore.Add(raw)
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

func (poolRegistry *PoolRegistry) GetBrickInfo(hostname string, device string) (registry.BrickInfo, error) {
	raw, error := poolRegistry.keystore.Get(getBrickInfoKey(hostname, device))
	var value registry.BrickInfo
	json.Unmarshal(bytes.NewBufferString(raw.Value).Bytes(), &value)
	return value, error
}

func (*PoolRegistry) WatchHostBrickAllocations(hostname string,
	callback func(old registry.BrickAllocation, new registry.BrickAllocation)) {
	panic("implement me")
}
