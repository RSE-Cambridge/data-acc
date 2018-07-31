package fake

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"gopkg.in/yaml.v2"
	"log"
)

type HostInfo struct {
	MDTS []string       `yaml:"mdts,omitempty,flow"`
	OSTS map[string]int `yaml:"osts,omitempty,flow"`
}

type FSInfo struct {
	Hosts map[string]HostInfo `yaml:"hosts"`
	Vars  map[string]string   `yaml:"vars,flow"`
}

type FileSystems struct {
	Children map[string]FSInfo `yaml:"children"`
}

type Wrapper struct {
	All FileSystems
}

func printLustreInfo(volume registry.Volume, brickAllocations []registry.BrickAllocation) string {
	var mdt registry.BrickAllocation
	osts := make(map[string][]registry.BrickAllocation)
	for _, allocation := range brickAllocations {
		if allocation.AllocatedIndex == 0 {
			mdt = allocation
			current, success := osts[allocation.Hostname]
			if !success {
				// ensure hostname will be iterated through below to output mdt if required
				osts[allocation.Hostname] = current
			}
		} else {
			osts[allocation.Hostname] = append(osts[allocation.Hostname], allocation)
		}
	}

	hosts := make(map[string]HostInfo)
	for host, allocations := range osts {
		osts := make(map[string]int)
		for _, allocation := range allocations {
			osts[allocation.Device] = int(allocation.AllocatedIndex)
		}
		hostInfo := HostInfo{OSTS: osts}
		if mdt.Hostname == host {
			hostInfo.MDTS = []string{mdt.Device}
		}
		hosts[host] = hostInfo
	}
	fsinfo := FSInfo{
		Vars:  map[string]string{"mgsnode": mdt.Hostname},
		Hosts: hosts,
	}
	fsname := fmt.Sprintf("fs%s", volume.Name)
	data := Wrapper{All: FileSystems{Children: map[string]FSInfo{fsname: fsinfo}}}

	output, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	return string(output)
}
