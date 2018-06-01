package fakewarp

import (
	"errors"
	"fmt"
)

type PersistentBufferRequest struct {
	Token    string
	Caller   string
	Capacity string
	User     int
	Group    int
	Access   string
	Type     string
}

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func CreatePersistentBuffer(c CliContext) (string, error) {
	request := PersistentBufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), c.String("access"), c.String("type")}
	fmt.Printf("--token %s --caller %s --user %d --groupid %d --capacity %s --access %s --type %s\n",
		request.Token, request.Caller, request.User, request.Group, request.Capacity, request.Access, request.Type)
	error := processCreatePersistentBuffer(&request)
	return request.Token, error
}

func processCreatePersistentBuffer(request *PersistentBufferRequest) error {
	if request.Token == "bob" {
		return errors.New("unable to create buffer")
	}
	return nil
}
