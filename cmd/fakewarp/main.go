package main

import (
	"log"
	"os"
)

func stripFunctionArg() []string {
	system := os.Args
	if len(system)>2 && system[1]=="--function" {
		newArgs := system[0:1]
		for _, arg := range system[2:] {
			newArgs = append(newArgs, arg)
		}
		return newArgs
	}
	return system[2:]
}

func main() {
	log.Printf("Hello! %s", stripFunctionArg())
}