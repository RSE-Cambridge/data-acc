package registry

type Pool struct {
	Name            string
	TotalSlices     uint
	FreeSlices      uint
	SliceCapacityGB uint
	Hosts           []Host
}
