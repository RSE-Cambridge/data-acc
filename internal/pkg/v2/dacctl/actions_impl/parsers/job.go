package parsers

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"log"
	"strings"
)

type jobSummary struct {
	PerJobBuffer *cmdPerJobBuffer
	Swap         *cmdAttachPerJobSwap
	Attachments  []datamodel.SessionName
	DataIn       []datamodel.DataCopyRequest
	DataOut      []datamodel.DataCopyRequest
	// TODO: support create and destroy persistent?
	//createPersistent  *cmdCreatePersistent
	//destroyPersistent *cmdDestroyPersistent
}

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
				return jobSummary{}, fmt.Errorf("only one per job buffer allowed")
			}
		case cmdAttachPersistent:
			summary.Attachments = append(summary.Attachments, datamodel.SessionName(c))
		case cmdAttachPerJobSwap:
			if summary.Swap != nil {
				return jobSummary{}, fmt.Errorf("only one swap request allowed")
			}
			summary.Swap = &c
		case cmdStageOutData:
			summary.DataOut = append(summary.DataOut, datamodel.DataCopyRequest{
				SourceType:  c.SourceType,
				Source:      c.Source,
				Destination: c.Destination,
			})
		case cmdStageInData:
			summary.DataIn = append(summary.DataIn, datamodel.DataCopyRequest{
				SourceType:  c.SourceType,
				Source:      c.Source,
				Destination: c.Destination,
			})
		}
	}
	return summary, nil
}

type jobCommand interface{}

var stringToAccessMode = map[string]datamodel.AccessMode{
	"":                datamodel.Striped,
	"striped":         datamodel.Striped,
	"private":         datamodel.Private,
	"private,striped": datamodel.PrivateAndStriped,
	"striped,private": datamodel.PrivateAndStriped,
}

func accessModeFromString(raw string) datamodel.AccessMode {
	return stringToAccessMode[strings.ToLower(raw)]
}

var stringToBufferType = map[string]datamodel.BufferType{
	"":        datamodel.Scratch,
	"scratch": datamodel.Scratch,
	"cache":   datamodel.Cache,
}

type cmdCreatePersistent struct {
	Name          string
	CapacityBytes int
	AccessMode    datamodel.AccessMode
	BufferType    datamodel.BufferType
	GenericCmd    bool
}

func bufferTypeFromString(raw string) datamodel.BufferType {
	return stringToBufferType[strings.ToLower(raw)]
}

type cmdDestroyPersistent string

type cmdAttachPersistent string

type cmdPerJobBuffer struct {
	CapacityBytes int
	AccessMode    datamodel.AccessMode
	BufferType    datamodel.BufferType
	GenericCmd    bool
}

type cmdAttachPerJobSwap struct {
	SizeBytes int
}

var stringToStageType = map[string]datamodel.SourceType{
	"directory": datamodel.Directory,
	"file":      datamodel.File,
	"list":      datamodel.List,
}

func sourceTypeFromString(raw string) datamodel.SourceType {
	return stringToStageType[strings.ToLower(raw)]
}

type cmdStageInData datamodel.DataCopyRequest

type cmdStageOutData datamodel.DataCopyRequest

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
			size, err := ParseSize(argKeyPair["capacity"])
			if err != nil {
				log.Println(err)
				return nil, err
			}
			command = cmdCreatePersistent{
				Name:          argKeyPair["name"],
				CapacityBytes: size,
				GenericCmd:    isGeneric,
				AccessMode:    accessModeFromString(argKeyPair["access_mode"]),
				BufferType:    bufferTypeFromString(argKeyPair["type"]),
			}
		case "destroy_persistent":
			command = cmdDestroyPersistent(argKeyPair["name"])
		case "persistentdw":
			command = cmdAttachPersistent(argKeyPair["name"])
		case "jobdw":
			size, err := ParseSize(argKeyPair["capacity"])
			if err != nil {
				log.Println(err)
				return nil, err
			}
			command = cmdPerJobBuffer{
				CapacityBytes: size,
				GenericCmd:    isGeneric,
				AccessMode:    accessModeFromString(argKeyPair["access_mode"]),
				BufferType:    bufferTypeFromString(argKeyPair["type"]),
			}
		case "swap":
			if len(args) != 1 {
				return nil, fmt.Errorf("unable to parse swap command: %s", line)
			}
			if size, err := ParseSize(args[0]); err != nil {
				log.Println(err)
				return nil, err
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
			return nil, fmt.Errorf("unrecognised command: %s with arguments: %s", cmd, args)
		}
		commands = append(commands, command)
	}
	return commands, nil
}
