package registry_impl

import (
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl/actions_impl/parsers"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"log"
)

func NewAllocationRegistry(store store.Keystore) registry.AllocationRegistry {
	// TODO: create brickHostRegistry
	return &allocationRegistry{store, nil, nil}
}

type allocationRegistry struct {
	store             store.Keystore
	brickHostRegistry registry.BrickHostRegistry
	sessionRegistry   registry.SessionRegistry
}

const poolPrefix = "/Pool/"
const allocationLockKey = "/LockAllocation/"

func (a *allocationRegistry) GetAllocationMutex() (store.Mutex, error) {
	return a.store.NewMutex(allocationLockKey)
}

func getPoolKey(poolName datamodel.PoolName) string {
	if !parsers.IsValidName(string(poolName)) {
		log.Panicf("invalid session PrimaryBrickHost")
	}
	return fmt.Sprintf("%s%s", poolPrefix, poolName)
}

func (a *allocationRegistry) EnsurePoolCreated(poolName datamodel.PoolName, granularityBytes uint) (datamodel.Pool, error) {
	if granularityBytes <= 0 {
		log.Panicf("granularity must be greater than 0")
	}
	key := getPoolKey(poolName)
	poolExists, err := a.store.IsExist(key)
	if err != nil {
		return datamodel.Pool{}, fmt.Errorf("unable to check if pool exists: %s", err)
	}

	if poolExists {
		pool, err := a.GetPool(poolName)
		if err != nil {
			return pool, fmt.Errorf("unable to get pool due to: %s", err)
		}
		if pool.GranularityBytes != granularityBytes {
			return pool, fmt.Errorf("granularity doesn't match existing pool: %d", pool.GranularityBytes)
		}
		return pool, nil
	}

	// TODO: need an admin tool to delete a "bad" pool
	// create the pool
	pool := datamodel.Pool{Name: poolName, GranularityBytes: granularityBytes}
	value, err := json.Marshal(pool)
	if err != nil {
		log.Panicf("failed to convert pool to json: %s", err)
	}
	_, err = a.store.Create(key, value)
	return pool, err
}

func (a *allocationRegistry) GetPool(poolName datamodel.PoolName) (datamodel.Pool, error) {
	key := getPoolKey(poolName)
	keyValueVersion, err := a.store.Get(key)
	pool := datamodel.Pool{}
	if err != nil {
		return pool, fmt.Errorf("unable to get pool due to: %s", err)
	}

	err = json.Unmarshal(keyValueVersion.Value, &pool)
	if err != nil {
		log.Panicf("unable to parse pool")
	}
	return pool, nil
}

func (a *allocationRegistry) getAllPools() (map[datamodel.PoolName]datamodel.Pool, error) {
	allKeyValues, err := a.store.GetAll(poolPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to get pools due to: %s", err)
	}
	pools := make(map[datamodel.PoolName]datamodel.Pool)
	for _, keyValueVersion := range allKeyValues {
		pool := datamodel.Pool{}
		err = json.Unmarshal(keyValueVersion.Value, &pool)
		if err != nil {
			log.Panicf("unable to parse pool")
		}
		pools[pool.Name] = pool
	}
	return pools, nil
}

func (a *allocationRegistry) GetAllPoolInfos() ([]datamodel.PoolInfo, error) {
	pools, err := a.getAllPools()
	if err != nil {
		return nil, fmt.Errorf("unable to get pools due to: %s", err)
	}
	sessions, err := a.sessionRegistry.GetAllSessions()
	if err != nil {
		return nil, fmt.Errorf("unable to get all sessions due to: %s", err)
	}
	brickHosts, err := a.brickHostRegistry.GetAllBrickHosts()
	if err != nil {
		return nil, fmt.Errorf("unable to get all briks due to: %s", err)
	}

	var allPoolInfos []datamodel.PoolInfo

	for _, pool := range pools {
		poolInfo := datamodel.PoolInfo{Pool: pool}

		allocatedDevicesByBrickHost := make(map[datamodel.BrickHostName][]string)
		for _, session := range sessions {
			for i, brick := range session.AllocatedBricks {
				if brick.PoolName == pool.Name {
					poolInfo.AllocatedBricks = append(poolInfo.AllocatedBricks, datamodel.BrickAllocation{
						Session:        session.Name,
						Brick:          brick,
						AllocatedIndex: uint(i),
					})
					allocatedDevicesByBrickHost[brick.BrickHostName] = append(allocatedDevicesByBrickHost[brick.BrickHostName], brick.Device)
				}
			}
		}

		for _, brickHost := range brickHosts {
			for _, brick := range brickHost.Bricks {
				log.Println(brick)
				// Check if in allocated list
				allocated := false
				for _, allocatedDevice := range allocatedDevicesByBrickHost[brick.BrickHostName] {
					if allocatedDevice == brick.Device {
						if allocated {
							log.Panicf("detected duplicated brick allocation: %+v", brick)
						}
						allocated = true
					}
				}
				if !allocated {
					poolInfo.AvailableBricks = append(poolInfo.AvailableBricks, brick)
				}
			}
		}

		allPoolInfos = append(allPoolInfos, poolInfo)
	}
	return allPoolInfos, nil
}

func (a *allocationRegistry) GetPoolInfo(poolName datamodel.PoolName) (datamodel.PoolInfo, error) {
	allInfo, err := a.GetAllPoolInfos()
	if err != nil {
		return datamodel.PoolInfo{}, err
	}

	for _, poolInfo := range allInfo {
		if poolInfo.Pool.Name == poolName {
			return poolInfo, nil
		}
	}
	return datamodel.PoolInfo{}, fmt.Errorf("unable to find pool %s", poolName)
}
