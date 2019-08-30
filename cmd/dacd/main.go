package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacd"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacd/brick_manager_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store_impl"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func waitForShutdown(manager dacd.BrickManager) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	log.Println("I have been asked to shutdown, doing tidy up...")
	manager.Shutdown()
	os.Exit(1)
}

func main() {
	log.Println("Starting data-accelerator's brick manager")

	keystore := store_impl.NewKeystore()
	defer func() {
		log.Println("keystore closed with error: ", keystore.Close())
	}()

	manager := brick_manager_impl.NewBrickManager(keystore)
	manager.Startup()

	log.Println("Brick manager started for:", manager.Hostname())

	waitForShutdown(manager)
}
