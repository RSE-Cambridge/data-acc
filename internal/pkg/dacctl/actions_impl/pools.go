package actions_impl

import (
	"encoding/json"
	"log"
)

type pool struct {
	Id          string `json:"id"`
	Units       string `json:"units"`
	Granularity uint   `json:"granularity"`
	Quantity    uint   `json:"quantity"`
	Free        uint   `json:"free"`
}

type pools []pool

func getPoolsAsString(list pools) string {
	message := map[string]pools{"pools": list}
	output, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(output)
}
