package fakewarp

import (
	"encoding/json"
	"log"
)

func toJson(message interface{}) string {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	return string(b)
}
