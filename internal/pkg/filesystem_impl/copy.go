package filesystem_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"log"
	"strings"
)

func processDataCopy(session datamodel.Session, request datamodel.DataCopyRequest) error {
	cmd, err := generateDataCopyCmd(session, request)
	if err != nil {
		return err
	}
	if cmd == "" {
		log.Println("No files to copy for:", session.Name)
		return nil
	}

	log.Printf("Doing copy: %s", cmd)
	return runner.Execute("localhost", cmd)
}

func generateDataCopyCmd(session datamodel.Session, request datamodel.DataCopyRequest) (string, error) {
	rsync, err := generateRsyncCmd(session, request)
	if err != nil || rsync == "" {
		return "", err
	}

	if len(session.Paths) < 1 {
		log.Panicf("trying to do data copy in for session with no paths %+v", session)
	}
	var exports []string
	for name, value := range session.Paths {
		exports = append(exports, fmt.Sprintf("export %s='%s'", name, value))
	}
	exportString := strings.Join(exports, " && ")

	cmd := fmt.Sprintf("sudo -g '#%d' -u '#%d' %s", session.Group, session.Owner, rsync)
	cmd = fmt.Sprintf("bash -c \"%s && %s\"", exportString, cmd)
	return cmd, nil
}

func generateRsyncCmd(session datamodel.Session, request datamodel.DataCopyRequest) (string, error) {
	if request.Source == "" && request.Destination == "" {
		return "", nil
	}

	var flags string
	if request.SourceType == datamodel.Directory {
		flags = "-r -ospgu --stats"
	} else if request.SourceType == datamodel.File {
		flags = "-ospgu --stats"
	} else {
		return "", fmt.Errorf("unsupported source type %s for volume: %s", request.SourceType, session.Name)
	}

	return fmt.Sprintf("rsync %s %s %s", flags,
		escapePath(request.Source),
		escapePath(request.Destination)), nil
}

func escapePath(path string) string {
	return strings.Replace(path, "$DW_JOB_STRIPED", "\\$DW_JOB_STRIPED", 1)
}
