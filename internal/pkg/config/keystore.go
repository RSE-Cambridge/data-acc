package config

import (
	"log"
	"strings"
)

type KeystoreConfig struct {
	Endpoints []string
	CertFile  string
	KeyFile   string
	CAFile    string
}

func GetKeystoreConfig(env ReadEnvironemnt) KeystoreConfig {
	config := KeystoreConfig{
		CertFile: getString(env, "ETCDCTL_CERT_FILE", ""),
		KeyFile:  getString(env, "ETCDCTL_KEY_FILE", ""),
		CAFile:   getString(env, "ETCDCTL_CA_FILE", ""),
	}
	endpointsStr := getString(env, "ETCDCTL_ENDPOINTS", "")
	if endpointsStr == "" {
		endpointsStr = getString(env, "ETCD_ENDPOINTS", "")
	}
	if endpointsStr == "" {
		log.Fatalf("Must set ETCDCTL_ENDPOINTS environment variable, e.g. export ETCDCTL_ENDPOINTS=127.0.0.1:2379")
	}
	config.Endpoints = strings.Split(endpointsStr, ",")
	return config
}
