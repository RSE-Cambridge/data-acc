package fileio

import (
	"bufio"
	"os"
)

// Allow tests to replace the source of lines in given file
type Reader interface {
	Lines(filename string) ([]string, error)
}

func NewReader() Reader {
	return &linesFromFile{}
}

// Read lines from a file
// Implements the GetLines interface
type linesFromFile struct{}

func (linesFromFile) Lines(filename string) ([]string, error) {
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
