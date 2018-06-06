package registry

type VolumeRegistry interface {
	// Get all registered jobs and their volumes
	Jobs() ([]Job, error)

	// Get information about specific volume
	Volume(name VolumeName) (Volume, error)

	DeleteVolume(volume Volume)

	// Get all callback on all volume changes
	// If the volume is new, old = nil
	// used by the primary brick to get volume updates
	WatchVolumeChanges(volumeName string, callback func(old Volume, new Volume)) error

	// Creates volume and returns once volume is created
	// At this point all bricks are assigned and configured
	// And the list of Paths is ready
	WaitForVolumeReady(volume Volume) error

	WaitForVolumeDataIn(volumeName string) error
	WaitForVolumeAttached(volumeName string) error
	WaitForVolumeDetached(volumeName string) error
	WaitForVolumeDataOut(volumeName string) error
	WaitForVolumeDeleted(volumeName string) error
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
	readyForAttach  bool
	dataInRequested bool
	dataInComplete  bool
}

// TODO: define constants
type VolumeDriver string

type Configuration struct {
	// Define if used as a transparent cache or
	// if attached directly
	Type ConfigurationType

	// Size of the configuration
	// 0 means unrestricted, could consume whole volume
	// >0 means statically allocated space from volume
	Size uint

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
	Filesytem ConfigurationType = "filesystem"
	Cache     ConfigurationType = "cache"
	Swap      ConfigurationType = "swap"
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
