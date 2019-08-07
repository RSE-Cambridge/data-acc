package model

import (
	"bytes"
	"encoding/json"
)

type VolumeName string

type Volume struct {
	// e.g. job1 or Foo
	Name VolumeName

	// its 8 characters long, so works nicely with lustre
	UUID string

	// True if multiple jobs can attach to this volume
	MultiJob bool

	// Requested pool of bricks for volume
	Pool string // TODO: PoolName?
	// Number of bricks requested, calculated from requested capacity
	SizeBricks uint
	// Actual size of the volume
	SizeGB uint

	// Back reference to what job created this volume
	JobName string
	// e.g. 1001
	Owner uint
	// If empty defaults to User
	Group uint
	// e.g. SLURM or Manila
	CreatedBy string
	// The unix (utc) timestamp of when this volume was created
	CreatedAt uint

	// TODO: need to fill these in...
	// They all related to how the volume is attached

	// All current attachments
	Attachments []Attachment
	// Attach all attachments to a shared global namespace
	// Allowed for any volume type
	AttachGlobalNamespace bool
	// Have an attachment specific namespace mounted, only for non multi job
	AttachPrivateNamespace bool
	// If not zero, swap of the requested amount mounted for each attachment
	// Not allowed for multi job
	AttachAsSwapBytes uint
	// Add attachment specific cache for each given filesystem path
	// Not allowed for multi job
	// Note: assumes the same path is cached for all attachments
	AttachPrivateCache []string

	// TODO: maybe data copy should be a slice associated with the job?
	// Request certain files to be staged in
	// Not currently allowed for multi job volumes
	StageInRequests []DataCopyRequest
	// Request certain files to be staged in
	// Not currently allowed for multi job volumes
	StageOutRequests []DataCopyRequest

	// BeeGFS wants each fs to be assigned a unique port number
	ClientPort int

	// Track if we have had bricks assigned
	// if we request delete, no bricks ever assigned, don't ait for dacd!
	HadBricksAssigned bool
}

type Attachment struct {
	// Hostname, Job and Volume name uniquely identify an attachment
	Hostname string

	// Associated jobName
	Job string

	State AttachmentState

	// If any error happened, it is reported here
	Error error
}

type AttachmentState int

const (
	UnknownAttachmentState AttachmentState = iota
	RequestAttach
	Attached
	RequestDetach
	Detached        AttachmentState = 400 // all bricks correctly deprovisioned unless host down or gone to ERROR
	AttachmentError AttachmentState = 500
)

var attachStateStrings = map[AttachmentState]string{
	UnknownAttachmentState: "",
	RequestAttach:          "RequestAttach",
	Attached:               "Attached",
	RequestDetach:          "RequestDetach",
	Detached:               "Detached",
	AttachmentError:        "AttachmentError",
}
var stringToAttachmentState = map[string]AttachmentState{
	"":                UnknownAttachmentState,
	"RequestAttach":   RequestAttach,
	"Attached":        Attached,
	"RequestDetach":   RequestDetach,
	"Detached":        Detached,
	"AttachmentError": AttachmentError,
}

func (attachmentState AttachmentState) String() string {
	return attachStateStrings[attachmentState]
}

func (attachmentState AttachmentState) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(attachStateStrings[attachmentState])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (attachmentState *AttachmentState) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	*attachmentState = stringToAttachmentState[str]
	return nil
}
