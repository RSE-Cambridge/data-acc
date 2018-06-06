package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func testGetPools(pool registry.PoolRegistry) {
	if pools, err := pool.Pools(); err != nil {
		log.Fatal(err)
	} else {
		log.Println(pools)
	}
}

func TestKeystorePoolRegistry() {
	log.Println("Testing keystoreregistry.pool")
	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()
	cleanAllKeys(keystore)

	pool := keystoreregistry.NewPoolRegistry(keystore)

	// TODO: update hosts first?
	testGetPools(pool)
}
