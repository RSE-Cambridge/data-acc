package jobfile

import (
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/parsers/capacity"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"log"
	"strings"
)

type jobSummary struct {
	PerJobBuffer *cmdPerJobBuffer
	Swap         *cmdAttachPerJobSwap
	Attachments  []model.VolumeName
	DataIn       []model.DataCopyRequest
	DataOut      []model.DataCopyRequest
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
			summary.Attachments = append(summary.Attachments, model.VolumeName(c))
		case cmdAttachPerJobSwap:
			if summary.Swap != nil {
				return summary, fmt.Errorf("only one swap request allowed")
			}
			summary.Swap = &c
		case cmdStageOutData:
			summary.DataOut = append(summary.DataOut, model.DataCopyRequest{
				SourceType:  c.SourceType,
				Source:      c.Source,
				Destination: c.Destination,
			})
		case cmdStageInData:
			summary.DataIn = append(summary.DataIn, model.DataCopyRequest{
				SourceType:  c.SourceType,
				Source:      c.Source,
				Destination: c.Destination,
			})
		default:
			// do nothing
		}
	}
	return summary, nil
}

type jobCommand interface{}

var stringToAccessMode = map[string]model.AccessMode{
	"":                model.Striped,
	"striped":         model.Striped,
	"private":         model.Private,
	"private,striped": model.PrivateAndStriped,
	"striped,private": model.PrivateAndStriped,
}

func AccessModeFromString(raw string) model.AccessMode {
	return stringToAccessMode[strings.ToLower(raw)]
}

var stringToBufferType = map[string]model.BufferType{
	"":        model.Scratch,
	"scratch": model.Scratch,
	"cache":   model.Cache,
}

type cmdCreatePersistent struct {
	Name          string
	CapacityBytes int
	AccessMode    model.AccessMode
	BufferType    model.BufferType
	GenericCmd    bool
}

func BufferTypeFromString(raw string) model.BufferType {
	return stringToBufferType[strings.ToLower(raw)]
}

type cmdDestroyPersistent string

type cmdAttachPersistent string

type cmdPerJobBuffer struct {
	CapacityBytes int
	AccessMode    model.AccessMode
	BufferType    model.BufferType
	GenericCmd    bool
}

type cmdAttachPerJobSwap struct {
	SizeBytes int
}

var stringToStageType = map[string]model.SourceType{
	"directory": model.Directory,
	"file":      model.File,
	"list":      model.List,
}

func sourceTypeFromString(raw string) model.SourceType {
	return stringToStageType[strings.ToLower(raw)]
}

type cmdStageInData model.DataCopyRequest

type cmdStageOutData model.DataCopyRequest

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
			size, err := capacity.ParseSize(argKeyPair["capacity"])
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
			command = cmdDestroyPersistent(argKeyPair["name"])
		case "persistentdw":
			command = cmdAttachPersistent(argKeyPair["name"])
		case "jobdw":
			size, err := capacity.ParseSize(argKeyPair["capacity"])
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
			if size, err := capacity.ParseSize(args[0]); err != nil {
				log.Println(err)
				continue
			} else {
				command = cmdAttachPerJobSwap{SizeBytes: size}
			}
		case "stage_in":
			command = cmdStageInData{
				Source:      argKeyPair["source"],
				Destination: argKeyPair["destination"],
				SourceType:  sourceTypeFromString(argKeyPair["type"]),
			}
		case "stage_out":
			command = cmdStageOutData{
				Source:      argKeyPair["source"],
				Destination: argKeyPair["destination"],
				SourceType:  sourceTypeFromString(argKeyPair["type"]),
			}
		default:
			log.Println("unrecognised command:", cmd, "with argument length", len(args))
			continue
		}
		commands = append(commands, command)
	}
	return commands, nil
}
