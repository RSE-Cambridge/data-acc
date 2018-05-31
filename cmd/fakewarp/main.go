package main

import (
	"github.com/RSE-Cambridge/data-acc/pkg/version"
	"github.com/urfave/cli"
	"log"
	"os"
)

func stripFunctionArg(systemArgs []string) []string {
	if len(systemArgs) > 2 && systemArgs[1] == "--function" {
		return append(systemArgs[0:1], systemArgs[2:]...)
	}
	return systemArgs
}

func main() {
	app := cli.NewApp()
	app.Name = "FakeWarp CLI"
	app.Usage = "This is to integrate data-acc with Slurm's Burst Buffer plugin."
	app.Version = version.VERSION

	app.Commands = []cli.Command{
		{
			Name:    "pools",
			Aliases: []string{"p"},
			Usage:   "List all the buffer pools",
			Action:  pools,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "lang, l",
					Value: "english",
					Usage: "Language for the greeting",
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Load configuration from `FILE`",
				},
			},
		},
	}

	if err := app.Run(stripFunctionArg(os.Args)); err != nil {
		log.Fatal(err)
	}
}
