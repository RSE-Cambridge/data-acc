package fake

import (
	"bytes"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type HostInfo struct {
	MGS  string         `yaml:"mgs,omitempty"`
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
			hostInfo.MGS = "nvme0n1" // TODO: horrible hack!!
		}
		hosts[host] = hostInfo
	}
	fsinfo := FSInfo{
		Vars:  map[string]string{"mgsnode": mdt.Hostname},
		Hosts: hosts,
	}
	fsname := fmt.Sprintf("%s", volume.UUID)
	data := Wrapper{All: FileSystems{Children: map[string]FSInfo{fsname: fsinfo}}}

	output, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	return string(output)
}

func printLustrePlaybook(volume registry.Volume) string {
	return fmt.Sprintf(`---
- name: Install Lustre
  hosts: %s
  any_errors_fatal: true
  become: yes
  gather_facts: no
  roles:
    - role: lustre`, volume.UUID)
}

func executeTempAnsible(volume registry.Volume, brickAllocations []registry.BrickAllocation, teardown bool) error {
	dir, err := ioutil.TempDir("", fmt.Sprintf("fs%s_", volume.Name))
	if err != nil {
		return err
	}
	log.Println("Using ansible tempdir:", dir)

	playbook := printLustrePlaybook(volume)
	tmpPlaybook := filepath.Join(dir, "dac.yml")
	if err := ioutil.WriteFile(tmpPlaybook, bytes.NewBufferString(playbook).Bytes(), 0666); err != nil {
		return err
	}
	log.Println(playbook)

	inventory := printLustreInfo(volume, brickAllocations)
	tmpInventory := filepath.Join(dir, "inventory")
	if err := ioutil.WriteFile(tmpInventory, bytes.NewBufferString(inventory).Bytes(), 0666); err != nil {
		return err
	}
	log.Println(inventory)

	cmd := exec.Command("cp", "-r",
		"/home/centos/go/src/github.com/JohnGarbutt/data-acc/fs-ansible/environment/roles", dir)
	output, err := cmd.CombinedOutput()
	log.Println("copy roles", string(output))
	if err != nil {
		return err
	}
	cmd = exec.Command("cp", "-r",
		"/home/centos/go/src/github.com/JohnGarbutt/data-acc/fs-ansible/environment/.venv", dir)
	output, err = cmd.CombinedOutput()
	log.Println("copy venv", string(output))
	if err != nil {
		return err
	}

	if !teardown {
		formatArgs := "dac.yml -i inventory --tag format_mgs --tag reformat_mdts --tag reformat_osts"
		err = executeAnsiblePlaybook(dir, formatArgs)
		if err != nil {
			return err
		}

		startupArgs := "dac.yml -i inventory --tag start_mgs --tag start_mdts --tag start_osts --tag mount_fs"
		err = executeAnsiblePlaybook(dir, startupArgs)
		if err != nil {
			return err
		}

	} else {
		stopArgs := "dac.yml -i inventory --tag umount_fs --tag stop_osts --tag stop_mdts"
		err = executeAnsiblePlaybook(dir, stopArgs)
		if err != nil {
			return err
		}

		formatArgs := "dac.yml -i inventory --tag reformat_mdts --tag reformat_osts"
		err = executeAnsiblePlaybook(dir, formatArgs)
		if err != nil {
			return err
		}
	}

	// only delete if everything worked, to aid debugging
	os.RemoveAll(dir)
	return nil
}

func executeAnsiblePlaybook(dir string, args string) error {
	// TODO: downgrade debug log!
	cmdStr := fmt.Sprintf(`cd %s; . .venv/bin/activate; ansible-playbook %s;`, dir, args)
	log.Println("Requested ansible:", cmdStr)

	var err error
	for i := 1; i <= 5; i++ {
		cmdMount := exec.Command("mount")
		output, _ := cmdMount.CombinedOutput()
		log.Println("Current mounts:", string(output))

		log.Println("Attempt", i, "of ansible:", cmdStr)
		cmd := exec.Command("bash", "-c", cmdStr)
		output, currentErr := cmd.CombinedOutput()

		if currentErr == nil {
			log.Println("Completed ansible run:", cmdStr)
			log.Println(string(output))
			return nil
		} else {
			log.Println("Error in ansible run:", string(output))
			err = currentErr
			time.Sleep(time.Second * 2)
		}
	}
	return err
}
