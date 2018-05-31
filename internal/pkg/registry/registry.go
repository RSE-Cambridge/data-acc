package registry

type Registry interface {
	Buffers() ([]Buffer, error)
	Buffer(name string) (Buffer, error)

	Pools() ([]Pool, error)
	Bricks() ([]Brick, error)
	BricksForHost(host string) ([]Brick, error)

	// Update host, or add if not already present
	UpdateHost(host Host) error
	AddBuffer(buffer Buffer) error
	UpdateBuffer(buffer Buffer) (Buffer, error)
	RemoveBuffer(buffer Buffer)
}

type BufferUpdater interface {
	ProcessBufferRequest(buffer Buffer) (Buffer, error)
	ProvisionBuffer(buffer Buffer)
	StageIn(buffer Buffer)
	StageOut(buffer Buffer)
	AttachBuffer(buffer Buffer, hostnames []string)
}

type BufferWatcher interface {
	WatchNewBuffers(callback func(buffer Buffer))
	WatchBuffer(bufferName string, callback func(buffer Buffer))
}

type BrickWatcher interface {
	WatchBrick(brickUuid string, callback func(brick Brick))
}

type HostReporter interface {
	// Generally run this in a goroutine to signal the host is being actively managed
	KeepAliveHost(hostname string)
}
