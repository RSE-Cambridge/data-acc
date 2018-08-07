package fake

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func processDataCopy(volumeName registry.VolumeName, request registry.DataCopyRequest) error {
	cmd, err := generateDataCopyCmd(volumeName, request)
	log.Printf("FAKE %s", cmd)
	return err
}

func generateDataCopyCmd(volumeName registry.VolumeName, request registry.DataCopyRequest) (string, error) {
	if request.Source == "" && request.Destination == "" {
		log.Println("No files to copy for:", volumeName)
		return "", nil
	}

	var flags string
	if request.SourceType == registry.Directory {
		flags = "-r "
	} else if request.SourceType == registry.File {
		flags = ""
	} else {
		log.Println("Unspported source type", request.SourceType, "for volume:", volumeName)
		return "", nil
	}

	return fmt.Sprintf("rsync %s%s %s", flags, request.Source, request.Destination), nil
}
