package datamodel

type SessionName string

type Session struct {
	// Job name or persistent buffer name
	Name SessionName

	// unix uid and gid
	Owner     uint
	Group     uint

	// utc unix timestamp when buffer created
	CreatedAt uint

	// The hosts that want to mount the storage
	// Note: to allow for copy in/out the brick hosts are assumed to have an attachment
	AttachHosts []string

	// If non-zero capacity requested, a volume is created for this job
	// It may be exposed to the attach hosts in a variety of ways, as defined by the volume
	JobVolume VolumeName

	// There maybe be attachments to multiple shared volumes
	MultiJobVolumes []VolumeName

	// Environment variables for each volume associated with the job
	Paths map[string]string
}