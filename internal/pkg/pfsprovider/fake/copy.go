package fake

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func processDataCopy(volumeName registry.VolumeName, request registry.DataCopyRequest) error {
	if request.Source == "" && request.Destination == "" {
		log.Println("No files to copy for:", volumeName)
		return nil
	}

	var flags string
	if request.SourceType != registry.Directory {
		flags = "-r"
	} else if request.SourceType == registry.File {
		flags = ""
	} else {
		log.Println("Unspported source type", request.SourceType, "for volume:", volumeName)
		return nil
	}

	log.Printf("FAKE rsync %s %s %s", flags, request.Source, request.Destination)
	return nil
}
