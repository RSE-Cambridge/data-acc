package actions

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"strings"
)

type CliContext interface {
	String(name string) string
	Int(name string) int
}

type FakewarpActions interface {
	CreatePersistentBuffer(c CliContext) error
	DeleteBuffer(c CliContext) error
	CreatePerJobBuffer(c CliContext) error
	ShowInstances() error
	ShowSessions() error
	ListPools() error
	ShowConfigurations() error
	ValidateJob(c CliContext) error
	RealSize(c CliContext) error
	DataIn(c CliContext) error
	Paths(c CliContext) error
	PreRun(c CliContext) error
	PostRun(c CliContext) error
	DataOut(c CliContext) error
}

func NewFakewarpActions(
	poolRegistry registry.PoolRegistry, volumeRegistry registry.VolumeRegistry, disk fileio.Disk) FakewarpActions {

	return &fakewarpActions{poolRegistry, volumeRegistry, disk}
}

type fakewarpActions struct {
	poolRegistry   registry.PoolRegistry
	volumeRegistry registry.VolumeRegistry
	disk           fileio.Disk
}

func (fwa *fakewarpActions) CreatePersistentBuffer(c CliContext) error {
	checkRequiredStrings(c, "token", "caller", "capacity", "user", "access", "type")
	request := fakewarp.BufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), fakewarp.AccessModeFromString(c.String("access")),
		fakewarp.BufferTypeFromString(c.String("type")), true}
	if request.Group == 0 {
		request.Group = request.User
	}
	err := fakewarp.CreateVolumesAndJobs(fwa.volumeRegistry, fwa.poolRegistry, request)
	if err == nil {
		// Slurm is looking for the string "created" to know this worked
		fmt.Printf("created %s\n", request.Token)
	}
	return err
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
	if summary, err := fakewarp.ParseJobFile(fwa.disk, c.String("job")); err != nil {
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

func (fwa *fakewarpActions) ShowInstances() error {
	instances, err := fakewarp.GetInstances(fwa.volumeRegistry)
	if err != nil {
		return err
	}
	fmt.Println(instances)
	return nil
}

func (fwa *fakewarpActions) ShowSessions() error {
	sessions, err := fakewarp.GetSessions(fwa.volumeRegistry)
	if err != nil {
		return err
	}
	fmt.Println(sessions)
	return nil
}

func (fwa *fakewarpActions) ListPools() error {
	pools, err := fakewarp.GetPools(fwa.poolRegistry)
	if err != nil {
		return err
	}
	fmt.Println(pools)
	return nil
}

func (fwa *fakewarpActions) ShowConfigurations() error {
	fmt.Print(fakewarp.GetConfigurations())
	return nil
}

func (fwa *fakewarpActions) ValidateJob(c CliContext) error {
	checkRequiredStrings(c, "job")
	if summary, err := fakewarp.ParseJobFile(fwa.disk, c.String("job")); err != nil {
		return err
	} else {
		// TODO check valid pools, etc, etc.
		log.Println("Summary of job file:", summary)
	}
	return nil
}

func (fwa *fakewarpActions) RealSize(c CliContext) error {
	checkRequiredStrings(c, "token")
	// TODO need to fetch volume and get size, return in correct format
	fmt.Printf("--token %s\n", c.String("token"))
	return nil
}

func (fwa *fakewarpActions) DataIn(c CliContext) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n", c.String("token"), c.String("job"))

	volume, err := fwa.volumeRegistry.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
	return vlm.DataIn()
}

func (fwa *fakewarpActions) Paths(c CliContext) error {
	checkRequiredStrings(c, "token", "job", "pathfile")
	fmt.Printf("--token %s --job %s --pathfile %s\n",
		c.String("token"), c.String("job"), c.String("pathfile"))
	volume, err := fwa.volumeRegistry.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}
	return fwa.disk.Write(c.String("pathfile"), volume.Paths)
}

var testVLM lifecycle.VolumeLifecycleManager

func (fwa *fakewarpActions) getVolumeLifecycleManger(volume registry.Volume) lifecycle.VolumeLifecycleManager {
	if testVLM != nil {
		return testVLM
	}
	return lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
}

func (fwa *fakewarpActions) PreRun(c CliContext) error {
	checkRequiredStrings(c, "token", "job", "nodehostnamefile")
	fmt.Printf("--token %s --job %s --nodehostnamefile %s\n",
		c.String("token"), c.String("job"), c.String("nodehostnamefile"))

	// TODO: really we should get the job and mount all the volumes?
	volume, err := fwa.volumeRegistry.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	hosts, err := fwa.disk.Lines(c.String("nodehostnamefile"))
	if err != nil {
		return err
	}
	if len(hosts) < 1 {
		return errors.New("unable to mount to zero compute hosts")
	}

	// TODO: update the job with the list of hosts, so we have the list for unmount?

	vlm := fwa.getVolumeLifecycleManger(volume)
	return vlm.Mount(hosts)
}

func (fwa *fakewarpActions) PostRun(c CliContext) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	volume, err := fwa.volumeRegistry.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
	return vlm.Unmount()
}

func (fwa *fakewarpActions) DataOut(c CliContext) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	volume, err := fwa.volumeRegistry.Volume(registry.VolumeName(c.String("token")))
	if err != nil {
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
	return vlm.DataOut()
}
