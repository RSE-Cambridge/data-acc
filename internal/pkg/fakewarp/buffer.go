package fakewarp

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"time"
)

func DeleteBufferComponents(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	token string) error {

	job, err := volumeRegistry.Job(token)
	if err != nil {
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
	token string, user int, group int, capacity string, caller string, jobFile string) error {
	summary, err := ParseJobFile(disk, jobFile)
	if err != nil {
		return err
	}

	pool, bricksRequired, err := getPoolAndBrickCount(poolRegistry, capacity)
	if err != nil {
		return err
	}
	adjustedSizeGB := bricksRequired * pool.GranularityGB

	createdAt := uint(time.Now().Unix())
	job := registry.Job{
		Name:      token,
		Owner:     uint(user),
		CreatedAt: createdAt,
		Paths:     make(map[string]string),
	}

	for _, attachment := range summary.Attachments {
		name := registry.VolumeName(attachment.Name)
		_, err := volumeRegistry.Volume(name)
		if err != nil {
			return err
		}
		job.MultiJobVolumes = append(job.MultiJobVolumes, name)
		job.Paths[fmt.Sprintf("DW_PERSISTENT_STRIPED_%s", name)] = fmt.Sprintf(
			"/mnt/dac/job/%s/multijob/%s", job.Name, name)
	}

	if bricksRequired > 0 && summary.PerJobBuffer != nil {
		job.JobVolume = registry.VolumeName(token)
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
		if summary.PerJobBuffer.BufferType == scratch &&
			summary.PerJobBuffer.AccessMode != private {
			perJobVolume.AttachGlobalNamespace = true
			job.Paths["DW_JOB_PRIVATE"] = fmt.Sprintf("/mnt/dac/job/%s/private", job.Name)
		}
		if summary.PerJobBuffer.BufferType == scratch &&
			summary.PerJobBuffer.AccessMode != striped {
			perJobVolume.AttachPrivateNamespace = true
			job.Paths["DW_JOB_STRIPED"] = fmt.Sprintf("/mnt/dac/job/%s/global", job.Name)
		}
		if summary.PerJobBuffer.BufferType == scratch && summary.Swap != nil {
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
			perJobVolume.StageOut.Source = summary.DataIn.Source
			perJobVolume.StageOut.Destination = summary.DataIn.Destination
			switch summary.DataIn.StageType {
			case file:
				perJobVolume.StageIn.SourceType = registry.File
			case directory:
				perJobVolume.StageIn.SourceType = registry.Directory
			}
		}

		err := volumeRegistry.AddVolume(perJobVolume)
		if err != nil {
			return err
		}
	}

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
		err = vlm.ProvisionBricks(*pool)
		if err != nil {
			volumeRegistry.DeleteVolume(job.JobVolume)
			volumeRegistry.DeleteJob(job.Name)
			return err
		}
	}
	return nil
}
