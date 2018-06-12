package fakewarp

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Parse a given job file
func parseJobFile(lineSrc GetLines, filename string) error {
	lines, err := lineSrc.Lines(filename)
	if err != nil {
		return err
	}
	_, err = parseJobRequest(lines) // TODO use return value!
	return err
}

type jobCommand interface{}

type cmdCreatePersistent struct {
	Name          string
	CapacityBytes int
	AccessMode    AccessMode
	BufferType    BufferType
	GenericCmd    bool
}

type AccessMode int

const (
	striped           AccessMode = 1
	private                      = 2
	privateAndStriped            = 3
)

type BufferType int

const (
	scratch BufferType = iota
	cache
)

type cmdDestroyPersistent struct{
	Name string
}

type cmdAttachPersistent struct{
	Name string
}

type cmdPerJobBuffer struct{}

type cmdAttachPerJobSwap struct {
	SizeBytes int
}

type cmdStageInData struct{}

type cmdStageOutData struct{}

func parseSize(raw string) (int, error) {
	switch { // TODO move to a dict that contains conversions?
	case strings.HasSuffix(raw, "GiB"):
		rawGb := strings.TrimSuffix(raw, "GiB")
		intGb, err := strconv.Atoi(rawGb)
		if err != nil {
			return 0, err
		}
		return intGb * GbInBytes, nil
	case strings.HasSuffix(raw, "GB"):
		rawGb := strings.TrimSuffix(raw, "GB")
		intGb, err := strconv.Atoi(rawGb)
		if err != nil {
			return 0, err
		}
		return intGb * GbInBytes, nil // TODO... one of these is wrong!
	case strings.HasSuffix(raw, "TiB"):
		rawTb := strings.TrimSuffix(raw, "TiB")
		intTb, err := strconv.Atoi(rawTb)
		if err != nil {
			return 0, err
		}
		return intTb * GbInBytes * 1024, nil
	default:
		return 0, fmt.Errorf("unable to parse size: %s", raw)
	}
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
			log.Println("Skip badly formatted line:", line)
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
				Name: argKeyPair["name"],
				CapacityBytes: size,
				GenericCmd: isGeneric, // TODO... other fields
			}
		case "destroy_persistent":
			command = cmdDestroyPersistent{Name: argKeyPair["name"]}
		case "persistentdw":
			command = cmdAttachPersistent{Name: argKeyPair["name"]}
		case "jobdw":
			command = cmdPerJobBuffer{}
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
			command = cmdStageInData{}
		case "stage_out":
			command = cmdStageOutData{}
		default:
			log.Println("unrecognised command:", cmd, "with argument length", len(args))
			continue
		}
		commands = append(commands, command)
	}
	return commands, nil
}

// Allow tests to replace the source of lines in given file
type GetLines interface {
	Lines(filename string) ([]string, error)
}

// Read lines from a file
// Implements the GetLines interface
type LinesFromFile struct{}

func (LinesFromFile) Lines(filename string) ([]string, error) {
	var lines []string

	file, err := os.Open(filename)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}
	return lines, nil
}
