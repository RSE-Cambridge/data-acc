package dacctl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"strings"
	"time"
)

func DeleteBufferComponents(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	token string) error {

	job, err := volumeRegistry.Job(token)
	if err != nil {
		if strings.Contains(err.Error(), "unable to find any values for key") {
			log.Println("Unable to find job, must be deleted already or never created.")
			return nil
		}
		return err
	}

	if job.JobVolume != "" {
		volume, err := volumeRegistry.Volume(job.JobVolume)
		if err != nil {
			return err
		} else {
			vlm := lifecycle.NewVolumeLifecycleManager(volumeRegistry, poolRegistry, volume)
			if err := vlm.Delete(); err != nil {
				return err
			}
		}
	}

	return volumeRegistry.DeleteJob(token)
}

func CreatePerJobBuffer(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry, disk fileio.Disk,
	token string, user int, group int, capacity string, caller string, jobFile string, nodeFile string) error {
	summary, err := ParseJobFile(disk, jobFile)
	if err != nil {
		return err
	}

	if nodeFile != "" {
		// TODO we could add this into the volume as a scheduling hint, when its available?
		log.Printf("Ignoring nodeFile in setup: %s", nodeFile)
	}

	pool, bricksRequired, err := getPoolAndBrickCount(poolRegistry, capacity)
	if err != nil {
		return err
	}

	createdAt := uint(time.Now().Unix())
	job := registry.Job{
		Name:      token,
		Owner:     uint(user),
		CreatedAt: createdAt,
	}

	var perJobVolume *registry.Volume
	if bricksRequired > 0 && summary.PerJobBuffer != nil {
		perJobVolume = getPerJobVolume(token, pool, bricksRequired,
			user, group, caller, createdAt, summary)

		err := volumeRegistry.AddVolume(*perJobVolume)
		if err != nil {
			return err
		}

		job.JobVolume = perJobVolume.Name
	}

	for _, attachment := range summary.Attachments {
		name := registry.VolumeName(attachment.Name)
		volume, err := volumeRegistry.Volume(name)
		if err != nil {
			return err
		}
		// TODO: need to check permissions and not just go for it!
		if !volume.MultiJob {
			return fmt.Errorf("%s is not a multijob volume", volume.Name)
		}
		job.MultiJobVolumes = append(job.MultiJobVolumes, volume.Name)
	}

	job.Paths = setPaths(perJobVolume, job)

	err = volumeRegistry.AddJob(job)
	if err != nil {
		if job.JobVolume != "" {
			volumeRegistry.DeleteVolume(job.JobVolume)
		}
		return err
	}

	if job.JobVolume != "" {
		volume, err := volumeRegistry.Volume(job.JobVolume)
		vlm := lifecycle.NewVolumeLifecycleManager(volumeRegistry, poolRegistry, volume)
		err = vlm.ProvisionBricks()
		if err != nil {
			log.Println("Bricks may be left behnd, not deleting volume due to: ", err)
			return err
		}
	}
	return nil
}

func setPaths(perJobVolume *registry.Volume, job registry.Job) map[string]string {
	paths := make(map[string]string)
	if perJobVolume != nil {
		if perJobVolume.AttachPrivateNamespace {
			paths["DW_JOB_PRIVATE"] = fmt.Sprintf("/dac/%s_job_private", job.Name)
		}
		if perJobVolume.AttachGlobalNamespace {
			paths["DW_JOB_STRIPED"] = fmt.Sprintf("/dac/%s_job/global", job.Name)
		}
	}
	for _, multiJobVolume := range job.MultiJobVolumes {
		paths[fmt.Sprintf("DW_PERSISTENT_STRIPED_%s", multiJobVolume)] = fmt.Sprintf(
			"/dac/%s_persistent_%s", job.Name, multiJobVolume)
	}
	return paths
}

func getPerJobVolume(token string, pool *registry.Pool, bricksRequired uint,
	user int, group int, caller string, createdAt uint, summary jobSummary) *registry.Volume {
	adjustedSizeGB := bricksRequired * pool.GranularityGB
	perJobVolume := registry.Volume{
		Name:       registry.VolumeName(token),
		MultiJob:   false,
		State:      registry.Registered,
		Pool:       pool.Name,
		SizeBricks: bricksRequired,
		SizeGB:     adjustedSizeGB,
		JobName:    token,
		Owner:      uint(user),
		Group:      uint(group),
		CreatedBy:  caller,
		CreatedAt:  createdAt,
	}
	jobBuffer := summary.PerJobBuffer
	if jobBuffer.BufferType == scratch &&
		(jobBuffer.AccessMode == private || jobBuffer.AccessMode == privateAndStriped) {
		perJobVolume.AttachPrivateNamespace = true
	}
	if jobBuffer.BufferType == scratch &&
		(jobBuffer.AccessMode == striped || jobBuffer.AccessMode == privateAndStriped) {
		perJobVolume.AttachGlobalNamespace = true
	}
	if jobBuffer.BufferType == scratch && summary.Swap != nil {
		perJobVolume.AttachAsSwapBytes = uint(summary.Swap.SizeBytes)
	}
	// TODO that can be many data_in and data_out, we only allow one relating to striped job buffer
	if summary.DataIn != nil && summary.DataIn.Source != "" {
		// TODO check destination includes striped buffer path?
		perJobVolume.StageIn.Source = summary.DataIn.Source
		perJobVolume.StageIn.Destination = summary.DataIn.Destination
		switch summary.DataIn.StageType {
		case file:
			perJobVolume.StageIn.SourceType = registry.File
		case directory:
			perJobVolume.StageIn.SourceType = registry.Directory
		}
	}
	if summary.DataOut != nil && summary.DataOut.Source != "" {
		// TODO check source includes striped buffer path?
		perJobVolume.StageOut.Source = summary.DataOut.Source
		perJobVolume.StageOut.Destination = summary.DataOut.Destination
		switch summary.DataOut.StageType {
		case file:
			perJobVolume.StageOut.SourceType = registry.File
		case directory:
			perJobVolume.StageOut.SourceType = registry.Directory
		}
	}
	return &perJobVolume
}
