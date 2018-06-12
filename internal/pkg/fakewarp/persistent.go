package fakewarp

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"strconv"
	"strings"
	"time"
	"math"
)

type BufferRequest struct {
	Token      string
	Caller     string
	Capacity   string
	User       int
	Group      int
	Access     string
	Type       string
	Persistent bool
}

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func CreatePersistentBuffer(c CliContext, volReg registry.VolumeRegistry,
	poolReg registry.PoolRegistry) (string, error) {

	request := BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), c.String("access"), c.String("type"),
		true}
	if request.Group == 0 {
		request.Group = request.User
	}
	return request.Token, CreateVolumesAndJobs(volReg, poolReg, request)
}

func parseCapacity(raw string) (string, int, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("must format capacity correctly and include pool")
	}
	pool := parts[0]
	rawCapacity := parts[1]
	capacityParts := strings.Split(rawCapacity, "GiB")
	if len(capacityParts) > 2 {
		return "", 0, fmt.Errorf("must format capacity units correctly: %s", rawCapacity)
	}
	capacityInt, err := strconv.Atoi(capacityParts[0])
	if len(capacityParts) == 1 {
		capacityInt = int(capacityInt / bytesInGB)
	}
	if err != nil {
		return "", 0, fmt.Errorf("must format capacity amount: %s", rawCapacity)
	}
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

	err = volReg.AddVolume(registry.Volume{
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
	})
	if err != nil {
		return err
	}
	err = volReg.AddJob(registry.Job{
		Name:      request.Token,
		Volumes:   []registry.VolumeName{registry.VolumeName(request.Token)},
		Owner:     uint(request.User),
		CreatedAt: createdAt,
	})
	if err != nil {
		volReg.DeleteVolume(registry.VolumeName(request.Token))
	}
	// TODO: get bricks assigned to volume (i.e. ensure we have capacity)
	// TODO: wait for bricks to be provisioned correctly?
	return err
}
