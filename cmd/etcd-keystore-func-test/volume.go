package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func TestKeystoreVolumeRegistry(keystore keystoreregistry.Keystore) {
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)
	testVolumeCRUD(volumeRegistry)
}

func testVolumeCRUD(registry registry.VolumeRegistry) {
	log.Println(registry.Volume("test"))
}
