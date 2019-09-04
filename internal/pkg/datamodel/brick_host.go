package datamodel

type BrickHostName string

type BrickHost struct {
	Name BrickHostName

	// Returns all bricks
	Bricks []Brick

	// True if allowing new volumes to use bricks from this host
	Enabled bool
}

type BrickHostStatus struct {
	BrickHost BrickHost

	// True is current keepalive key exists
	Alive bool
}

type Brick struct {
	// Bricks are identified by device and hostname
	// It must only contain the characters A-Za-z0-9
	// e.g. sdb, not /dev/sdb
	Device string

	// It must only contain the characters "A-Za-z0-9."
	BrickHostName BrickHostName

	// The bool a brick is associated with
	// It must only contain the characters A-Za-z0-9
	PoolName PoolName

	// TODO: move this to bytes, and make bytes a special type?
	// Size of the brick, defines the pool granularity
	CapacityGiB uint
}
