package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func DeleteBufferComponents(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	token string) error {

	volumeName := registry.VolumeName(token)
	volume, err := volumeRegistry.Volume(volumeName)
	if err != nil {
		// TODO should check this error relates to the volume being missing
		log.Println(err)
		return nil
	}

	vlm := lifecycle.NewVolumeLifecycleManager(volumeRegistry, poolRegistry, volume)
	if err := vlm.Delete(); err != nil {
		return err
	}

	return volumeRegistry.DeleteJob(token)
}

func CreatePerJobBuffer(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry, disk fileio.Disk,
	token string, user int, group int, capacity string, caller string, job string) error {
	if summary, err := ParseJobFile(disk, job); err != nil {
		return err
	} else {
		log.Println("Summary of job file:", summary)
	}
	return CreateVolumesAndJobs(volumeRegistry, poolRegistry, BufferRequest{
		Token:    token,
		User:     user,
		Group:    group,
		Capacity: capacity,
		Caller:   caller,
	})
}
