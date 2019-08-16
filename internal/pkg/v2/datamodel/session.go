package datamodel

type SessionName string

// This object is updated by dacctl
// And actions are sent relative to a Session
// and the primary brick waits for the session
type Session struct {
	// Job name or persistent buffer name
	Name SessionName

	// Currently stored revision
	// this is checked when an update is requested
	Revision int

	// unix uid
	Owner uint

	// unix group id
	Group uint

	// utc unix timestamp when buffer created
	CreatedAt uint

	// Details of what was requested
	VolumeRequest VolumeRequest

	// Flags about current state of the buffer
	Status SessionStatus

	// Request certain files to be staged in
	// Not currently allowed for multi job volumes
	StageInRequests []DataCopyRequest

	// Request certain files to be staged in
	// Not currently allowed for multi job volumes
	StageOutRequests []DataCopyRequest

	// There maybe be attachments to multiple shared volumes
	MultiJobAttachments []SessionName

	// Environment variables for each volume associated with the job
	Paths map[string]string

	// Resources used by session once pool granularity is taken into account
	ActualSizeBytes int

	// List of the bricks allocated to implement the JobVolume
	// One is the primary brick that should be watching for all actions
	Allocations []BrickAllocation

	// Where session requests should be sent
	PrimaryBrickHost BrickHostName

	// Compute hosts for this session
	// Note: should be empty for multi-job volumes
	RequestedAttachHosts []string

	// Used by filesystem provider to store internal state
	// and track if the filesystem had a recent error
	FilesystemStatus FilesystemStatus

	// For multi-job volumes these are always other sessions
	// for job volumes this is always for just this session
	CurrentAttachments map[SessionName]AttachmentSessionStatus
}

type FilesystemStatus struct {
	Error        error
	InternalName string
	InternalData string
}

type AttachmentSession struct {
	SessionName SessionName
	Hosts       []string
}

type AttachmentSessionStatus struct {
	AttachmentSession AttachmentSession

	GlobalMount  bool
	PrivateMount bool
	SwapBytes    int

	DetachRequested bool
	Error           error
}

type SessionStatus struct {
	// If not nil, the session has an unresolved error
	// and can't be mounted by any new sessions
	// but it can be deleted
	Error error

	// CreateVolume has succeeded, so other actions can now happen
	FileSystemCreated bool

	// Assuming one data in / data out cycle per job
	CopyDataInComplete  bool
	CopyDataOutComplete bool

	// Records if we have started trying to delete
	DeleteRequested bool

	// Records if we should skip copy data out on delete
	DeleteSkipCopyDataOut bool
}

type VolumeRequest struct {
	MultiJob           bool
	Caller             string
	TotalCapacityBytes int
	PoolName           PoolName
	Access             AccessMode
	Type               BufferType
	SwapBytes          int
}

type AccessMode int

const (
	NoAccess AccessMode = iota
	Striped
	Private
	PrivateAndStriped
)

type BufferType int

const (
	Scratch BufferType = iota
	Cache
)

type DataCopyRequest struct {
	// Source points to a File or a Directory,
	// or a file that contains a list of source and destinations,
	// with each pair on a new line
	SourceType SourceType
	// The path is either to a file or a directory or
	// a list with source and destination file space separated pairs, each on a new line
	Source string
	// Must be empty string for type list, otherwise specified location
	Destination string
	// Used to notify if copy in has been requested
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
	// Provided a file that has source and destination file space separated pairs, each on a new line
	List SourceType = "list"
)
