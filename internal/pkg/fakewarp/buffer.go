package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func DeleteBuffer(c CliContext, volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry) error {
	token := c.String("token")
	return DeleteBufferComponents(volumeRegistry, poolRegistry, token)
}

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

func CreatePerJobBuffer(c CliContext, volReg registry.VolumeRegistry, poolReg registry.PoolRegistry,
	reader fileio.Reader) error {

	if summary, err := parseJobFile(reader, c.String("job")); err != nil {
		return err
	} else {
		log.Println("Summary of job file:", summary)
	}
	return CreateVolumesAndJobs(volReg, poolReg, BufferRequest{
		Token:    c.String("token"),
		User:     c.Int("user"),
		Group:    c.Int("group"),
		Capacity: c.String("capacity"),
		Caller:   c.String("caller"),
	})
}
