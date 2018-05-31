package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
)

func printJson(w io.Writer, message interface{}) {
	b, error := json.Marshal(message)
	if error != nil {
		log.Fatal(error)
	}
	buff := new(bytes.Buffer)
	buff.Write(b)
	buff.Write([]byte("\n"))
	if _, error = buff.WriteTo(w); error != nil {
		log.Fatal(error)
	}
}
