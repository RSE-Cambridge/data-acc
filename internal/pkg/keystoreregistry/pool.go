package keystoreregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"strings"
)

func NewPoolRegistry(keystore Keystore) registry.PoolRegistry {
	return &poolRegistry{keystore}
}

type poolRegistry struct {
	keystore Keystore
}

const registeredBricksPrefix = "/bricks/registered/"

func getBrickInfoKey(hostname string, device string) string {
	return fmt.Sprintf("%s%s/%s/", registeredBricksPrefix, hostname, device)
}

const allocatedBricksPrefix = "/bricks/allocated/host/"

func getPrefixAllocationHost(hostname string) string {
	return fmt.Sprintf("%s%s/", allocatedBricksPrefix, hostname)
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

func (poolRegistry *poolRegistry) UpdateHost(bricks []registry.BrickInfo) error {
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

func getKeepAliveKey(hostname string) string {
	return fmt.Sprintf("/host/keepalive/%s", hostname)
}

func (poolRegistry *poolRegistry) KeepAliveHost(hostname string) error {
	return poolRegistry.keystore.KeepAliveKey(getKeepAliveKey(hostname))
}

func (poolRegistry *poolRegistry) HostAlive(hostname string) (bool, error) {
	keyValue, err := poolRegistry.keystore.Get(getKeepAliveKey(hostname))
	return keyValue.Key != "", err
}

func (poolRegistry *poolRegistry) AllocateBricks(allocations []registry.BrickAllocation) error {
	var bricks []registry.BrickInfo
	var volume registry.VolumeName
	var raw []KeyValue
	for i, allocation := range allocations {
		brick, err := poolRegistry.GetBrickInfo(allocation.Hostname, allocation.Device)
		if err != nil {
			return fmt.Errorf("unable to find brick for: %+v", allocation)
		}
		bricks = append(bricks, brick)

		if allocation.DeallocateRequested {
			return fmt.Errorf("should not requeste deallocated: %+v", allocation)
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

func (poolRegistry *poolRegistry) deallocate(raw []KeyValueVersion,
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

func (poolRegistry *poolRegistry) DeallocateBricks(volume registry.VolumeName) error {
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

func (poolRegistry *poolRegistry) HardDeleteAllocations(allocations []registry.BrickAllocation) error {
	var keys []string
	for _, allocation := range allocations {
		keys = append(keys, getBrickAllocationKeyHost(allocation))
		keys = append(keys, getBrickAllocationKeyVolume(allocation))
		if !allocation.DeallocateRequested {
			return fmt.Errorf("must first call deallocate on: %+v", allocation)
		}
	}

	var keyValues []KeyValueVersion
	for _, key := range keys {
		keyValue, err := poolRegistry.keystore.Get(key)
		if err != nil {
			return err
		}
		// TODO check we have already called deallocate properly
		keyValues = append(keyValues, keyValue)
	}
	return poolRegistry.keystore.DeleteAll(keyValues)
}

func (poolRegistry *poolRegistry) getAllocations(prefix string) ([]registry.BrickAllocation, error) {
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

func (poolRegistry *poolRegistry) GetAllocationsForHost(hostname string) ([]registry.BrickAllocation, error) {
	return poolRegistry.getAllocations(getPrefixAllocationHost(hostname))
}

func (poolRegistry *poolRegistry) GetAllocationsForVolume(volume registry.VolumeName) ([]registry.BrickAllocation, error) {
	return poolRegistry.getAllocations(getPrefixAllocationVolume(volume))
}

func (poolRegistry *poolRegistry) GetBrickInfo(hostname string, device string) (registry.BrickInfo, error) {
	raw, error := poolRegistry.keystore.Get(getBrickInfoKey(hostname, device))
	var value registry.BrickInfo
	json.Unmarshal(bytes.NewBufferString(raw.Value).Bytes(), &value)
	return value, error
}

func (poolRegistry *poolRegistry) WatchHostBrickAllocations(hostname string,
	callback func(old *registry.BrickAllocation, new *registry.BrickAllocation)) {
	key := getPrefixAllocationHost(hostname)
	poolRegistry.keystore.WatchPrefix(key, func(old *KeyValueVersion, new *KeyValueVersion) {
		oldBrick := &registry.BrickAllocation{}
		newBrick := &registry.BrickAllocation{}
		if old != nil {
			if old.Value != "" {
				json.Unmarshal(bytes.NewBufferString(old.Value).Bytes(), &oldBrick)
			}
		}
		if new != nil {
			if new.Value != "" {
				json.Unmarshal(bytes.NewBufferString(new.Value).Bytes(), &newBrick)
			}
		}
		callback(oldBrick, newBrick)
	})
}

func (poolRegistry *poolRegistry) getBricks(prefix string) ([]registry.BrickInfo, error) {
	raw, err := poolRegistry.keystore.GetAll(prefix)
	if err != nil {
		return nil, err
	}
	var allocations []registry.BrickInfo
	for _, entry := range raw {
		rawValue := entry.Value
		var allocation registry.BrickInfo
		json.Unmarshal(bytes.NewBufferString(rawValue).Bytes(), &allocation)
		allocations = append(allocations, allocation)
	}
	return allocations, nil
}

func (poolRegistry *poolRegistry) Pools() ([]registry.Pool, error) {
	allBricks, _ := poolRegistry.getBricks(registeredBricksPrefix)
	allAllocations, _ := poolRegistry.getAllocations(allocatedBricksPrefix)

	allocationLookup := make(map[string]registry.BrickAllocation)
	for _, allocation := range allAllocations {
		key := fmt.Sprintf("%s/%s", allocation.Hostname, allocation.Device)
		allocationLookup[key] = allocation
	}

	pools := make(map[string]*registry.Pool)
	hosts := make(map[string]*registry.HostInfo)
	for _, brick := range allBricks {
		pool, ok := pools[brick.PoolName]
		if !ok {
			pool = &registry.Pool{
				Name:            brick.PoolName,
				GranularityGB:   brick.CapacityGB,
				AllocatedBricks: []registry.BrickAllocation{},
				AvailableBricks: []registry.BrickInfo{},
				Hosts:           make(map[string]registry.HostInfo),
			}
		}

		if brick.CapacityGB != pool.GranularityGB {
			log.Printf("brick doesn't match pool granularity: %+v\n", brick)
			if brick.CapacityGB < pool.GranularityGB {
				pool.GranularityGB = brick.CapacityGB
			}
		}

		host, ok := hosts[brick.Hostname]
		if !ok {
			hostAlive, _ := poolRegistry.HostAlive(brick.Hostname)
			host = &registry.HostInfo{
				Hostname: brick.Hostname,
				Alive:    hostAlive,
			}
		}

		if _, ok := pool.Hosts[brick.Hostname]; !ok {
			pool.Hosts[brick.Hostname] = *host
		}

		key := fmt.Sprintf("%s/%s", brick.Hostname, brick.Device)
		allocation, ok := allocationLookup[key]
		if ok {
			pool.AllocatedBricks = append(pool.AllocatedBricks, allocation)
			hosts[brick.Hostname] = host
			pools[brick.PoolName] = pool
		} else {
			if host.Alive {
				pool.AvailableBricks = append(pool.AvailableBricks, brick)
				hosts[brick.Hostname] = host
				pools[brick.PoolName] = pool
			}
		}
	}

	var poolList []registry.Pool
	for _, value := range pools {
		poolList = append(poolList, *value)
	}
	return poolList, nil
}
