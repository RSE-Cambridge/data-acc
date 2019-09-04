package config

import (
	"log"
	"os"
	"strconv"
)

type ReadEnvironemnt interface {
	LookupEnv(key string) (string, bool)
	Hostname() (string, error)
}

func getHostname(env ReadEnvironemnt) string {
	hostname, err := env.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}

func getUint(env ReadEnvironemnt, key string, defaultVal uint) uint {
	val, ok := env.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	intVal, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		log.Printf("error parsing %s", key)
		return defaultVal
	}
	return uint(intVal)
}

func getString(env ReadEnvironemnt, key string, defaultVal string) string {
	val, ok := env.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return val
}

func getBool(env ReadEnvironemnt, key string, defaultVal bool) bool {
	val, ok := env.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("error parsing %s", key)
		return defaultVal
	}
	return boolVal
}

type systemEnv struct{}

func (systemEnv) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
	//return "", false
}

func (systemEnv) Hostname() (string, error) {
	return os.Hostname()
	//return "hostname", nil
}

var DefaultEnv ReadEnvironemnt = systemEnv{}
