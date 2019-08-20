package registry_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl/actions_impl/parsers"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"log"
)

func NewBrickHostRegistry(keystore store.Keystore) registry.BrickHostRegistry {
	return &brickHostRegistry{keystore}
}

type brickHostRegistry struct {
	store store.Keystore
}

const brickHostPrefix = "/BrickHostStore/"
const keepAlivePrefix = "/BrickHostAlive/"

func (b *brickHostRegistry) UpdateBrickHost(brickHostInfo datamodel.BrickHost) error {
	// find out granularity for each reported pool
	if len(brickHostInfo.Bricks) == 0 {
		log.Panicf("brick host must have some bricks: %s", brickHostInfo.Name)
	}
	poolGranularityGiBMap := make(map[datamodel.PoolName]uint)
	for _, brick := range brickHostInfo.Bricks {
		for !parsers.IsValidName(string(brick.PoolName)) {
			log.Panicf("invalid pool name: %+v", brick)
		}
		poolGranularity, ok := poolGranularityGiBMap[brick.PoolName]
		if !ok {
			if brick.CapacityGiB <= 0 {
				log.Panicf("invalid brick size: %+v", brick)
			}
			poolGranularityGiBMap[brick.PoolName] = brick.CapacityGiB
		} else {
			if brick.CapacityGiB != poolGranularity {
				log.Panicf("inconsistent brick size: %+v", brick)
			}
		}
	}

	// TODO: odd dependencies here!!
	allocationRegistry := NewAllocationRegistry(b.store)

	// Check existing pools match what this brick host is reporting
	for poolName, granularityGiB := range poolGranularityGiBMap {
		_, err := allocationRegistry.EnsurePoolCreated(poolName, parsers.GetBytes(granularityGiB, "GiB"))
		if err != nil {
			return fmt.Errorf("unable to create pool due to: %s", err)
		}
	}

	if !parsers.IsValidName(string(brickHostInfo.Name)) {
		log.Panicf("invalid brick host name: %s", brickHostInfo.Name)
	}
	key := fmt.Sprintf("%s%s", brickHostPrefix, brickHostInfo.Name)
	value, err := json.Marshal(brickHostInfo)
	if err != nil {
		log.Panicf("unable to covert brick host to json: %s", brickHostInfo.Name)
	}

	// Always overwrite any pre-existing key
	_, err = b.store.Update(key, value, 0)
	return err
}

func (b *brickHostRegistry) GetAllBrickHosts() ([]datamodel.BrickHost, error) {
	allKeyValues, err := b.store.GetAll(brickHostPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to get all bricks hosts due to: %s", err)
	}

	var allBrickHosts []datamodel.BrickHost
	for _, keyValueVersion := range allKeyValues {
		brickHost := datamodel.BrickHost{}
		err := json.Unmarshal(keyValueVersion.Value, &brickHost)
		if err != nil {
			log.Panicf("unable to parse brick host due to: %s", err)
		}
		allBrickHosts = append(allBrickHosts, brickHost)
	}
	return allBrickHosts, nil
}

func getKeepAliveKey(brickHostName datamodel.BrickHostName) string {
	if !parsers.IsValidName(string(brickHostName)) {
		log.Panicf("invalid brick host name: %s", brickHostName)
	}
	return fmt.Sprintf("%s%s", keepAlivePrefix, brickHostName)
}

func (b *brickHostRegistry) KeepAliveHost(ctxt context.Context, brickHostName datamodel.BrickHostName) error {
	return b.store.KeepAliveKey(ctxt, getKeepAliveKey(brickHostName))
}

func (b *brickHostRegistry) IsBrickHostAlive(brickHostName datamodel.BrickHostName) (bool, error) {
	return b.store.IsExist(getKeepAliveKey(brickHostName))
}
