package actions_impl

import (
	"encoding/json"
	"github.com/prometheus/common/log"
)

type instanceCapacity struct {
	Bytes uint `json:"bytes"`
	Nodes uint `json:"nodes"`
}

type instanceLinks struct {
	Session string `json:"session"`
}

type instance struct {
	Id       string           `json:"id"`
	Capacity instanceCapacity `json:"capacity"`
	Links    instanceLinks    `json:"links"`
}

type instances []instance

func instancesToString(list []instance) string {
	message := map[string]instances{"instances": list}
	output, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(output)
}
