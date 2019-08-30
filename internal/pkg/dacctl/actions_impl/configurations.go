package actions_impl

import (
	"encoding/json"
	"log"
)

type configurations []string

func configurationToString(list configurations) string {
	message := map[string]configurations{"configurations": list}
	output, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(output)
}
