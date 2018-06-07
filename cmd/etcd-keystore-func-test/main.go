package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry"
)

func main() {
	keystore := etcdregistry.NewKeystore()
	defer keystore.Close()

	TestEtcdKeystore(keystore)
	fmt.Println("")

	TestKeystorePoolRegistry(keystore)
}
