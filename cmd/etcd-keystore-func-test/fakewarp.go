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

	err := fakewarp.CreateVolumesAndJobs(volumeRegistry, poolRegistry, bufferRequest)
	if err != nil {
		log.Fatal(err)
	}

	bufferRequest2 := bufferRequest
	bufferRequest2.Token = "fakebuffer2"
	err = fakewarp.CreateVolumesAndJobs(volumeRegistry, poolRegistry, bufferRequest2)
	if err != nil {
		log.Fatal(err)
	}

	bufferRequest3 := fakewarp.BufferRequest{
		Token:      "fakebuffer3",
		Capacity:   "a:0",
		Persistent: true,
		Caller:     "test",
	}
	err = fakewarp.CreateVolumesAndJobs(volumeRegistry, poolRegistry, bufferRequest3)
	if err != nil {
		log.Fatal(err)
	}

	debugStatus(volumeRegistry, poolRegistry)

	// TODO go through state machine for a given volume...?
	// TODO fix up paths, real_size, etc
	// TODO record all the data for fake data_in, etc
	// TODO add wait for actions into volume state machine

	log.Println(fakewarp.DeleteBufferComponents(volumeRegistry, poolRegistry, bufferToken))
	log.Println(fakewarp.DeleteBufferComponents(volumeRegistry, poolRegistry, "fakebuffer2"))
	log.Println(fakewarp.DeleteBufferComponents(volumeRegistry, poolRegistry, "fakebuffer3"))

	debugStatus(volumeRegistry, poolRegistry)
}

func TestFakewarp(keystore keystoreregistry.Keystore) {
	log.Println("Testing fakewarp")

	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)
	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)

	testPersistent(volumeRegistry, poolRegistry)
}
