package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func TestKeystoreVolumeRegistry(keystore keystoreregistry.Keystore) {
	log.Println("Testing keystoreregistry.volume")
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)

	testVolumeCRD(volumeRegistry)
}

func testVolumeCRD(volRegistry registry.VolumeRegistry) {
	volume := registry.Volume{Name: "asdf"}
	volume2 := registry.Volume{Name: "asdf2"}
	if err := volRegistry.AddVolume(volume); err != nil {
		log.Fatal(err)
	}
	if err := volRegistry.AddVolume(volume); err == nil {
		log.Fatal("expected an error")
	} else {
		log.Println(err)
	}

	if volume, err := volRegistry.Volume(volume.Name); err != nil {
		log.Fatal(err)
	} else {
		log.Println(volume)
	}

	if err := volRegistry.DeleteVolume(volume.Name); err != nil {
		log.Fatal(err)
	}
	if err := volRegistry.DeleteVolume(volume.Name); err == nil {
		log.Fatal("expected error")
	} else {
		log.Println(err)
	}

	// leave around for following tests
	volRegistry.AddVolume(volume)
	volRegistry.AddVolume(volume2)
}
