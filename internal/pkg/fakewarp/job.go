package fakewarp

import (
	"log"
	"strings"
)

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

// returns an error if the format of the file is bad
// or requests something unsupported
func parseJobFile(filename string) error {
	/*file, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return parseJobRequest(lines)*/
	return nil
}
