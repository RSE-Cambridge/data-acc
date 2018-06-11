package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/urfave/cli"
	"log"
	"strings"
)

func showInstances(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)

	instances, err := fakewarp.GetInstances(volumeRegistry)
	if err != nil {
		return err
	}
	fmt.Println(instances)
	return nil
}

func showSessions(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)

	sessions, err := fakewarp.GetSessions(volumeRegistry)
	if err != nil {
		return err
	}
	fmt.Println(sessions)
	return nil
}

func listPools(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)

	pools, err := fakewarp.GetPools(poolRegistry)
	if err != nil {
		return err
	}
	fmt.Println(pools)
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
	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	error := fakewarp.DeleteBuffer(c, volReg)
	return error
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
	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	error := fakewarp.CreatePerJobBuffer(c, volReg, lines)
	return error
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

var testKeystore keystoreregistry.Keystore
var lines fakewarp.GetLines

func getKeystore() keystoreregistry.Keystore {
	// TODO must be a better way to test this, proper factory?
	keystore := testKeystore
	if keystore == nil {
		keystore = etcdregistry.NewKeystore()
	}
	if lines == nil {
		lines = fakewarp.LinesFromFile{}
	}
	return keystore
}

func createPersistent(c *cli.Context) error {
	checkRequiredStrings(c, "token", "caller", "capacity", "user", "access", "type")
	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	name, error := fakewarp.CreatePersistentBuffer(c, volReg)
	if error == nil {
		// Slurm is looking for the string "created" to know this worked
		fmt.Printf("created %s\n", name)
	}
	return error
}
