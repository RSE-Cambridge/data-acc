package main

import (
	"github.com/urfave/cli"
	"io"
	"os"
)

type pool struct {
	Id          string `json:"id"`
	Units       string `json:"units"`
	Granularity uint   `json:"granularity"`
	Quantity    uint   `json:"quantity"`
	Free        uint   `json:"free"`
}

func getPools() []pool {
	fakePool := pool{"fake", "bytes", 214748364800, 40, 3}
	return []pool{fakePool}
}

func printPools(writer io.Writer) {
	message := map[string][]pool{"pools": getPools()}
	printJson(writer, message)
}

func pools(_ *cli.Context) error {
	printPools(os.Stdout)
	return nil
}
