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

func findPool(poolRegistry registry.PoolRegistry, poolName string) (pool *registry.Pool, err error) {
	pools, err := poolRegistry.Pools()
	if err != nil {
		return
	}

	for _, p := range pools {
		if p.Name == poolName {
			pool = &p
		}
	}

	if pool.Name == "" {
		err = fmt.Errorf("unable to find pool: %s", poolName)
		return
	}
	return
}

func getPoolAndBrickCount(poolRegistry registry.PoolRegistry, capacity string) (pool *registry.Pool,
	bricksRequired uint, err error) {

	poolName, capacityGB, err := parseCapacity(capacity)
	if err != nil {
		return
	}

	pool, err = findPool(poolRegistry, poolName)
	if err != nil {
		return
	}

	bricksRequired = uint(math.Ceil(float64(capacityGB) / float64(pool.GranularityGB)))
	return
}

// TODO: ideally this would be private, if not for testing
func CreateVolumesAndJobs(volReg registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	request BufferRequest) error {

	createdAt := uint(time.Now().Unix())

	pool, bricksRequired, err := getPoolAndBrickCount(poolRegistry, request.Capacity)
	if err != nil {
		return err
	}
	adjustedSizeGB := bricksRequired * pool.GranularityGB

	// TODO should populate configurations also, from job file
	var suffix string
	if request.Persistent {
		// TODO: sanitize token!!!
		suffix = fmt.Sprintf("PERSISTENT_STRIPED_%s=/mnt/dac/persistent/%s", request.Token, request.Token)
	} else {
		// TODO: not only striped, long term
		suffix = fmt.Sprintf("JOB_STRIPED=/mnt/dac/job/%s/striped", request.Token)
	}
	paths := []string{
		fmt.Sprintf("BB_%s", suffix),
		fmt.Sprintf("DW_%s", suffix),
	}

	volume := registry.Volume{
		Name:       registry.VolumeName(request.Token),
		JobName:    request.Token,
		Owner:      uint(request.User),
		CreatedAt:  createdAt,
		CreatedBy:  request.Caller,
		Group:      uint(request.Group),
		SizeGB:     adjustedSizeGB,
		SizeBricks: bricksRequired,
		Pool:       pool.Name,
		State:      registry.Registered,
		MultiJob:   request.Persistent,
		Paths:      paths,
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
	err = vlm.ProvisionBricks(*pool)
	if err != nil {
		volReg.DeleteVolume(volume.Name)
		volReg.DeleteJob(job.Name)
	}
	return err
}
