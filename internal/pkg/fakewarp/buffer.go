package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"time"
)

func DeleteBufferComponents(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry,
	token string) error {

	volumeName := registry.VolumeName(token)
	volume, err := volumeRegistry.Volume(volumeName)
	if err != nil {
		// TODO should check this error relates to the volume being missing
		log.Println(err)
		return nil
	}

	vlm := lifecycle.NewVolumeLifecycleManager(volumeRegistry, poolRegistry, volume)
	if err := vlm.Delete(); err != nil {
		return err
	}

	return volumeRegistry.DeleteJob(token)
}

func CreatePerJobBuffer(volumeRegistry registry.VolumeRegistry, poolRegistry registry.PoolRegistry, disk fileio.Disk,
	token string, user int, group int, capacity string, caller string, jobFile string) error {
	summary, err := ParseJobFile(disk, jobFile)
	if err != nil {
		return err
	} else {
		log.Println("Summary of job file:", summary)
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
	}
	if bricksRequired > 0 && summary.PerJobBuffer != nil {
		job.JobVolume = registry.VolumeName(token)
		perJobVolume := registry.Volume{
			Name:       registry.VolumeName(token),
			MultiJob:   false,
			Pool:       pool.Name,
			SizeBricks: bricksRequired,
			SizeGB:     adjustedSizeGB,
			JobName:    token,
			Owner:      uint(user),
			Group:      uint(group),
			CreatedBy:  caller,
			CreatedAt:  createdAt,
			AttachGlobalNamespace: summary.PerJobBuffer.BufferType == scratch &&
				summary.PerJobBuffer.AccessMode != private,
			AttachPrivateNamespace: summary.PerJobBuffer.BufferType == scratch &&
				summary.PerJobBuffer.AccessMode != striped,
		}
		if summary.PerJobBuffer.BufferType == scratch && summary.Swap != nil {
			perJobVolume.AttachAsSwapBytes = uint(summary.Swap.SizeBytes)
		}
		// TODO data in and data out, maybe do this later anyways?
		err := volumeRegistry.AddVolume(perJobVolume)
		if err != nil {
			return err
		}
	}
	for _, attachment := range summary.Attachments {
		name := registry.VolumeName(attachment.Name)
		_, err := volumeRegistry.Volume(name)
		if err != nil {
			return err
		}
		job.MultiJobVolumes = append(job.MultiJobVolumes, name)
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
