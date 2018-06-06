package keystoreregistry

import (
	"bytes"
	"encoding/json"
	"log"
)

type Keystore interface {
	// Used to clean up any resources being used
	// ... such as a connection to etcd
	Close() error

	// Removes any key starting with the given prefix
	// An error is returned if nothing was deleted,
	// which some users may choose to safely ignore
	CleanPrefix(prefix string) error

	AtomicAdd(key string, value string)
	WatchPutPrefix(prefix string, onPut func(string, string))
}

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	buffer := new(bytes.Buffer)
	buffer.Write(b)
	buffer.Write([]byte("\n"))
	return buffer.String()
}
