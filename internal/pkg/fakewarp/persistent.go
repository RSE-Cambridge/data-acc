package fakewarp

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"strconv"
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
func CreatePersistentBuffer(c CliContext, keystore keystoreregistry.Keystore) (string, error) {
	request := BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), c.String("access"), c.String("type"),
		true}
	if request.Group == 0 {
		request.Group = request.User
	}
	return request.Token, createVolumesAndJobs(keystore, request)
}

func createVolumesAndJobs(keystore keystoreregistry.Keystore, request BufferRequest) error {
	createdAt := uint(time.Now().Unix())
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	capacity, err := strconv.Atoi(request.Capacity) // TODO lots of proper parsing to do here, get poolname, etc
	if err != nil {
		return errors.New("please format capacity correctly")
	}
	err = volReg.AddVolume(registry.Volume{
		Name:       registry.VolumeName(request.Token),
		JobName:    request.Token,
		Owner:      request.User,
		CreatedAt:  createdAt,
		CreatedBy:  request.Caller,
		Group:      request.Group,
		SizeGB:     uint(capacity),
		SizeBricks: 3,         // TODO... check pool granularity
		Pool:       "default", // TODO....
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
