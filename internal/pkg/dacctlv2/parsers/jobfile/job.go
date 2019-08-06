package jobfile

import (
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"log"
	"strconv"
	"strings"
)

type jobSummary struct {
	PerJobBuffer *cmdPerJobBuffer
	Swap         *cmdAttachPerJobSwap
	Attachments  []cmdAttachPersistent
	DataIn       *cmdStageInData
	DataOut      *cmdStageOutData
	//createPersistent  *cmdCreatePersistent
	//destroyPersistent *cmdDestroyPersistent
}

func (s jobSummary) String() string {
	return toJson(s)
}

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	return string(b)
}

// Parse a given job file
func ParseJobFile(disk fileio.Disk, filename string) (jobSummary, error) {
	lines, err := disk.Lines(filename)
	if err != nil {
		return jobSummary{}, err
	}
	return getJobSummary(lines)
}

func getJobSummary(lines []string) (jobSummary, error) {
	var summary jobSummary
	jobCommands, err := parseJobRequest(lines)
	if err != nil {
		return summary, err
	}

	for _, cmd := range jobCommands {
		switch c := cmd.(type) {
		case cmdPerJobBuffer:
			if summary.PerJobBuffer == nil {
				summary.PerJobBuffer = &c
			} else {
				return summary, fmt.Errorf("only one per job buffer allowed")
			}
		case cmdAttachPersistent:
			summary.Attachments = append(summary.Attachments, c)
		case cmdAttachPerJobSwap:
			if summary.Swap != nil {
				// TODO check amount isn't too big for per job buffer
				return summary, fmt.Errorf("only one swap request allowed")
			}
			summary.Swap = &c
		case cmdStageOutData:
			if summary.DataOut != nil {
				// TODO really should check if data out matches one of the requested buffers
				return summary, fmt.Errorf("only one per data out requested allowed")
			}
			summary.DataOut = &c
		case cmdStageInData:
			if summary.DataIn != nil {
				// TODO really should check if data in matches one of the requested buffers
				return summary, fmt.Errorf("only one per data in requested allowed")
			}
			summary.DataIn = &c
		default:
			// do nothing
		}
	}
	return summary, nil
}

type jobCommand interface{}

type AccessMode int

const (
	striped           AccessMode = 0
	private                      = 1
	privateAndStriped            = 2
)

var stringToAccessMode = map[string]AccessMode{
	"":                striped,
	"striped":         striped,
	"private":         private,
	"private,striped": privateAndStriped,
	"striped,private": privateAndStriped,
}

func AccessModeFromString(raw string) AccessMode {
	return stringToAccessMode[strings.ToLower(raw)]
}

type BufferType int

const (
	scratch BufferType = iota
	cache
)

var stringToBufferType = map[string]BufferType{
	"":        scratch,
	"scratch": scratch,
	"cache":   cache,
}

type cmdCreatePersistent struct {
	Name          string
	CapacityBytes int
	AccessMode    AccessMode
	BufferType    BufferType
	GenericCmd    bool
}

func BufferTypeFromString(raw string) BufferType {
	return stringToBufferType[strings.ToLower(raw)]
}

type cmdDestroyPersistent struct {
	Name string
}

type cmdAttachPersistent struct {
	Name string
}

type cmdPerJobBuffer struct {
	CapacityBytes int
	AccessMode    AccessMode
	BufferType    BufferType
	GenericCmd    bool
}

type cmdAttachPerJobSwap struct {
	SizeBytes int
}

type StageType int

const (
	directory StageType = iota
	file                // TODO there is also list, but we ignore that for now
)

var stringToStageType = map[string]StageType{
	"":          directory,
	"directory": directory,
	"file":      file,
}

func stageTypeFromString(raw string) StageType {
	return stringToStageType[strings.ToLower(raw)]
}

type cmdStageInData struct {
	Source      string
	Destination string
	StageType   StageType
}

