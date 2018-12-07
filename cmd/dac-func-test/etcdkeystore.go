package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"log"
	"runtime"
	"time"
)

func cleanAllKeys(keystore keystoreregistry.Keystore) {
	if err := keystore.CleanPrefix(""); err != nil {
		log.Println("Error cleaning: ", err)
	}
}

func testAddValues(keystore keystoreregistry.Keystore) {
	values := []keystoreregistry.KeyValue{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	if err := keystore.Add(values); err != nil {
		log.Fatalf("Error with add values")
	}

	if err := keystore.Add(values); err == nil {
		log.Fatalf("Expected an error")
	} else {
		log.Println(err)
	}
}

func testGet(keystore keystoreregistry.Keystore) {
	value, _ := keystore.Get("key1")
	log.Println(value)
	_, err := keystore.Get("key3")
	if err == nil {
		log.Fatalf("failed to raise error")
	} else {
		log.Println(err)
	}

	values, _ := keystore.GetAll("key")
	log.Println(values)
	_, err = keystore.GetAll("key3")
	if err == nil {
		log.Fatalf("failed to raise error")
	} else {
		log.Println(err)
	}
}

func testUpdate(keystore keystoreregistry.Keystore) {
	values, err := keystore.GetAll("key")
	if err != nil {
		log.Fatal(err)
	}

	values[0].Value = "asdf"
	values[1].Value = "asdf2"

	err = keystore.Update(values)
	if err != nil {
		log.Fatal(err)
	}

	// Error if ModVersion out of sync
	err = keystore.Update(values)
	if err == nil {
		log.Fatal("Failed to raise error")
	} else {
		log.Println(err)
	}

	// Ensure success if told to ignore ModRevision
	values[0].ModRevision = 0
	values[1].ModRevision = 0
	values[1].Key = "key3" // add value via update
	err = keystore.Update(values)
	if err != nil {
		log.Fatal(err)
	}
}

func testDeleteAll(keystore keystoreregistry.Keystore) {
	values, err := keystore.GetAll("key")
	if err != nil {
		log.Fatal(err)
	}

	err = keystore.DeleteAll(values)
	if err != nil {
		log.Fatal(err)
	}
}

func testKeepAlive(keystore keystoreregistry.Keystore) {
	err := keystore.KeepAliveKey("/mytesthost")
	if err != nil {
		log.Fatal(err)
	}

	err = keystore.KeepAliveKey("/mytesthost")
	if err == nil {
		log.Fatal("expected error")
	} else {
		log.Println(err)
	}
}

func TestEtcdKeystore(keystore keystoreregistry.Keystore) {
	log.Println("Testing etcdkeystore...")

	keystore.WatchPrefix("ke",
		func(old *keystoreregistry.KeyValueVersion, new *keystoreregistry.KeyValueVersion) {
			log.Println("Watch spotted an update:")
			log.Println(" new:", new)
			log.Println(" old:", old)
		})

	testAddValues(keystore)
	testGet(keystore)
	testUpdate(keystore)
	testDeleteAll(keystore)
	testKeepAlive(keystore)

	// Give background things time to finish
	time.Sleep(time.Millisecond * 100)
	runtime.Gosched()
}
