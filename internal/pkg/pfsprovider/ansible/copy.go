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
	if cmd == "" {
		log.Println("No files to copy for:", volume.Name)
		return nil
	}

	log.Printf("FAKE copy: %s", cmd)
	return nil
}

func generateDataCopyCmd(volume registry.Volume, request registry.DataCopyRequest) (string, error) {
	rsync, err := generateRsyncCmd(volume, request)
	if err != nil || rsync == "" {
		return "", err
	}

	cmd := fmt.Sprintf("sudo -g '#%d' -u '#%d' %s", volume.Group, volume.Owner, rsync)
	cmd = fmt.Sprintf("bash -c \"export JOB='%s' && %s\"", volume.JobName, cmd)
	return cmd, nil
}

func generateRsyncCmd(volume registry.Volume, request registry.DataCopyRequest) (string, error) {
	if request.Source == "" && request.Destination == "" {
		return "", nil
	}

	var flags string
	if request.SourceType == registry.Directory {
		flags = "-r -ospgu --stats"
	} else if request.SourceType == registry.File {
		flags = "-ospgu --stats"
	} else {
		return "", fmt.Errorf("unsupported source type %s for volume: %s", request.SourceType, volume.Name)
	}

	return fmt.Sprintf("rsync %s %s %s", flags, request.Source, request.Destination), nil
}
