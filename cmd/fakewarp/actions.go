package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/urfave/cli"
	"log"
	"strings"
)

func showInstances(_ *cli.Context) error {
	fmt.Print(fakewarp.GetInstances())
	return nil
}

func showSessions(_ *cli.Context) error {
	fmt.Print(fakewarp.GetSessions())
	return nil
}

func listPools(_ *cli.Context) error {
	fmt.Print(fakewarp.GetPools())
	return nil
}

func showConfigurations(_ *cli.Context) error {
	fmt.Print(fakewarp.GetConfigurations())
	return nil
}

func checkRequiredStrings(c *cli.Context, flags ...string) {
	errors := []string{}
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errors = append(errors, flag)
		}
	}
	if len(errors) > 0 {
		log.Fatalf("Please provide these required parameters: %s", strings.Join(errors, ", "))
	}
}

func teardown(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("token: %s job: %s hurry:%t\n",
		c.String("token"), c.String("job"), c.Bool("hurry"))
	return nil
}

func jobProcess(c *cli.Context) error {
	checkRequiredStrings(c, "job")
	fmt.Printf("job: %s\n", c.String("job"))
	return nil
}

func setup(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "caller", "capacity")
	fmt.Printf("--token %s --job %s --caller %s --user %d --groupid %d --capacity %s\n",
		c.String("token"), c.String("job"), c.String("caller"), c.Int("user"),
		c.Int("groupid"), c.String("capacity"))
	return nil
}

func realSize(c *cli.Context) error {
	checkRequiredStrings(c, "token")
	fmt.Printf("--token %s\n", c.String("token"))
	return nil
}

func dataIn(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n", c.String("token"), c.String("job"))
	return nil
}

func paths(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "pathfile")
	fmt.Printf("--token %s --job %s --pathfile %s\n",
		c.String("token"), c.String("job"), c.String("pathfile"))
	return nil
}

func preRun(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "nodehostnamefile")
	fmt.Printf("--token %s --job %s --nodehostnamefile %s\n",
		c.String("token"), c.String("job"), c.String("nodehostnamefile"))
	return nil
}

func postRun(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))
	return nil
}

func dataOut(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))
	return nil
}

func createPersistent(c *cli.Context) error {
	checkRequiredStrings(c, "token", "caller", "capacity", "user", "access", "type")
	fmt.Printf("--token %s --caller %s --user %d --groupid %d --capacity %s " +
		"--access %s --type %s\n",
		c.String("token"), c.String("caller"), c.Int("user"),
		c.Int("groupid"), c.String("capacity"), c.String("access"), c.String("type"))
	return nil
}