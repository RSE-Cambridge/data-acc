package datamodel

// You can only have zero or one allocations records for each Brick
type BrickAllocation struct {
	// Brick that is allocated
	Brick Brick

	// Name of the session that owns the brick
	Session SessionName

	// 0 index allocation is the primary brick,
	// which is responsible for provisioning the associated volume
	AllocatedIndex uint
}
