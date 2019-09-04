package config

func GetDacctlLog() string {
	return getString(DefaultEnv, "DACCTL_LOG", "/var/log/dacctl.log")
}
