package config

type FilesystemConfig struct {
	MGSDevice   string
	MGSHost     string
	MaxMDTs     uint
	HostGroup   string
	AnsibleDir  string
	SkipAnsible bool
	LnetSuffix  string
	MDTSizeMB   uint
}

func GetFilesystemConfig() FilesystemConfig {
	env := DefaultEnv
	conf := FilesystemConfig{
		MGSDevice:   getString(env, "DAC_MGS_DEV", "sdb"),
		MGSHost:     getString(env, "DAC_MGS_HOST", "localhost"),
		MaxMDTs:     getUint(env, "DAC_MAX_MDT_COUNT", 24),
		HostGroup:   getString(env, "DAC_HOST_GROUP", "dac-prod"),
		AnsibleDir:  getString(env, "DAC_ANSIBLE_DIR", "/var/lib/data-acc/fs-ansible/"),
		SkipAnsible: getBool(env, "DAC_SKIP_ANSIBLE", false),
		LnetSuffix:  getString(env, "DAC_LNET_SUFFIX", ""),
	}
	mdtSizeMB := getUint(env, "DAC_MDT_SIZE_GB", 0) * 1024
	if mdtSizeMB == 0 {
		mdtSizeMB = getUint(env, "DAC_MDT_SIZE_MB", uint(20*1024))
	}
	conf.MDTSizeMB = mdtSizeMB
	return conf
}
