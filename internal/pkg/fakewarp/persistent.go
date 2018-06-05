package fakewarp

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/oldregistry"
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
func CreatePersistentBuffer(c CliContext, keystore keystoreregistry.Keystore) (string, error) {
	request := PersistentBufferRequest{c.String("token"), c.String("caller"),
		c.String("capacity"), c.Int("user"),
		c.Int("groupid"), c.String("access"), c.String("type")}
	fmt.Printf("--token %s --caller %s --user %d --groupid %d --capacity %s --access %s --type %s\n",
		request.Token, request.Caller, request.User, request.Group, request.Capacity, request.Access, request.Type)
	error := processCreatePersistentBuffer(&request, keystore)
	return request.Token, error
}

func processCreatePersistentBuffer(request *PersistentBufferRequest, keystore keystoreregistry.Keystore) error {
	if request.Token == "bob" {
		return errors.New("unable to create buffer")
	}
	r := keystoreregistry.NewBufferRegistry(keystore)
	// TODO: lots more validation needed to ensure valid key, etc
	buf := oldregistry.Buffer{Name: request.Token, Owner: fmt.Sprintf("%d", request.User)}
	r.AddBuffer(buf)
	return nil
}
