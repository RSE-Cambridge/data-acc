package fakewarp

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"math"
	"math/rand"
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
func CreatePersistentBuffer(c CliContext, volReg registry.VolumeRegistry,
	poolReg registry.PoolRegistry) (string, error) {

	request := BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), accessModeFromString(c.String("access")),
		bufferTypeFromString(c.String("type")), true}
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

	err = getBricksForBuffer(poolRegistry, pool, volume)
	if err != nil {
		volReg.DeleteVolume(volume.Name)
		volReg.DeleteJob(job.Name)
	}

	// if there are no bricks requested, don't wait for a provision that will never happen
	if volume.SizeBricks != 0 {
		volReg.WaitForState(volume.Name, registry.BricksProvisioned)
	}
	return err
}

func getBricksForBuffer(poolRegistry registry.PoolRegistry,
	pool registry.Pool, volume registry.Volume) error {

	if volume.SizeBricks == 0 {
		// No bricks requested, so return right away
		return nil
	}

	availableBricks := pool.AvailableBricks
	availableBricksByHost := make(map[string][]registry.BrickInfo)
	for _, brick := range availableBricks {
		hostBricks := availableBricksByHost[brick.Hostname]
		availableBricksByHost[brick.Hostname] = append(hostBricks, brick)
	}

	var chosenBricks []registry.BrickInfo

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	var hosts []string
	for key := range availableBricksByHost {
		hosts = append(hosts, key)
	}

	randomWalk := rand.Perm(len(availableBricksByHost))
	for _, i := range randomWalk {
		hostBricks := availableBricksByHost[hosts[i]]
		candidateBrick := hostBricks[r.Intn(len(hostBricks))]

		goodCandidate := true
		for _, brick := range chosenBricks {
			if brick == candidateBrick {
				goodCandidate = false
				break
			}
			if brick.Hostname == candidateBrick.Hostname {
				goodCandidate = false
				break
			}
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}
		if uint(len(chosenBricks)) >= volume.SizeBricks {
			break
		}
	}

	if uint(len(chosenBricks)) != volume.SizeBricks {
		return fmt.Errorf("unable to get number of requested bricks (%d) for given pool (%s)",
			volume.SizeBricks, pool.Name)
	}

	var allocations []registry.BrickAllocation
	for _, brick := range chosenBricks {
		allocations = append(allocations, registry.BrickAllocation{
			Device:              brick.Device,
			Hostname:            brick.Hostname,
			AllocatedVolume:     volume.Name,
			DeallocateRequested: false,
		})
	}
	err := poolRegistry.AllocateBricks(allocations)
	if err != nil {
		return err
	}
	_, err = poolRegistry.GetAllocationsForVolume(volume.Name) // TODO return result, wait for updates
	return err
}
