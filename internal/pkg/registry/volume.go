package registry

import (
	"bytes"
	"encoding/json"
)

type VolumeRegistry interface {
	// Get all registered jobs and their volumes
	Jobs() ([]Job, error)

	// Add job and associated volumes
	// Fails to add job if volumes are in a bad state
	AddJob(job Job) error

	// Remove job from the system
	// TODO: fails if volumes are not in the deleted state?
	DeleteJob() error

	// Get information about specific volume
	// TODO: remove add/detele and only
	AddVolume(volume Volume) error

	// Get information about a specific volume
	Volume(name VolumeName) (Volume, error)

	// TODO: this should error if volume is not in correct state?
	DeleteVolume(name VolumeName) error

	// Move between volume states, but only one by one
	UpdateState(name VolumeName, state VolumeState) error

	// Used to add or remove attachments
	// TODO: only allowed in certain states?
	UpdateConfiguration(name VolumeName, configuration []Configuration) error

	// Get all callback on all volume changes
	// If the volume is new, old = nil
	// used by the primary brick to get volume updates
	WatchVolumeChanges(volumeName string, callback func(old *Volume, new *Volume)) error
}

type Job struct {
	// Name of the job
	Name string

	// Zero or One PerJob volumes
	// and Zero or more MultiJob volumes
	Volumes []VolumeName
}

type VolumeName string

// Volume information
// To get assigned bricks see PoolRegistry
type Volume struct {
	// e.g. job1 or Foo
	Name VolumeName
	// e.g. 1001
	Owner int
	// If empty defaults to User
	Group int
	// e.g. SLURM or Manila
	CreatedBy string
	// Requested pool of bricks for volume
	Pool string
	// Requested size of volume
	SizeGB uint
	// True if multiple jobs can attach to this volume
	MultiJob bool

	// Current uses of the volume capacity and its attachments
	Configurations []Configuration

	// Volume drivers e.g. Lustre, Lustre+Loopback,
	// BeeGFS, NVMe-over-Fabrics, etc
	Driver VolumeDriver

	// TODO:....
	Paths []string

	// TODO: track state machine...
	State VolumeState
}

func (volume Volume) String() string {
	rawVolume, _ := json.Marshal(volume)
	return string(rawVolume)
}

type VolumeState int

const (
	Unknown VolumeState = iota
	Registered
	BricksAssigned
	Test2
	Test3
	Ready   VolumeState = 200
	Deleted VolumeState = 400
	Error   VolumeState = 500
)

var volumeStateStrings = map[VolumeState]string{
	Unknown:        "",
	Registered:     "Registered",
	BricksAssigned: "BricksAssigned",
	Test2:          "Test2",
	Test3:          "Test3",
}
var stringToVolumeState = map[string]VolumeState{
	"":               Unknown,
	"Registered":     Registered,
	"BricksAssigned": BricksAssigned,
	"Test2":          Test2,
	"Test3":          Test3,
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
}

type SourceType string

const (
	File      SourceType = "file"
	Directory SourceType = "directory"
	List      SourceType = "list"
)

type Attachment struct {
	Hostname        string
	Attached        bool
	DetachRequested bool
}
