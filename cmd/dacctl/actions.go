package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/urfave/cli"
)

var testKeystore keystoreregistry.Keystore
var testDisk fileio.Disk
var testActions actions.FakewarpActions

func getKeystore() keystoreregistry.Keystore {
	// TODO must be a better way to test this, proper factory?
	if testKeystore != nil {
		return testKeystore
	}
	return etcdregistry.NewKeystore()
}

func getActions(keystore keystoreregistry.Keystore) actions.FakewarpActions {
	if testActions != nil {
		return testActions
	}
	volReg := keystoreregistry.NewVolumeRegistry(keystore)
	poolReg := keystoreregistry.NewPoolRegistry(keystore)
	disk := testDisk
	if testDisk == nil {
		disk = fileio.NewDisk()
	}
	return actions.NewFakewarpActions(poolReg, volReg, disk)
}

func createPersistent(c *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).CreatePersistentBuffer(c)
}

func showInstances(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ShowInstances()
}

func showSessions(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ShowSessions()
}

func listPools(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ListPools()
}

func showConfigurations(_ *cli.Context) error {
	keystore := getKeystore()
	defer keystore.Close()
	return getActions(keystore).ShowConfigurations()
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
	return getActions(keystore).RealSize(c)
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
