package fakewarp

import (
	"bufio"
	"log"
	"os"
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
	CapacityBytes int64
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

type cmdDestroyPersistent struct{}

type cmdAttachPersistent struct{}

type cmdPerJobBuffer struct{}

type cmdAttachPerJobSwap struct{}

type cmdStageInData struct{}

type cmdStageOutData struct{}

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

		var command jobCommand
		switch cmd {
		case "create_persistent":
			command = cmdCreatePersistent{GenericCmd: isGeneric}
		case "destroy_persistent":
			command = cmdDestroyPersistent{}
		case "persistentdw":
			command = cmdAttachPersistent{}
		case "jobdw":
			command = cmdPerJobBuffer{}
		case "swap":
			command = cmdAttachPerJobSwap{}
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