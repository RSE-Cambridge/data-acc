package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"time"
)

func TestKeystoreVolumeRegistry(keystore keystoreregistry.Keystore) {
	log.Println("Testing keystoreregistry.volume")
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)

	testVolumeCRUD(volumeRegistry)
	testJobCRUD(volumeRegistry)

	// give watches time to print
	time.Sleep(time.Second)
}

func testVolumeCRUD(volRegistry registry.VolumeRegistry) {
	volRegistry.WatchVolumeChanges("asdf", func(old *registry.Volume, new *registry.Volume) {
		log.Printf("Volume update detected. old: %s new: %s", old.State, new.State)
	})

	volume := registry.Volume{Name: "asdf", State: registry.Registered, JobName: "foo", SizeBricks:2, SizeGB:200}
	volume2 := registry.Volume{Name: "asdf2", JobName: "foo", SizeBricks:3, SizeGB:300}
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

	if err := volRegistry.UpdateState(volume.Name, registry.BricksAssigned); err != nil {
		log.Fatal(err)
	}
	if err := volRegistry.UpdateState("badname", registry.BricksAssigned); err == nil {
		log.Fatal("expected error")
	}
	if err := volRegistry.UpdateState(volume.Name, registry.BricksAssigned); err == nil {
		log.Fatal("expected error with repeated update")
	}
	if err := volRegistry.UpdateState(volume.Name, registry.Test3); err == nil {
		log.Fatal("expected error with out of order update")
	}
	volRegistry.UpdateState(volume2.Name, registry.Registered)

	if volumes, err := volRegistry.AllVolumes(); err != nil {
		log.Fatal(err)
	} else {
		log.Println(volumes)
	}
}

func testJobCRUD(volRegistry registry.VolumeRegistry) {
	job := registry.Job{Name: "foo",
		Volumes:   []registry.VolumeName{"asdf", "asdf2"},
		Owner:     1001,
		CreatedAt: uint(time.Now().Unix()),
	}
	if err := volRegistry.AddJob(job); err != nil {
		log.Fatal(err)
	}

	if err := volRegistry.AddJob(job); err == nil {
		log.Fatal("expected an error adding duplicate job")
	}
	badJob := registry.Job{Name: "bar", Volumes: []registry.VolumeName{"asdf", "asdf3"}}
	if err := volRegistry.AddJob(badJob); err == nil {
		log.Fatal("expected an error for invalid volume name")
	}

	jobs, err := volRegistry.Jobs()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(jobs)

	err = volRegistry.DeleteJob("foo")
	if err != nil {
		panic(err)
	}
	err = volRegistry.DeleteJob("foo")
	if err == nil {
		panic(err)
	}
	volRegistry.AddJob(job)
}
