package fakewarp

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"math/rand"
	"strings"
	"time"
)

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func DeleteBuffer(c CliContext, volReg registry.VolumeRegistry) error {
	token := c.String("token")
	if err := volReg.DeleteVolume(registry.VolumeName(token)); err != nil {
		return err
	}
	return volReg.DeleteJob(token)
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

func CreatePerJobBuffer(c CliContext, volReg registry.VolumeRegistry) error {
	// TODO need to read and parse the job file...
	if err := parseJobFile(c.String("job")); err != nil {
		return err
	}
	return createVolumesAndJobs(volReg, BufferRequest{
		Token:    c.String("token"),
		User:     c.Int("user"),
		Group:    c.Int("group"),
		Capacity: c.String("capacity"),
		Caller:   c.String("caller"),
	})
}

// TODO: need to reuse this with the new logic
func getBricksForBuffer(volRegistry registry.VolumeRegistry, buffer *registry.Volume) []registry.BrickInfo {
	log.Println("Add fakebuffer and match to bricks")

	availableBricks := make(map[string][]registry.BrickInfo) // hostname to available bricks, getAvailableBricks(cli)
	var chosenBricks []registry.BrickInfo

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	// TODO: should look at buffer to get number of bricks required
	requestedBricks := 2

	var hosts []string
	for key := range availableBricks {
		hosts = append(hosts, key)
	}

	randomWalk := rand.Perm(len(availableBricks))
	for _, i := range randomWalk {
		hostBricks := availableBricks[hosts[i]]
		candidateBrick := hostBricks[r.Intn(len(hostBricks))]

		goodCandidate := true
		for _, brick := range chosenBricks {
			if brick == candidateBrick {
				goodCandidate = false
				break
			}
			if brick.Hostname == candidateBrick.Hostname {
				goodCandidate = false
				break
			}
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}

		if len(chosenBricks) >= requestedBricks {
			break
		}
	}
	// TODO: check we have enough bricks?

	// TODO: should be done in a single transaction, and retry if clash
	for i, brick := range chosenBricks {
		chosenKey := fmt.Sprintf("/bricks/inuse/%s/%s", brick.Hostname, brick.Device)
		log.Println("Add to etcd:", chosenKey, fmt.Sprintf("%s:%d", buffer.Name, i))
		//TODO: keystore.AtomicAdd(chosenKey, fmt.Sprintf("%s:%d", buffer.Name, i))
	}

	return chosenBricks
}
