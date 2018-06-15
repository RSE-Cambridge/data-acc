package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"strings"
)

type FakewarpActions interface {
	CreatePersistentBuffer(c CliContext) (string, error)
}

func NewFakewarpActions(
	poolRegistry registry.PoolRegistry, volumeRegistry registry.VolumeRegistry, reader fileio.Reader) FakewarpActions {

	return &fakewarpActions{poolRegistry, volumeRegistry, reader}
}

type fakewarpActions struct {
	poolRegistry   registry.PoolRegistry
	volumeRegistry registry.VolumeRegistry
	reader         fileio.Reader
}

func (fwa *fakewarpActions) CreatePersistentBuffer(c CliContext) (string, error) {
	checkRequiredStrings(c, "token", "caller", "capacity", "user", "access", "type")
	request := BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), accessModeFromString(c.String("access")),
		bufferTypeFromString(c.String("type")), true}
	if request.Group == 0 {
		request.Group = request.User
	}
	return request.Token, CreateVolumesAndJobs(fwa.volumeRegistry, fwa.poolRegistry, request)
}

func checkRequiredStrings(c CliContext, flags ...string) {
	errors := []string{}
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errors = append(errors, flag)
		}
	}
	if len(errors) > 0 {
		log.Fatalf("Please provide these required parameters: %s", strings.Join(errors, ", "))
	}
}
