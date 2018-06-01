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
	app.Usage = "This CLI is used to integrate data-acc with Slurm's Burst Buffer plugin."
	app.Version = version.VERSION

	app.Commands = []cli.Command{
		{
			Name:   "pools",
			Usage:  "List all the buffer pools",
			Action: listPools,
		},
		{
			Name:   "show_instances",
			Usage:  "List the buffer instances.",
			Action: showInstances,
		},
		{
			Name:   "show_sessions",
			Usage:  "List the buffer sessions.",
			Action: showSessions,
		},
		{
			Name:  "teardown",
			Usage: "Destroy the given buffer.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "token, t",
					Usage: "Job ID or Persistent Buffer name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "Path to burst buffer request file.",
				},
				cli.BoolFlag{
					Name: "hurry",
				},
			},
		},
		{
			Name:  "job_process",
			Usage: "Initial call to validate buffer script",
		},
		{
			Name:  "setup",
			Usage: "Create transient burst buffer, called after waiting for enough free capacity.",
		},
		{
			Name:  "real_size",
			Usage: "Report actual size of created buffer.",
		},
		{
			Name:  "data_in",
			Usage: "Copy data into given buffer.",
		},
		{
			Name:  "paths",
			Usage: "Environment variables describing where the buffer will be mounted.",
		},
		{
			Name:  "pre_run",
			Usage: "Attach given buffers to compute nodes specified.",
		},
		{
			Name:  "post_run",
			Usage: "Detach buffers before releasing compute nodes.",
		},
		{
			Name:  "data_out",
			Usage: "Copy data out of buffer.",
		},
		{
			Name:  "create_persistent",
			Usage: "Create a persistent buffer.",
		},
		{
			Name:  "show_configuration",
			Usage: "Returns fake data to keep burst buffer plugin happy.",
		},
	}

	if err := app.Run(stripFunctionArg(os.Args)); err != nil {
		log.Fatal(err)
	}
}
