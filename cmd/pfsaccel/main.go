package main

import (
	"fmt"
	"github.com/JohnGarbutt/pfsaccel/internal/pkg/oldregistry"
	"os/exec"
	"runtime"
	"sync"
)

func main() {
	fmt.Println("Hello from pfsaccel demo.")

	registry := oldregistry.NewBufferRegistry()
	defer registry.Close()

	// tidy up keys before we start and after we are finished
	registry.ClearAllData()
	defer registry.ClearAllData()

	// list of "available" slice ids
	sliceIds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sliceIndex := 0

	// watch for buffers, create slice on put
	var waitBuffer sync.WaitGroup
	makeSlice := func(key string, value string) {
		registry.AddSlice(sliceIds[sliceIndex], key)
		sliceIndex++
		waitBuffer.Done()
	}
	go registry.WatchNewBuffer(makeSlice)

	// watch for slice updates
	var waitSlice sync.WaitGroup
	printEvent := func(key string, value string) {
		bufferKey := value
		fakeMountpoint, err := exec.Command("date", "-u", "-Ins").Output()
		if err != nil {
			panic(err)
		}
		registry.AddMountpoint(bufferKey, string(fakeMountpoint))
		waitSlice.Done()
	}
	go registry.WatchNewSlice(printEvent)

	// watch for buffer setup complete
	printBufferReady := func(key string, value string) {
		fmt.Printf("Buffer ready %s with mountpoint %s", key, value)
	}
	go registry.WatchNewReady(printBufferReady)

	// add some fake buffers to test the watch
	ids := []int{1, 2, 3, 4, 5}
	for _, id := range ids {
		waitBuffer.Add(1)
		waitSlice.Add(1)
		registry.AddBuffer(id)
	}
	waitBuffer.Add(1)
	waitSlice.Add(1)
	registry.AddBuffer(16)

	runtime.Gosched()

	// Wait for all the buffer work to happen
	waitBuffer.Wait()
	waitSlice.Wait()
}
