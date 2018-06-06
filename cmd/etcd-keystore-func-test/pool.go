package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func testGetPools(poolRegistry registry.PoolRegistry) {
	/*if pools, err := poolRegistry.Pools(); err != nil {
		log.Fatal(err)
	} else {
		log.Println(pools)
	}*/
}

func testUpdateHost(poolRegistry registry.PoolRegistry) {
	brickInfo := []registry.BrickInfo{
		{Hostname: "foo", Device: "vbdb1", PoolName: "a", CapacityGB: 10},
		{Hostname: "foo", Device: "nvme3n1", PoolName: "b", CapacityGB: 20},
		{Hostname: "foo", Device: "nvme2n1", PoolName: "b", CapacityGB: 20},
	}
	err := poolRegistry.UpdateHost(brickInfo)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("added some keys")
	}

	// Do not allow multiple hostnames to be updated
	brickInfo = []registry.BrickInfo{
		{Hostname: "foo", Device: "vbdb1", PoolName: "a", CapacityGB: 10},
		{Hostname: "bar", Device: "nvme3n1", PoolName: "b", CapacityGB: 20},
	}
	err = poolRegistry.UpdateHost(brickInfo)
	if err == nil {
		log.Fatal("expected error")
	} else {
		log.Println(err)
	}
}

func testGetBricks(poolRegistry registry.PoolRegistry) {
	if raw, err := poolRegistry.GetBrickInfo("foo", "vbdb1"); err != nil {
		log.Fatal(err)
	} else {
		log.Println(raw)
	}

	if _, err := poolRegistry.GetBrickInfo("asdf", "vbdb1"); err != nil {
		log.Println(err)
	} else {
		log.Fatal("expected error")
	}
}

func testAllocateBricks(poolRegistry registry.PoolRegistry) {
	allocations := []registry.BrickAllocation{
		{Hostname: "foo", Device: "vbdb1", AllocatedVolume: "vol1"},
		{Hostname: "foo", Device: "nvme3n1", AllocatedVolume: "vol1"},
	}
	if err := poolRegistry.AllocateBricks(allocations); err != nil {
		log.Fatal(err)
	}
}

func testGetAllocations(poolRegistry registry.PoolRegistry) {
	allocations, err := poolRegistry.GetAllocationsForHost("foo")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(allocations)

	allocations, err = poolRegistry.GetAllocationsForVolume("vol1")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(allocations)
}

func TestKeystorePoolRegistry() {
	log.Println("Testing keystoreregistry.pool")
	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()
	cleanAllKeys(keystore)

	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)
	testUpdateHost(poolRegistry)
	testGetBricks(poolRegistry)
	testAllocateBricks(poolRegistry)
	testGetAllocations(poolRegistry)

	// TODO: update hosts first?
	testGetPools(poolRegistry)
}
