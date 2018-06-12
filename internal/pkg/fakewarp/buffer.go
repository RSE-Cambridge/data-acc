package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
)

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func DeleteBuffer(c CliContext, volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry) error {
	token := c.String("token")
	return DeleteBufferComponents(volumeRegistry, poolRegistry, token)
}

func DeleteBufferComponents(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	token string) error {

	// TODO... ignore delete brick errors when allocations don't exist!
	volumeName := registry.VolumeName(token)
	err := poolRegistry.DeallocateBricks(volumeName)
	if err != nil {
		return err
	}
	allocations, err := poolRegistry.GetAllocationsForVolume(volumeName)
	if err != nil {
		return err
	}
	poolRegistry.HardDeleteAllocations(allocations)

	if err := volumeRegistry.DeleteVolume(volumeName); err != nil {
		return err
	}
	return volumeRegistry.DeleteJob(token)
}

func CreatePerJobBuffer(c CliContext, volReg registry.VolumeRegistry, poolReg registry.PoolRegistry,
	lineSrc GetLines) error {

	// TODO need to read and parse the job file...
	if err := parseJobFile(lineSrc, c.String("job")); err != nil {
		return err
	}
	return CreateVolumesAndJobs(volReg, poolReg, BufferRequest{
		Token:    c.String("token"),
		User:     c.Int("user"),
		Group:    c.Int("group"),
		Capacity: c.String("capacity"),
		Caller:   c.String("caller"),
	})
}
