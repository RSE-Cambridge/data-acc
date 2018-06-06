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

func testAddValues(keystore keystoreregistry.Keystore) {
	values := []keystoreregistry.KeyValue{
		keystoreregistry.KeyValue{},
	}
	if err := keystore.Add(values); err != nil {
		log.Fatalf("Error with add values")
	}
}

func main() {
	log.Println("Creating keystore")
	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()

	cleanAllKeys(keystore)

	testAddValues(keystore)
}
