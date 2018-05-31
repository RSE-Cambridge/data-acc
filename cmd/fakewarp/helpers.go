package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
)

func printJson(message interface{}) {
	b, error := json.MarshalIndent(message, "", "  ")
	if error != nil {
		log.Fatal(error)
	}
	buff := new(bytes.Buffer)
	buff.Write(b)
	buff.Write([]byte("\n"))
	if _, error = buff.WriteTo(os.Stdout); error != nil {
		log.Fatal(error)
	}
}
