package fileio

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
)

// Allow tests to replace the source of lines in given file
type Disk interface {
	Lines(filename string) ([]string, error)
	Write(filename string, lines []string) error
}

func NewDisk() Disk {
	return &linesFromFile{}
}

// Read lines from a file
// Implements the GetLines interface
type linesFromFile struct{}

func (linesFromFile) Write(filename string, lines []string) error {
	content := strings.Join(lines, "\n")
	data := bytes.NewBufferString(content).Bytes()
	return ioutil.WriteFile(filename, data, 0644)
}

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
