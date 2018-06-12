package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"log"
)

func TestFakewarp(keystore keystoreregistry.Keystore) {
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	poolReg := keystoreregistry.NewPoolRegistry(keystore)

	bufferRequest := fakewarp.BufferRequest{
		Token: "fakebuffer1",
		Capacity: "b:10GiB",
		Persistent: true,
		Caller: "test",
	}
	log.Println(fakewarp.GetPools(poolReg))
	log.Println(fakewarp.GetInstances(volReg))
	log.Println(fakewarp.GetSessions(volReg))

	log.Println(fakewarp.CreateVolumesAndJobs(volReg, bufferRequest))

	log.Println(fakewarp.GetPools(poolReg))
	log.Println(fakewarp.GetInstances(volReg))
	log.Println(fakewarp.GetSessions(volReg))
}
