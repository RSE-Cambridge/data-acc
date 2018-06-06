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

func getPrefixAllocationHost(hostname string) string {
	return fmt.Sprintf("/bricks/allocated/host/%s/", hostname)
}
func getBrickAllocationKeyHost(allocation registry.BrickAllocation) string {
	prefix := getPrefixAllocationHost(allocation.Hostname)
	return fmt.Sprintf("%s%s", prefix, allocation.Device)
}

func getPrefixAllocationVolume(volume registry.VolumeName) string {
	return fmt.Sprintf("/bricks/allocated/volume/%s/", volume)
}
func getBrickAllocationKeyVolume(allocation registry.BrickAllocation) string {
	prefix := getPrefixAllocationVolume(allocation.AllocatedVolume)
	return fmt.Sprintf("%s%d/%s/%s",
		prefix, allocation.AllocatedIndex, allocation.Hostname, allocation.Device)
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
	var volume registry.VolumeName
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
		if volume == "" {
			volume = allocation.AllocatedVolume
		}
		if volume != allocation.AllocatedVolume {
			return fmt.Errorf("all allocations must be for same volume")
		}
		// TODO: this error checking suggest we specify the wrong format here!

		allocation.AllocatedIndex = uint(i)
		raw = append(raw, KeyValue{
			Key:   getBrickAllocationKeyHost(allocation),
			Value: toJson(allocation),
		})
		// TODO: maybe just point to the other key?? this duplication is terrible
		raw = append(raw, KeyValue{
			Key:   getBrickAllocationKeyVolume(allocation),
			Value: toJson(allocation),
		})
	}
	return poolRegistry.keystore.Add(raw)
}

func (poolRegistry *PoolRegistry) deallocate(raw []KeyValueVersion,
	updated []KeyValueVersion) ([]KeyValueVersion, []string) {
	var keys []string
	for _, entry := range raw {
		rawValue := entry.Value
		var allocation registry.BrickAllocation
		json.Unmarshal(bytes.NewBufferString(rawValue).Bytes(), &allocation)

		allocation.DeallocateRequested = true
		entry.Value = toJson(&allocation)
		updated = append(updated, entry)
		keys = append(keys, getBrickAllocationKeyHost(allocation))
	}
	return updated, keys
}

func (poolRegistry *PoolRegistry) DeallocateBricks(volume registry.VolumeName) error {
	var updated []KeyValueVersion

	volPrefix := getPrefixAllocationVolume(volume)
	raw, err := poolRegistry.keystore.GetAll(volPrefix)
	if err != nil {
		return nil
	}
	updated, keys := poolRegistry.deallocate(raw, updated)

	raw = []KeyValueVersion{}
	for _, key := range keys {
		entry, err := poolRegistry.keystore.Get(key)
		if err != nil {
			return err
		}
		raw = append(raw, entry)
	}
	updated, _ = poolRegistry.deallocate(raw, updated)
	return poolRegistry.keystore.Update(updated)
}

func (poolRegistry *PoolRegistry) getAllocations(prefix string) ([]registry.BrickAllocation, error) {
	raw, err := poolRegistry.keystore.GetAll(prefix)
	if err != nil {
		return nil, err
	}
	var allocations []registry.BrickAllocation
	for _, entry := range raw {
		rawValue := entry.Value
		var allocation registry.BrickAllocation
		json.Unmarshal(bytes.NewBufferString(rawValue).Bytes(), &allocation)
		allocations = append(allocations, allocation)
	}
	return allocations, nil
}
func (poolRegistry *PoolRegistry) GetAllocationsForHost(hostname string) ([]registry.BrickAllocation, error) {
	return poolRegistry.getAllocations(getPrefixAllocationHost(hostname))
}

func (poolRegistry *PoolRegistry) GetAllocationsForVolume(volume registry.VolumeName) ([]registry.BrickAllocation, error) {
	return poolRegistry.getAllocations(getPrefixAllocationVolume(volume))
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
