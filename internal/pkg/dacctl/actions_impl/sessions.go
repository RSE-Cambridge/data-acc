package actions_impl

import (
	"encoding/json"
	"log"
)

type session struct {
	Id      string `json:"id"`
	Created uint   `json:"created"`
	Owner   uint   `json:"owner"`
	Token   string `json:"token"`
}

type sessions []session

func sessonsToString(list []session) string {
	message := map[string]sessions{"sessions": list}
	output, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(output)
}
