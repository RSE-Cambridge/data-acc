package oldregistry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"time"
)

type BufferRegistry interface {
	AddBuffer(buffer Buffer) error
	UpdateBuffer(buffer Buffer) (Buffer, error)
	RemoveBuffer(buffer Buffer)
}

type BufferWatcher interface {
	WatchNewBuffers(callback func(buffer Buffer))
	WatchBuffer(bufferName string, callback func(buffer Buffer))
}

type Buffer struct {
	// e.g. Buffer1
	Name string
	// e.g. userid 1001
	Owner string
	// e.g. SLURM
	CreatedBy       string
	CreatedAt       time.Time
	CapacityGB      uint
	Pool            string
	Bricks          []registry.BrickInfo // TODO should really be allocations
	Provisioned     bool
	DeleteRequested bool
}
