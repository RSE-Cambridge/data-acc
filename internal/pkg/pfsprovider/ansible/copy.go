package ansible

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func processDataCopy(volume registry.Volume, request registry.DataCopyRequest) error {
	cmd, err := generateDataCopyCmd(volume, request)
	if err != nil {
		return err
	}

	log.Printf("FAKE %s", cmd)
	return nil
}

func generateDataCopyCmd(volume registry.Volume, request registry.DataCopyRequest) (string, error) {
	rsync, err := generateRsyncCmd(volume, request)
	if err != nil || rsync == "" {
		return "", err
	}

	cmd := fmt.Sprintf("sudo su `getent passwd %d | cut -d: -f1` %s", volume.Owner, rsync)
	return cmd, nil
}

func generateRsyncCmd(volume registry.Volume, request registry.DataCopyRequest) (string, error) {
	if request.Source == "" && request.Destination == "" {
		log.Println("No files to copy for:", volume.Name)
		return "", nil
	}

	var flags string
	if request.SourceType == registry.Directory {
		flags = "-r "
	} else if request.SourceType == registry.File {
		flags = ""
	} else {
		return "", fmt.Errorf("unsupported source type %s for volume: %s", request.SourceType, volume.Name)
	}

	return fmt.Sprintf("rsync %s%s %s", flags, request.Source, request.Destination), nil
}
