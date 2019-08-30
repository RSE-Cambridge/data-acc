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
		log.Println("ETCD_ENDPOINTS is deprecated please use ETCDCTL_ENDPOINTS")
		endpointsStr = getString(env, "ETCD_ENDPOINTS", "")
	}
	if endpointsStr == "" {
		log.Fatalf("Must set ETCDCTL_ENDPOINTS environemnt variable, e.g. export ETCDCTL_ENDPOINTS=127.0.0.1:2379")
	}
	config.Endpoints = strings.Split(endpointsStr, ",")
	return config
}
