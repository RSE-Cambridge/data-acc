package main

import (
	"log"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
)

func cleanAllKeys(keystore keystoreregistry.Keystore) {
	if err := keystore.CleanPrefix(""); err != nil {
		log.Println("Error cleaning: ", err)
	}
}

func main() {
	log.Println("Creating keystore")
	keystore := etcdregistry.NewKeystore()
	cleanAllKeys(keystore)
}
