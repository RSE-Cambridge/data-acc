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

	// Wait for a specific state, error returned if not possible
	WaitForState(name VolumeName, state VolumeState) error

	// Used to add or remove attachments
	// TODO: remove this??
	UpdateConfiguration(name VolumeName, configuration []Configuration) error

	// TODO: RequestVolumeAttachment(name VolumeName, hostnames string[])
	// TODO: RequestVolumeDetach(name VolumeName, hostnames string[])
	// TODO: RequestVolumeDataIn(name VolumeName, datain DataCopyRequest)
	// TODO: RequestVolumeDataOut(name VolumeName, datain DataCopyRequest)

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

	// TODO: remove once moved to above fields
	Volumes []VolumeName
}

type VolumeName string

// Volume information
// To get assigned bricks see PoolRegistry
type Volume struct {
	// e.g. job1 or Foo
	Name VolumeName
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

	// Each string contains an environment variable export
	// The paths handed to a job come from aggregating the paths
	// used by all volumes
	// TODO: should split into Name/Value pairs, or use a map?
	Paths []string

	//
	// TODO... delete all these fields, once they are no longer used!
	//

	// Current uses of the volume capacity and its attachments
	Configurations []Configuration

	// Volume drivers e.g. Lustre, Lustre+Loopback,
	// BeeGFS, NVMe-over-Fabrics, etc
	Driver VolumeDriver
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
	MountRequested
	MountComplete // compute nodes all mounted, or gone to ERROR
	UnmountRequested
	UnmountComplete // compute nodes all unmounted, or gone to ERROR
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
	MountRequested:    "MountRequested",
	MountComplete:     "MountComplete",
	UnmountRequested:  "UnmountRequested",
	UnmountComplete:   "UnmountComplete",
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
	"MountRequested":    MountRequested,
	"MountComplete":     MountComplete,
	"UnmountRequested":  UnmountRequested,
	"UnmountComplete":   UnmountComplete,
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

// TODO: define constants
type VolumeDriver string

// TODO: delete
type Configuration struct {
	Name string
	// Define if used as a transparent cache or
	// if attached directly
	Type ConfigurationType

	// Size of the configuration
	// 0 means unrestricted, could consume whole volume
	// >0 means statically allocated space from volume
	NamespaceSize uint

	// If true all attachments share the same namespace,
	// else each attachment gets a dedicated namespace
	// Note: content of dedicated namespace deleted
	// when attachment deleted
	SharedNamespace bool

	// All current attachments for this configuration
	Attachments []Attachment

	// Request certain files to be staged in and out
	// Currently only supported when SharedNamespace=True
	StageIn  DataCopyRequest
	StageOut DataCopyRequest

	// e.g. lustre + sparse file created for each attachment, etc..
	Driver AttachmentDriver
}

type AttachmentDriver string

type ConfigurationType string

const (
	Filesystem ConfigurationType = "filesystem"
	Cache      ConfigurationType = "cache"
	Swap       ConfigurationType = "swap"
)

type DataCopyRequest struct {
	// Source points to a File or a Directory,
	// or a file that contains a list of source and destinations,
	// with each pair on a new line
	SourceType SourceType
	// The path is either to a file or a directory or a
	Source string
	// Must be empty string for type list, otherwise specifes location
	Destination string
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

	// Report true when the mount has worked
	// Add attachment with false to request the mount
	Attached bool

	// Report if the detach was requested
	// Attachment is removed once detach is complete
	DetachRequested bool

	// If any error happened, it is reported here
	Error error
}
