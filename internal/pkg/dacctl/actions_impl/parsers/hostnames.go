package parsers

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"regexp"
)

// TODO: allowed "-" so uuid passes validation
var nameRegex = regexp.MustCompile("^[a-zA-Z0-9.-]+$")

func IsValidName(name string) bool {
	return nameRegex.Match([]byte(name))
}

var keyRegex = regexp.MustCompile("^[a-zA-Z0-9._-]+$")

func IsValidKey(name string) bool {
	return keyRegex.Match([]byte(name))
}

var pathRegex = regexp.MustCompile("^[a-zA-Z0-9.,_/$-]+$")

func IsValidPath(value string) bool {
	return pathRegex.Match([]byte(value))
}

func GetHostnamesFromFile(disk fileio.Disk, filename string) ([]string, error) {
	computeHosts, err := disk.Lines(filename)
	if err != nil {
		return nil, err
	}
	var invalidHosts []string
	for _, computeHost := range computeHosts {
		if !IsValidName(computeHost) {
			invalidHosts = append(invalidHosts, computeHost)
		}
	}
	if len(invalidHosts) > 0 {
		return nil, fmt.Errorf("invalid hostname in: %s", invalidHosts)
	}
	return computeHosts, nil
}
