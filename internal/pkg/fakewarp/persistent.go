package fakewarp

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"math"
	"strings"
	"time"
)

type BufferRequest struct {
	Token      string
	Caller     string
	Capacity   string
	User       int
	Group      int
	Access     AccessMode
	Type       BufferType
	Persistent bool
}

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned

func parseCapacity(raw string) (string, int, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("must format capacity correctly and include pool")
	}
	pool := parts[0]
	rawCapacity := parts[1]
	sizeBytes, err := parseSize(rawCapacity)
	if err != nil {
		return "", 0, err
	}
	capacityInt := int(sizeBytes / bytesInGB)
	return pool, capacityInt, nil
}

// TODO: ideally this would be private, if not for testing
func CreateVolumesAndJobs(volReg registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	request BufferRequest) error {

	createdAt := uint(time.Now().Unix())
	poolName, capacityGB, err := parseCapacity(request.Capacity) // TODO lots of proper parsing to do here, get poolname, etc
	if err != nil {
		return err
	}

	pools, err := poolRegistry.Pools()
	if err != nil {
		return err
	}

	var pool registry.Pool
	for _, p := range pools {
		if p.Name == poolName {
			pool = p
		}
	}

	if pool.Name == "" {
		return fmt.Errorf("unable to find pool: %s", poolName)
	}

	bricksRequired := uint(math.Ceil(float64(capacityGB) / float64(pool.GranularityGB)))
	adjustedSize := bricksRequired * pool.GranularityGB

	volume := registry.Volume{
		Name:       registry.VolumeName(request.Token),
		JobName:    request.Token,
		Owner:      request.User,
		CreatedAt:  createdAt,
		CreatedBy:  request.Caller,
		Group:      request.Group,
		SizeGB:     uint(adjustedSize),
		SizeBricks: bricksRequired,
		Pool:       pool.Name,
		State:      registry.Registered,
	}
	err = volReg.AddVolume(volume)
	if err != nil {
		return err
	}

	job := registry.Job{
		Name:      request.Token,
		Volumes:   []registry.VolumeName{registry.VolumeName(request.Token)},
		Owner:     uint(request.User),
		CreatedAt: createdAt,
	}
	err = volReg.AddJob(job)
	if err != nil {
		volReg.DeleteVolume(volume.Name)
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(volReg, poolRegistry, volume)
	err = vlm.ProvisionBricks(pool)
	if err != nil {
		volReg.DeleteVolume(volume.Name)
		volReg.DeleteJob(job.Name)
	}
	return err
}
