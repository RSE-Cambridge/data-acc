package registry

import "time"

type Buffer struct {
	// e.g. Buffer1
	Name string
	// e.g. userid 1001
	Owner string
	// e.g. SLURM
	CreatedBy         string
	CreatedAt         time.Time
	CapacityGB        uint
	Pool              Pool
	Bricks            []BrickInfo // TODO should really be allocations
	Mounts            []Mount
	Provisioned       bool
	DeleteRequested   bool
	AttachmentDetails AttachmentDetails
}

func (buffer *Buffer) ReadyToMount() bool {
	return buffer.Provisioned && buffer.Mounts != nil && !buffer.DeleteRequested
}

type AttachmentDetails struct {
	Type      Type
	StageIn   FileRequest
	StageOut  FileRequest
	MountMode MountMode
	// Does it target a single job, or is it more persistent
	Transient bool
}

type FileRequest struct {
	InputPath  string
	OutputPath string
}

type Type int

const (
	Scratch Type = iota
	Cache
)

type MountMode int

const (
	Private MountMode = iota
	Global
)
