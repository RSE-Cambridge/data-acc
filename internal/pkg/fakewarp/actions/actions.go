package actions

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"strings"
)

type CliContext interface {
	String(name string) string
	Int(name string) int
}

type FakewarpActions interface {
	CreatePersistentBuffer(c CliContext) (string, error)
	DeleteBuffer(c CliContext) error
	CreatePerJobBuffer(c CliContext) error
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
	request := fakewarp.BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), fakewarp.AccessModeFromString(c.String("access")),
		fakewarp.BufferTypeFromString(c.String("type")), true}
	if request.Group == 0 {
		request.Group = request.User
	}
	return request.Token, fakewarp.CreateVolumesAndJobs(fwa.volumeRegistry, fwa.poolRegistry, request)
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

func (fwa *fakewarpActions) DeleteBuffer(c CliContext) error {
	checkRequiredStrings(c, "token", "job")
	token := c.String("token")
	return fakewarp.DeleteBufferComponents(fwa.volumeRegistry, fwa.poolRegistry, token)
}

func (fwa *fakewarpActions) CreatePerJobBuffer(c CliContext) error {
	checkRequiredStrings(c, "token", "job", "caller", "capacity")
	if summary, err := fakewarp.ParseJobFile(fwa.reader, c.String("job")); err != nil {
		return err
	} else {
		log.Println("Summary of job file:", summary)
	}
	return fakewarp.CreateVolumesAndJobs(fwa.volumeRegistry, fwa.poolRegistry, fakewarp.BufferRequest{
		Token:    c.String("token"),
		User:     c.Int("user"),
		Group:    c.Int("group"),
		Capacity: c.String("capacity"),
		Caller:   c.String("caller"),
	})
}
