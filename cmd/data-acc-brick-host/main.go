package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/brickmanager"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func waitForShutdown() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	log.Println("I have been asked to shutdown, doing tidy up...")
	os.Exit(1)
}

func main() {
	log.Println("Starting data-accelerator's brick manager")

	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()
	poolRegistry := keystoreregistry.NewPoolRegistry(keystore)
	volumeRegistry := keystoreregistry.NewVolumeRegistry(keystore)

	manager := brickmanager.NewBrickManager(poolRegistry, volumeRegistry)
	if err := manager.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("Brick manager started for:", manager.Hostname())

	waitForShutdown()
}
