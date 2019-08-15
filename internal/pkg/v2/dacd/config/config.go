package config

import (
	"log"
	"os"
)

type BrickManagerConfig struct {
	Hostname string
}

func GetBrickManagerConfig() BrickManagerConfig {
	config := BrickManagerConfig{
		getHostname(),
	}
	return config
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}
