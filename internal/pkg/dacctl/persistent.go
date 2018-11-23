package dacctl

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
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

	if pool == nil {
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
	// Add one more for the metadata... TODO: lustre specific?
	if bricksRequired != 0 {
		bricksRequired += 1
	}
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
	}
	err = volReg.AddVolume(volume)
	if err != nil {
		return err
	}

	job := registry.Job{
		Name:      request.Token,
		Owner:     uint(request.User),
		CreatedAt: createdAt,
		JobVolume: volume.Name, // Even though its a persistent buffer, we add it here to ensure we delete buffer
		Paths:     make(map[string]string),
	}
	job.Paths[fmt.Sprintf("DW_PERSISTENT_STRIPED_%s", volume.Name)] = fmt.Sprintf(
		"/mnt/dac/job/%s/multijob/%s", job.Name, volume.Name)

	err = volReg.AddJob(job)
	if err != nil {
		delErr := volReg.DeleteVolume(volume.Name)
		log.Println("volume deleted: ", delErr) // TODO: remove debug logs later, once understood
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(volReg, poolRegistry, volume)
	err = vlm.ProvisionBricks(*pool)
	if err != nil {
		log.Println("Bricks may be left behnd, not deleting volume due to: ", err)
	}
	return err
}
