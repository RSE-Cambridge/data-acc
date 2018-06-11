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
	return parseJobRequest(lines)
}

func parseJobRequest(lines []string) error {
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		if len(tokens) < 3 {
			log.Println("Skip badly formatted line", line)
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
		log.Println("Is generic command:", isGeneric)

		switch cmd {
		case "asdf":
			log.Println("test!")
		default:
			log.Println("unrecognised command:", cmd, "with argument length", len(args))
		}
	}
	return nil
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
