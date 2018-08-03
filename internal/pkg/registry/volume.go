package registry

import (
	"bytes"
	"encoding/json"
)

type VolumeRegistry interface {
	// Get all registered jobs and their volumes
	Jobs() ([]Job, error)

	// Get a specific job
	Job(jobName string) (Job, error)

	// Add job and associated volumes
	// Fails to add job if volumes are in a bad state
	AddJob(job Job) error

	// Update specified job with given hosts
	// Fails if the job already has any hosts associated with it
	JobAttachHosts(jobName string, hosts []string) error

	// Remove job from the system
	// TODO: fails if volumes are not in the deleted state?
	DeleteJob(jobName string) error

	// Get information about specific volume
	// TODO: remove add/detele and only
	AddVolume(volume Volume) error

	// Get information about a specific volume
	Volume(name VolumeName) (Volume, error)

	// Get information about all volumes
	AllVolumes() ([]Volume, error)

	// TODO: this should error if volume is not in correct state?
	DeleteVolume(name VolumeName) error

	// Move between volume states, but only one by one
	UpdateState(name VolumeName, state VolumeState) error

	// Update all the specified attachments
	// if attachment doesn't exist, attachment is added
	UpdateVolumeAttachments(name VolumeName, attachments map[string]Attachment) error

	// Wait for a specific state, error returned if not possible
	WaitForState(name VolumeName, state VolumeState) error

	// Wait for a given condition
	// TODO: remove wait for state?
	WaitForCondition(volumeName VolumeName, condition func(old *Volume, new *Volume) bool) error

	// Get all callback on all volume changes
	// If the volume is new, old = nil
	// used by the primary brick to get volume updates
	WatchVolumeChanges(volumeName string, callback func(old *Volume, new *Volume)) error
}

// TODO: Attachment request, or session is probably a better name here...
type Job struct {
	// Name of the job
	Name      string // TODO: should we make a JobName type?
	Owner     uint
	CreatedAt uint

	// The hosts that want to mount the storage
	// Note: to allow for copy in/out the brick hosts are assumed to have an attachment
	AttachHosts []string

	// If non-zero capacity requested, a volume is created for this job
	// It may be exposed to the attach hosts in a variety of ways, as defined by the volume
	JobVolume VolumeName

	// There maybe be attachments to multiple shared volumes
	MultiJobVolumes []VolumeName

	// Environment variables for each volume associated with the job
	Paths map[string]string
}

type VolumeName string

// Volume information
// To get assigned bricks see PoolRegistry
type Volume struct {
	// e.g. job1 or Foo
	Name VolumeName
	// its 8 characters long, so works nicely with lustre
	UUID string
	// True if multiple jobs can attach to this volume
	MultiJob bool

	// Message requested actions to primary brick host
	// TODO: move mount and data copy actions to other parts of the volume state
	State VolumeState

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

	// All current attachments, by hostname
	Attachments map[string]Attachment
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
	StageIn DataCopyRequest
	// Request certain files to be staged in
	// Not currently allowed for multi job volumes
	StageOut DataCopyRequest

	// TODO: data model currently does not do these things well:
	// 1. correctly track multiple jobs at the same time attach to the same persistent buffer
	// 2. data in/out requests for persistent buffer
	// 3. track amount of space used by swap and/or metadata
}

func (volume Volume) String() string {
	rawVolume, _ := json.Marshal(volume)
	return string(rawVolume)
}

type VolumeState int

const (
	Unknown VolumeState = iota
	Registered
	BricksProvisioned // setup waits for this, updated by host manager, paths should be setup, or gone to ERROR
	DataInRequested
	DataInComplete // data_in waits for host manager to data in, or gone to ERROR
	DataOutRequested
	DataOutComplete             // data copied out by host manager, or gone to ERROR
	DeleteRequested VolumeState = 399
	BricksDeleted   VolumeState = 400 // all bricks correctly deprovisioned unless host down or gone to ERROR
	Error           VolumeState = 500
)

var volumeStateStrings = map[VolumeState]string{
	Unknown:           "",
	Registered:        "Registered",
	BricksProvisioned: "BricksProvisioned",
	DataInRequested:   "DataInRequested",
	DataInComplete:    "DataInComplete",
	DataOutRequested:  "DataOutRequested",
	DataOutComplete:   "DataOutComplete",
	DeleteRequested:   "DeleteRequested",
	BricksDeleted:     "BricksDeleted",
	Error:             "Error",
}
var stringToVolumeState = map[string]VolumeState{
	"":                  Unknown,
	"Registered":        Registered,
	"BricksProvisioned": BricksProvisioned,
	"DataInRequested":   DataInRequested,
	"DataInComplete":    DataInComplete,
	"DataOutRequested":  DataOutRequested,
	"DataOutComplete":   DataOutComplete,
	"DeleteRequested":   DeleteRequested,
	"BricksDeleted":     BricksDeleted,
	"Error":             Error,
}

func (volumeState VolumeState) String() string {
	return volumeStateStrings[volumeState]
}

func (volumeState VolumeState) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(volumeStateStrings[volumeState])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (volumeState *VolumeState) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	*volumeState = stringToVolumeState[str]
	return nil
}

type DataCopyRequest struct {
	// Source points to a File or a Directory,
	// or a file that contains a list of source and destinations,
	// with each pair on a new line
	SourceType SourceType
	// The path is either to a file or a directory or a
	Source string
	// Must be empty string for type list, otherwise specifes location
	Destination string
	// Used to notify if copy in has been requested
	// TODO: remove volume states and update this instead
	RequestCopyIn bool
	// Report if the copy has completed
	CopyCompleted bool
	// if there was problem, record it
	Error error
}

type SourceType string

const (
	File      SourceType = "file"
	Directory SourceType = "directory"
	// Provide a file that has source and destination file space separated pairs, each on a new line
	List SourceType = "list"
)

type Attachment struct {
	// Hostname and Volume name uniquely identify an attachment
	Hostname string

	// TODO: delete three fields above
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
