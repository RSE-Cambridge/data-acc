package datamodel

// You can only have zero or one allocations records for each Brick
type BrickAllocation struct {
	// Brick that is allocated
	Brick BrickInfo

	// Name of the volume that owns the brick
	AllocatedVolume VolumeName

	// 0 index allocation is the primary brick,
	// which is responsible for provisioning the associated volume
	AllocatedIndex uint

	// One primary prick per volume
	// this brick is responsible for watching for
	// associated job and volume actions
	IsPrimaryBrick bool

	// If any allocation sent to deallocate has a host that isn't
	// alive, this flag is set rather than have allocations removed.
	// A host should check for any allocations
	DeallocateRequested bool
}
