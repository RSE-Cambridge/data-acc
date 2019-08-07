package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/parsers/capacity"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/parsers/jobfile"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"log"
)

func (d *dacctlActions) ValidateJob(c actions.CliContext) error {
	checkRequiredStrings(c, "job")
	jobFile := c.String("job")
	summary, err := jobfile.ParseJobFile(d.disk, jobFile)
	if err != nil {
		return err
	} else {
		// TODO check valid pools, etc, etc.
		log.Println("Summary of job file:", summary)
	}
	return nil
}

func (d *dacctlActions) CreatePerJobBuffer(c actions.CliContext) error {
	checkRequiredStrings(c, "token", "job", "caller", "capacity")
	// TODO: need to specify user and group too

	jobFile := c.String("job")
	summary, err := jobfile.ParseJobFile(d.disk, jobFile)
	if err != nil {
		return err
	}

	nodeFile := c.String("nodehostnamefile")
	if nodeFile != "" {
		// TODO we could add this into the volume as a scheduling hint, when its available?
		log.Printf("Ignoring nodeFile in setup: %s", nodeFile)
	}

	pool, capacityBytes, err := capacity.ParseCapacityBytes(c.String("capacity"))
	if err != nil {
		return err
	}

	// extract info from job file
	swapBytes := 0
	if summary.Swap != nil {
		swapBytes = summary.Swap.SizeBytes
	}
	access := model.NoAccess
	bufferType := model.Scratch
	if summary.PerJobBuffer != nil {
		access = summary.PerJobBuffer.AccessMode
		if summary.PerJobBuffer.BufferType == model.Cache {
			return fmt.Errorf("cache is not supported")
		}
	}
	var multiJobVolumes []model.VolumeName
	for _, attachment := range summary.Attachments {
		multiJobVolumes = append(multiJobVolumes, attachment)
	}

	request := model.VolumeRequest{
		MultiJob:           false,
		Caller:             c.String("caller"),
		TotalCapacityBytes: capacityBytes,
		PoolName:           pool,
		Access:             access,
		Type:               bufferType,
		SwapBytes:          swapBytes,
	}
	session := model.Session{
		Name:             model.SessionName(c.String("token")),
		Owner:            uint(c.Int("user")),
		Group:            uint(c.Int("group")),
		CreatedAt:        getNow(),
		VolumeRequest:    request,
		MultiJobVolumes:  multiJobVolumes,
		StageInRequests:  summary.DataIn,
		StageOutRequests: summary.DataOut,
	}
	session.Paths = getPaths(session)

	session, err = d.registry.CreateSessionAllocations(session)
	if err != nil {
		return err
	}
	return d.actions.CreateSessionVolume(session)
}

func getPaths(session model.Session) map[string]string {
	paths := make(map[string]string)
	if session.VolumeRequest.MultiJob == false {
		if session.VolumeRequest.Access == model.Private || session.VolumeRequest.Access == model.PrivateAndStriped {
			paths["DW_JOB_PRIVATE"] = fmt.Sprintf("/dac/%s_job_private", session.Name)
		}
		if session.VolumeRequest.Access == model.Striped || session.VolumeRequest.Access == model.PrivateAndStriped {
			paths["DW_JOB_STRIPED"] = fmt.Sprintf("/dac/%s_job/global", session.Name)
		}
	}
	for _, multiJobVolume := range session.MultiJobVolumes {
		paths[fmt.Sprintf("DW_PERSISTENT_STRIPED_%s", multiJobVolume)] = fmt.Sprintf(
			"/dac/%s_persistent_%s", session.Name, multiJobVolume)
	}
	return paths
}
