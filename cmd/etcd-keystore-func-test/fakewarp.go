package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func debugStatus(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry) {
	log.Println(fakewarp.GetPools(poolRegistry))
	log.Println(fakewarp.GetInstances(volumeRegistry))
	log.Println(fakewarp.GetSessions(volumeRegistry))
	log.Println(volumeRegistry.AllVolumes())
}

func testPersistent(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry) {
	bufferToken := "fakebuffer1"
	bufferRequest := fakewarp.BufferRequest{
		Token:      bufferToken,
		Capacity:   "b:10GiB",
		Persistent: true,
		Caller:     "test",
	}
	debugStatus(volumeRegistry, poolRegistry)

	log.Println(fakewarp.CreateVolumesAndJobs(volumeRegistry, poolRegistry, bufferRequest))
	bufferRequest2 := bufferRequest
	bufferRequest2.Token = "fakebuffer2"
	log.Println(fakewarp.CreateVolumesAndJobs(volumeRegistry, poolRegistry, bufferRequest2))

	debugStatus(volumeRegistry, poolRegistry)

	log.Println(fakewarp.DeleteBufferComponents(volumeRegistry, poolRegistry, bufferToken))
	log.Println(fakewarp.DeleteBufferComponents(volumeRegistry, poolRegistry, "fakebuffer2"))

	debugStatus(volumeRegistry, poolRegistry)
}

func TestFakewarp(keystore keystoreregistry.Keystore) {
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)
	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)

	testPersistent(volumeRegistry, poolRegistry)
}
