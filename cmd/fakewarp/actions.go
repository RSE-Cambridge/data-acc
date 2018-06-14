package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
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
	poolReg := keystoreregistry.NewPoolRegistry(keystore)
	error := fakewarp.DeleteBuffer(c, volReg, poolReg)
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
	poolReg := keystoreregistry.NewPoolRegistry(keystore)
	error := fakewarp.CreatePerJobBuffer(c, volReg, poolReg, lines)
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

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	volumeName := registry.VolumeName(c.String("token"))
	err := volReg.UpdateState(volumeName, registry.DataInRequested)
	if err != nil {
		return err
	}
	return volReg.UpdateState(volumeName, registry.DataInComplete) // TODO should wait for host manager to do this
}

func paths(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "pathfile")
	fmt.Printf("--token %s --job %s --pathfile %s\n",
		c.String("token"), c.String("job"), c.String("pathfile"))
	// TODO get paths from the volume, and write out paths to given file
	return nil
}

func preRun(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job", "nodehostnamefile")
	fmt.Printf("--token %s --job %s --nodehostnamefile %s\n",
		c.String("token"), c.String("job"), c.String("nodehostnamefile"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	volumeName := registry.VolumeName(c.String("token"))
	err := volReg.UpdateState(volumeName, registry.MountRequested)
	if err != nil {
		return err
	}
	return volReg.UpdateState(volumeName, registry.MountComplete) // TODO should wait for host manager to do this
}

func postRun(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	volumeName := registry.VolumeName(c.String("token"))
	err := volReg.UpdateState(volumeName, registry.UnmountRequested)
	if err != nil {
		return err
	}
	return volReg.UpdateState(volumeName, registry.UnmountComplete) // TODO should wait for host manager to do this
}

func dataOut(c *cli.Context) error {
	checkRequiredStrings(c, "token", "job")
	fmt.Printf("--token %s --job %s\n",
		c.String("token"), c.String("job"))

	keystore := getKeystore()
	defer keystore.Close()
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	volumeName := registry.VolumeName(c.String("token"))
	err := volReg.UpdateState(volumeName, registry.DataOutRequested)
	if err != nil {
		return err
	}
	return volReg.UpdateState(volumeName, registry.DataOutComplete) // TODO should wait for host manager to do this
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
	poolReg := keystoreregistry.NewPoolRegistry(keystore)
	name, error := fakewarp.CreatePersistentBuffer(c, volReg, poolReg)
	if error == nil {
		// Slurm is looking for the string "created" to know this worked
		fmt.Printf("created %s\n", name)
	}
	return error
}
