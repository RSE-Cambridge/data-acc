package main

import (
	"github.com/urfave/cli"
	"io"
	"os"
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

func getInstances() []instance {
	fakeInstance := instance{
		"fakebuffer",
		instanceCapacity{3, 40},
		instanceLinks{"fakebuffer"}}
	return []instance{fakeInstance}
}

func printInstances(writer io.Writer) {
	message := map[string][]instance{"instances": getInstances()}
	printJson(writer, message)
}

func showInstances(_ *cli.Context) error {
	printInstances(os.Stdout)
	return nil
}