type cmdStageOutData struct {
	Source      string
	Destination string
	StageType   StageType
}

var sizeSuffixMulitiplyer = map[string]int{
	"TiB": 1099511627776,
	"TB":  1000000000000,
	"GiB": 1073741824,
	"GB":  1000000000,
	"MiB": 1048576,
	"MB":  1000000,
}

func parseSize(raw string) (int, error) {
	intVal, err := strconv.Atoi(raw)
	if err == nil {
		// specified raw bytes
		return intVal, nil
	}
	for suffix, multiplyer := range sizeSuffixMulitiplyer {
		if strings.HasSuffix(raw, suffix) {
			rawInt := strings.TrimSuffix(raw, suffix)
			intVal, err := strconv.Atoi(rawInt)
			if err != nil {
				return 0, err
			}
			return intVal * multiplyer, nil
		}
	}
	return 0, fmt.Errorf("unable to parse size: %s", raw)
}

func parseArgs(rawArgs []string) (map[string]string, error) {
	args := make(map[string]string, len(rawArgs))
	for _, arg := range rawArgs {
		parts := strings.Split(arg, "=")
		if len(parts) != 2 {
			return args, fmt.Errorf("unable to parse arg: %s", arg)
		}
		args[strings.ToLower(parts[0])] = parts[1]
	}
	return args, nil
}

func parseJobRequest(lines []string) ([]jobCommand, error) {
	var commands []jobCommand
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		if len(tokens) < 3 {
			if line != "" && line != "#!/bin/bash" {
				log.Println("Skip badly formatted line:", line)
			}
			continue
		}

		cmdType := tokens[0]
		cmd := tokens[1]
		args := tokens[2:]

		var isGeneric bool
		switch cmdType {
		case "#DW":
			isGeneric = false
		case "#BB":
			isGeneric = true
		default:
			log.Println("unrecognised command type:", cmdType)
			continue
		}

		argKeyPair, _ := parseArgs(args) // TODO deal with errors when not swap

		var command jobCommand
		switch cmd {
		case "create_persistent":
			size, err := parseSize(argKeyPair["capacity"])
			if err != nil {
				log.Println(err)
				continue
			}
			command = cmdCreatePersistent{
				Name:          argKeyPair["name"],
				CapacityBytes: size,
				GenericCmd:    isGeneric,
				AccessMode:    AccessModeFromString(argKeyPair["access_mode"]),
				BufferType:    BufferTypeFromString(argKeyPair["type"]),
			}
		case "destroy_persistent":
			command = cmdDestroyPersistent{Name: argKeyPair["name"]}
		case "persistentdw":
			command = cmdAttachPersistent{Name: argKeyPair["name"]}
		case "jobdw":
			size, err := parseSize(argKeyPair["capacity"])
			if err != nil {
				log.Println(err)
				continue
			}
			command = cmdPerJobBuffer{
				CapacityBytes: size,
				GenericCmd:    isGeneric,
				AccessMode:    AccessModeFromString(argKeyPair["access_mode"]),
				BufferType:    BufferTypeFromString(argKeyPair["type"]),
			}
		case "swap":
			if len(args) != 1 {
				log.Println("Unable to parse swap command:", line)
			}
			if size, err := parseSize(args[0]); err != nil {
				log.Println(err)
				continue
			} else {
				command = cmdAttachPerJobSwap{SizeBytes: size}
			}
		case "stage_in":
			command = cmdStageInData{
				Source:      argKeyPair["source"],
				Destination: argKeyPair["destination"],
				StageType:   stageTypeFromString(argKeyPair["type"]),
			}
		case "stage_out":
			command = cmdStageOutData{
				Source:      argKeyPair["source"],
				Destination: argKeyPair["destination"],
				StageType:   stageTypeFromString(argKeyPair["type"]),
			}
		default:
			log.Println("unrecognised command:", cmd, "with argument length", len(args))
			continue
		}
		commands = append(commands, command)
	}
	return commands, nil
}
