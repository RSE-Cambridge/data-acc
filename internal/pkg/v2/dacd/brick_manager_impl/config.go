package brick_manager_impl

import (
	"log"
	"os"
)

type brickManagerConfiguration struct {
	hostname string
}

func getConfig() brickManagerConfiguration {
	config := brickManagerConfiguration{
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