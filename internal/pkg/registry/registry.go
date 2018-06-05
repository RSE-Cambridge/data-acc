package registry

type Registry interface {
	PoolRegistry

	BufferRegistry

	Buffers() ([]Buffer, error)
	Buffer(name string) (Buffer, error)
}

type BufferRegistry interface {
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
