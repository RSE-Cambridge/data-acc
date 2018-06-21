package fakewarp

import (
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
