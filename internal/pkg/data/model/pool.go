package model

type PoolName string

type Pool struct {
	// The pool is derived from all the reported bricks
	Name PoolName

	// This is the allocation unit for the pool
	// It is the minimum size of any registered brick
	GranularityGB uint
}

type PoolInfo struct {
	Pool Pool

	// Bricks from alive hosts
	AvailableBricks []BrickInfo

	// All currently active bricks
	AllocatedBricks []BrickAllocation
}
