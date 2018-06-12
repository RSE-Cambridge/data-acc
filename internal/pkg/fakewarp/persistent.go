package fakewarp

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"strconv"
	"strings"
	"time"
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
func CreatePersistentBuffer(c CliContext, volReg registry.VolumeRegistry) (string, error) {
	request := BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), c.String("access"), c.String("type"),
		true}
	if request.Group == 0 {
		request.Group = request.User
	}
	return request.Token, CreateVolumesAndJobs(volReg, request)
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
		capacityInt = capacityInt / bytesInGB
	}
	if err != nil {
		return "", 0, fmt.Errorf("must format capacity amount: %s", rawCapacity)
	}
	return pool, capacityInt, nil
}

// TODO: ideally this would be private, if not for testing
func CreateVolumesAndJobs(volReg registry.VolumeRegistry, request BufferRequest) error {
	createdAt := uint(time.Now().Unix())
	pool, capacity, err := parseCapacity(request.Capacity) // TODO lots of proper parsing to do here, get poolname, etc
	if err != nil {
		return err
	}
	err = volReg.AddVolume(registry.Volume{
		Name:       registry.VolumeName(request.Token),
		JobName:    request.Token,
		Owner:      request.User,
		CreatedAt:  createdAt,
		CreatedBy:  request.Caller,
		Group:      request.Group,
		SizeGB:     uint(capacity),
		SizeBricks: 3,    // TODO... check pool granularity
		Pool:       pool, // TODO....
		State:      registry.Registered,
	})
	if err != nil {
		return err
	}
	// TODO: get bricks assigned to volume (i.e. ensure we have capacity)
	err = volReg.AddJob(registry.Job{
		Name:      request.Token,
		Volumes:   []registry.VolumeName{registry.VolumeName(request.Token)},
		Owner:     uint(request.User),
		CreatedAt: createdAt,
	})
	if err != nil {
		volReg.DeleteVolume(registry.VolumeName(request.Token))
	}
	// TODO: wait for bricks to be provisioned correctly?
	return err
}
