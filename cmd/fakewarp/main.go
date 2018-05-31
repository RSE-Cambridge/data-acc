package main

import (
	"log"
	"os"
)

func stripFunctionArg(systemArgs []string) []string {
	if len(systemArgs) > 2 && systemArgs[1] == "--function" {
		newArgs := systemArgs[0:1]
		for _, arg := range systemArgs[2:] {
			newArgs = append(newArgs, arg)
		}
		return newArgs
	}
	return systemArgs
}

func main() {
	log.Printf("Hello! Args: %s", stripFunctionArg(os.Args)[1:])
}
