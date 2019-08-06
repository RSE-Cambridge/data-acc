package pfsprovider

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
)

// A plugin must provide implementations for both interfaces
type Plugin interface {
	Mounter() Mounter
	VolumeProvider() VolumeProvider
}

// Actions on the host assigned to the primary brick
type VolumeProvider interface {
	SetupVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error
	TeardownVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error

	CopyDataIn(volume registry.Volume) error
	CopyDataOut(volume registry.Volume) error
}

// Actions that are sent to remote hosts,
// typically compute nodes and primary brick hosts
type Mounter interface {
	Mount(volume registry.Volume, brickAllocations []registry.BrickAllocation, attachments []registry.Attachment) error
	Unmount(volume registry.Volume, brickAllocations []registry.BrickAllocation, attachments []registry.Attachment) error
}
