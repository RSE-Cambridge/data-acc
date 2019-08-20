package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl/actions_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store_impl"
	"github.com/urfave/cli"
)

var testKeystore store.Keystore
var testDisk fileio.Disk
var testActions dacctl.DacctlActions

func getKeystore() store.Keystore {
	// TODO must be a better way to test this, proper factory?
	if testKeystore != nil {
		return testKeystore
	}
	return store_impl.NewKeystore()
}

func getActions(keystore store.Keystore) dacctl.DacctlActions {
	if testActions != nil {
		return testActions
	}
	disk := fileio.NewDisk()
	return actions_impl.NewDacctlActions(keystore, disk)
}

func createPersistent(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	err := getActions(keystore).CreatePersistentBuffer(c)
	if err == nil {
		// Slurm is looking for the string "created" to know this worked
		fmt.Printf("created %s\n", c.String("token"))
	}
	return err
}

func printOutput(function func() (string, error)) error {
	sessions, err := function()
	if err != nil {
		return err
	}
	fmt.Println(sessions)
	return nil
}

func showInstances(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return printOutput(getActions(keystore).ShowInstances)
}

func showSessions(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return printOutput(getActions(keystore).ShowSessions)
}

func listPools(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return printOutput(getActions(keystore).ListPools)
}

func showConfigurations(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return printOutput(getActions(keystore).ShowConfigurations)
}

func teardown(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).DeleteBuffer(c)
}

func jobProcess(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ValidateJob(c)
}

func setup(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).CreatePerJobBuffer(c)
}

func realSize(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return printOutput(func() (s string, e error) {
		return getActions(keystore).RealSize(c)
	})
}

func dataIn(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).DataIn(c)
}

func paths(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).Paths(c)
}

func preRun(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).PreRun(c)
}

func postRun(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).PostRun(c)
}

func dataOut(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).DataOut(c)
}
