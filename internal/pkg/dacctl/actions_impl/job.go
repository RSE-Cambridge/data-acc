package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl"
	parsers2 "github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl/actions_impl/parsers"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"log"
	"sort"
)

func (d *dacctlActions) ValidateJob(c dacctl.CliContext) error {
	err := checkRequiredStrings(c, "job")
	if err != nil {
		return err
	}

	jobFile := c.String("job")
	summary, err := parsers2.ParseJobFile(d.disk, jobFile)
	if err != nil {
		return err
	} else {
		// TODO check valid pools, etc, etc.
		log.Println("Summary of job file:", summary)
	}
	return nil
}

func (d *dacctlActions) CreatePerJobBuffer(c dacctl.CliContext) error {
	checkRequiredStrings(c, "token", "job", "caller", "capacity")
	// TODO: need to specify user and group too

	jobFile := c.String("job")
	summary, err := parsers2.ParseJobFile(d.disk, jobFile)
	if err != nil {
		return err
	}

	nodeFile := c.String("nodehostnamefile")
	if nodeFile != "" {
		// TODO we could add this into the volume as a scheduling hint, when its available?
		log.Printf("Ignoring nodeFile in setup: %s", nodeFile)
	}

	pool, capacityBytes, err := parsers2.ParseCapacityBytes(c.String("capacity"))
	if err != nil {
		return err
	}

	// extract info from job file
	swapBytes := 0
	if summary.Swap != nil {
		swapBytes = summary.Swap.SizeBytes
	}
	access := datamodel.NoAccess
	bufferType := datamodel.Scratch
	if summary.PerJobBuffer != nil {
		access = summary.PerJobBuffer.AccessMode
		if summary.PerJobBuffer.BufferType != datamodel.Scratch {
			return fmt.Errorf("cache is not supported")
		}
	}
	var multiJobVolumes []datamodel.SessionName
	for _, attachment := range summary.Attachments {
		multiJobVolumes = append(multiJobVolumes, attachment)
	}

	request := datamodel.VolumeRequest{
		MultiJob:           false,
		Caller:             c.String("caller"),
		TotalCapacityBytes: capacityBytes,
		PoolName:           datamodel.PoolName(pool),
		Access:             access,
		Type:               bufferType,
		SwapBytes:          swapBytes,
	}
	// TODO: must be a better way!
	// ensure multi job volumes are sorted, to avoid deadlocks (*cough*)
	sort.Slice(multiJobVolumes, func(i, j int) bool {
		return multiJobVolumes[i] < multiJobVolumes[j]
	})
	session := datamodel.Session{
		Name:                datamodel.SessionName(c.String("token")),
		Owner:               uint(c.Int("user")),
		Group:               uint(c.Int("group")),
		CreatedAt:           getNow(),
		VolumeRequest:       request,
		MultiJobAttachments: multiJobVolumes,
		StageInRequests:     summary.DataIn,
		StageOutRequests:    summary.DataOut,
	}
	session.Paths = getPaths(session)
	return d.session.CreateSession(session)
}

func getPaths(session datamodel.Session) map[string]string {
	paths := make(map[string]string)
	if session.VolumeRequest.MultiJob == false {
		if session.VolumeRequest.Access == datamodel.Private || session.VolumeRequest.Access == datamodel.PrivateAndStriped {
			paths["DW_JOB_PRIVATE"] = fmt.Sprintf("/dac/%s_job_private", session.Name)
		}
		if session.VolumeRequest.Access == datamodel.Striped || session.VolumeRequest.Access == datamodel.PrivateAndStriped {
			paths["DW_JOB_STRIPED"] = fmt.Sprintf("/dac/%s_job/global", session.Name)
		}
	}
	for _, multiJobVolume := range session.MultiJobAttachments {
		paths[fmt.Sprintf("DW_PERSISTENT_STRIPED_%s", multiJobVolume)] = fmt.Sprintf(
			"/dac/%s_persistent_%s/global", session.Name, multiJobVolume)
	}
	return paths
}
