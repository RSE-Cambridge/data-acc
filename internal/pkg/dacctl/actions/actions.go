package actions

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl"
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

type DacctlActions interface {
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

func NewDacctlActions(
	poolRegistry registry.PoolRegistry, volumeRegistry registry.VolumeRegistry, disk fileio.Disk) DacctlActions {

	return &dacctlActions{poolRegistry, volumeRegistry, disk}
}

type dacctlActions struct {
	poolRegistry   registry.PoolRegistry
	volumeRegistry registry.VolumeRegistry
	disk           fileio.Disk
}

func (fwa *dacctlActions) CreatePersistentBuffer(c CliContext) error {
	checkRequiredStrings(c, "token", "caller", "capacity", "user", "access", "type")
	request := dacctl.BufferRequest{Token: c.String("token"), Caller: c.String("caller"),
		Capacity: c.String("capacity"), User: c.Int("user"),
		Group: c.Int("groupid"), Access: dacctl.AccessModeFromString(c.String("access")),
		Type: dacctl.BufferTypeFromString(c.String("type")), Persistent: true}
	if request.Group == 0 {
		request.Group = request.User
	}
	err := dacctl.CreateVolumesAndJobs(fwa.volumeRegistry, fwa.poolRegistry, request)
	if err == nil {
		// Slurm is looking for the string "created" to know this worked
		fmt.Printf("created %s\n", request.Token)
	}
	return err
}

func checkRequiredStrings(c CliContext, flags ...string) {
	errs := []string{}
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errs = append(errs, flag)
		}
	}
	if len(errs) > 0 {
		log.Fatalf("Please provide these required parameters: %s", strings.Join(errs, ", "))
	}
}

func (fwa *dacctlActions) DeleteBuffer(c CliContext) error {
	checkRequiredStrings(c, "token")
	token := c.String("token")
	return dacctl.DeleteBufferComponents(fwa.volumeRegistry, fwa.poolRegistry, token)
}

func (fwa *dacctlActions) CreatePerJobBuffer(c CliContext) error {
	checkRequiredStrings(c, "token", "job", "caller", "capacity")
	return dacctl.CreatePerJobBuffer(fwa.volumeRegistry, fwa.poolRegistry, fwa.disk,
		c.String("token"), c.Int("user"), c.Int("group"), c.String("capacity"),
		c.String("caller"), c.String("job"), c.String("nodehostnamefile"))
}

func (fwa *dacctlActions) ShowInstances() error {
	instances, err := dacctl.GetInstances(fwa.volumeRegistry)
	if err != nil {
		return err
	}
	fmt.Println(instances)
	return nil
}

func (fwa *dacctlActions) ShowSessions() error {
	sessions, err := dacctl.GetSessions(fwa.volumeRegistry)
	if err != nil {
		return err
	}
	fmt.Println(sessions)
	return nil
}

func (fwa *dacctlActions) ListPools() error {
	pools, err := dacctl.GetPools(fwa.poolRegistry)
	if err != nil {
		return err
	}
	fmt.Println(pools)
	return nil
}

func (fwa *dacctlActions) ShowConfigurations() error {
	fmt.Print(dacctl.GetConfigurations())
	return nil
}

func (fwa *dacctlActions) ValidateJob(c CliContext) error {
	checkRequiredStrings(c, "job")
	if summary, err := dacctl.ParseJobFile(fwa.disk, c.String("job")); err != nil {
		return err
	} else {
		// TODO check valid pools, etc, etc.
		log.Println("Summary of job file:", summary)
	}
	return nil
}

func (fwa *dacctlActions) RealSize(c CliContext) error {
	checkRequiredStrings(c, "token")
	job, err := fwa.volumeRegistry.Job(c.String("token"))
	if err != nil {
		return err
	}

	if job.JobVolume == "" {
		return fmt.Errorf("no volume to report the size of: %s", job.Name)
	}

	volume, err := fwa.volumeRegistry.Volume(job.JobVolume)
	if err != nil {
		return err
	}
	// TODO get GiB vs GB correct here!
	fmt.Printf(`{"token":"%s", "capacity":%d, "units":"bytes"}`, volume.Name, volume.SizeGB*1073741824)
	return nil
}

func (fwa *dacctlActions) DataIn(c CliContext) error {
	checkRequiredStrings(c, "token")
	fmt.Printf("--token %s --job %s\n", c.String("token"), c.String("job"))

	job, err := fwa.volumeRegistry.Job(c.String("token"))
	if err != nil {
		return err
	}

	if job.JobVolume == "" {
		log.Print("No data in required")
		return nil
	}

	volume, err := fwa.volumeRegistry.Volume(job.JobVolume)
	if err != nil {
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
	return vlm.DataIn()
}

func (fwa *dacctlActions) Paths(c CliContext) error {
	checkRequiredStrings(c, "token", "pathfile")
	fmt.Printf("--token %s --job %s --pathfile %s\n",
		c.String("token"), c.String("job"), c.String("pathfile"))

	job, err := fwa.volumeRegistry.Job(c.String("token"))
	if err != nil {
		return err
	}

	paths := []string{}
	for key, value := range job.Paths {
		paths = append(paths, fmt.Sprintf("%s=%s", key, value))
	}
	return fwa.disk.Write(c.String("pathfile"), paths)
}

var testVLM lifecycle.VolumeLifecycleManager

func (fwa *dacctlActions) getVolumeLifecycleManger(volume registry.Volume) lifecycle.VolumeLifecycleManager {
	if testVLM != nil {
		return testVLM
	}
	return lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
}

func (fwa *dacctlActions) PreRun(c CliContext) error {
	checkRequiredStrings(c, "token", "nodehostnamefile")
	fmt.Printf("--token %s --job %s --nodehostnamefile %s\n",
		c.String("token"), c.String("job"), c.String("nodehostnamefile"))

	job, err := fwa.volumeRegistry.Job(c.String("token"))
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

	err = fwa.volumeRegistry.JobAttachHosts(job.Name, hosts)
	if err != nil {
		return err
	}

	if job.JobVolume == "" {
		log.Print("No job volume to mount")
	} else {
		volume, err := fwa.volumeRegistry.Volume(job.JobVolume)
		if err != nil {
			return err
		}
		vlm := fwa.getVolumeLifecycleManger(volume)
		if err := vlm.Mount(hosts, job.Name); err != nil {
			return err
		}
	}

	for _, volumeName := range job.MultiJobVolumes {
		volume, err := fwa.volumeRegistry.Volume(volumeName)
		if err != nil {
			return err
		}
		vlm := fwa.getVolumeLifecycleManger(volume)
		if err := vlm.Mount(hosts, job.Name); err != nil {
			return err
		}
	}

	return nil
}

func (fwa *dacctlActions) PostRun(c CliContext) error {
	checkRequiredStrings(c, "token")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	job, err := fwa.volumeRegistry.Job(c.String("token"))
	if err != nil {
		return err
	}

	return dacctl.Unmount(fwa.volumeRegistry, fwa.poolRegistry, job)
}

func (fwa *dacctlActions) DataOut(c CliContext) error {
	checkRequiredStrings(c, "token")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	job, err := fwa.volumeRegistry.Job(c.String("token"))
	if err != nil {
		return err
	}

	if job.JobVolume == "" {
		log.Print("No data out required")
		return nil
	}

	volume, err := fwa.volumeRegistry.Volume(job.JobVolume)
	if err != nil {
		return err
	}

	vlm := lifecycle.NewVolumeLifecycleManager(fwa.volumeRegistry, fwa.poolRegistry, volume)
	return vlm.DataOut()
}
