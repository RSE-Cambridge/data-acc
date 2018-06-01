package fakewarp

import (
	"errors"
	"fmt"
)

type PersistentBufferRequest struct {
	Token    string
	Caller   string
	Capacity string
	User     string
	Group    string
	Access   string
	Type     string
}

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func CreatePersistentBuffer(c CliContext) (string, error) {
	fmt.Printf("--token %s --caller %s --user %d --groupid %d --capacity %s "+
		"--access %s --type %s\n",
		c.String("token"), c.String("caller"), c.Int("user"),
		c.Int("groupid"), c.String("capacity"), c.String("access"), c.String("type"))
	if c.String("token") == "bob" {
		return "", errors.New("unable to create buffer")
	}
	return c.String("name"), nil
}
