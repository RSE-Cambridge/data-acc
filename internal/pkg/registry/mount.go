package registry

type Mount struct {
	Hostname         string
	Config           MountConfig
	Mounted          bool
	UnmountRequested bool
}

type MountConfig struct {
	Filesystem string
	Host       string
	Path       string
	Options    []string
}
