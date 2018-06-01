package fakewarp

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
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
	error := processCreatePersistentBuffer(&request, etcdregistry.NewKeystore())
	return request.Token, error
}

func processCreatePersistentBuffer(request *PersistentBufferRequest, keystore keystoreregistry.Keystore) error {
	if request.Token == "bob" {
		return errors.New("unable to create buffer")
	}
	r := keystoreregistry.BufferRegistry{keystore}
	// TODO: lots more validation needed to ensure valid key, etc
	buf := registry.Buffer{Name: request.Token, Owner: fmt.Sprintf("%d", request.User)}
	r.AddBuffer(buf)
	return nil
}
