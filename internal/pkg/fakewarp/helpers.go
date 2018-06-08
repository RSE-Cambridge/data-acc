package fakewarp

import (
	"encoding/json"
	"log"
)

type CliContext interface {
	String(name string) string
	Int(name string) int
}

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	return string(b)
}
