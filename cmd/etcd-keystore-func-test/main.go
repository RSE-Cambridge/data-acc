package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"log"
)

func cleanAllKeys(keystore keystoreregistry.Keystore) {
	if err := keystore.CleanPrefix(""); err != nil {
		log.Println("Error cleaning: ", err)
	}
}

func main() {
	log.Println("Creating keystore")
	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()
	cleanAllKeys(keystore)
}
