package ansible

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
	"strings"
	"time"
	"path"
)

type HostInfo struct {
	MGS  string         `yaml:"mgs,omitempty"`
	MDTS string         `yaml:"mdt,omitempty"`
	OSTS map[string]int `yaml:"osts,omitempty,flow"`
}

type FSInfo struct {
	Hosts map[string]HostInfo `yaml:"hosts"`
	Vars  map[string]string   `yaml:"vars"`
}

type FileSystems struct {
	Children map[string]FSInfo `yaml:"children"`
}

type Wrapper struct {
	All FileSystems
}

// TODO: should come from configuration?
var hostGroup = "dac-fake"

func getInventory(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) string {
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
			hostInfo.MDTS = mdt.Device
			if fsType == Lustre {
				hostInfo.MGS = "nvme0n1" // TODO: horrible hack!!
			} else {
				hostInfo.MGS = mdt.Device
			}
		}
		hosts[host] = hostInfo
	}
	for host := range volume.Attachments {
		hosts[host] = HostInfo{}
	}
	fsinfo := FSInfo{
		Vars: map[string]string{
			"mgsnode":     mdt.Hostname,
			"client_port": fmt.Sprintf("%d", volume.ClientPort)},
		Hosts: hosts,
	}
	fsname := fmt.Sprintf("%s", volume.UUID)
	data := Wrapper{All: FileSystems{Children: map[string]FSInfo{fsname: fsinfo}}}

	output, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	strOut := string(output)
	strOut = strings.Replace(strOut, " mgs:", fmt.Sprintf(" %s_mgs:", fsname), -1)
	strOut = strings.Replace(strOut, " mdt:", fmt.Sprintf(" %s_mdt:", fsname), -1)
	strOut = strings.Replace(strOut, " osts:", fmt.Sprintf(" %s_osts:", fsname), -1)
	strOut = strings.Replace(strOut, " mgsnode:", fmt.Sprintf(" %s_mgsnode:", fsname), -1)
	strOut = strings.Replace(strOut, " client_port:", fmt.Sprintf(" %s_client_port:", fsname), -1)
	strOut = strings.Replace(strOut, "all:", hostGroup+":", -1)
	return strOut
}

func getPlaybook(fsType FSType, volume registry.Volume) string {
	role := "lustre"
	if fsType == BeegFS {
		role = "beegfs"
	}
	return fmt.Sprintf(`---
- name: Setup FS
  hosts: %s
  any_errors_fatal: true
  become: yes
  roles:
    - role: %s
      vars:
        fs_name: %s`, volume.UUID, role, volume.UUID)
}

func getAnsibleDir(suffix string) string {
	ansibleDir := os.Getenv("DAC_ANSIBLE_DIR")
	if ansibleDir == "" {
		ansibleDir = "/var/lib/data-acc/fs-ansbile/"
	}
	return path.Join(ansibleDir, suffix)
}

func setupAnsible(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) (string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("fs%s_", volume.Name))
	if err != nil {
		return dir, err
	}
	log.Println("Using ansible tempdir:", dir)

	playbook := getPlaybook(fsType, volume)
	tmpPlaybook := filepath.Join(dir, "dac.yml")
	if err := ioutil.WriteFile(tmpPlaybook, bytes.NewBufferString(playbook).Bytes(), 0666); err != nil {
		return dir, err
	}
	log.Println(playbook)

	inventory := getInventory(fsType, volume, brickAllocations)
	tmpInventory := filepath.Join(dir, "inventory")
	if err := ioutil.WriteFile(tmpInventory, bytes.NewBufferString(inventory).Bytes(), 0666); err != nil {
		return dir, err
	}
	log.Println(inventory)

	cmd := exec.Command("cp", "-r", getAnsibleDir("roles"), dir)
	output, err := cmd.CombinedOutput()
	log.Println("copy roles", string(output))
	if err != nil {
		return dir, err
	}
	cmd = exec.Command("cp", "-r", getAnsibleDir(".venv"), dir)
	output, err = cmd.CombinedOutput()
	log.Println("copy venv", string(output))
	if err != nil {
		return dir, err
	}
	cmd = exec.Command("cp", "-r", getAnsibleDir("group_vars"), dir)
	output, err = cmd.CombinedOutput()
	log.Println("copy group vars", string(output))
	return dir, err
}

func executeAnsibleSetup(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	dir, err := setupAnsible(fsType, volume, brickAllocations)
	if err != nil {
		return err
	}

	formatArgs := "dac.yml -i inventory --tag format"
	err = executeAnsiblePlaybook(dir, formatArgs)
	if err != nil {
		return err
	}

	startupArgs := "dac.yml -i inventory --tag mount,create_mdt,create_mgs,create_osts,client_mount"
	err = executeAnsiblePlaybook(dir, startupArgs)
	if err != nil {
		return err
	}

	// only delete if everything worked, to aid debugging
	os.RemoveAll(dir)
	return nil
}

func executeAnsibleTeardown(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	dir, err := setupAnsible(fsType, volume, brickAllocations)
	if err != nil {
		return err
	}

	stopArgs := "dac.yml -i inventory --tag stop_all,unmount,client_unmount"
	err = executeAnsiblePlaybook(dir, stopArgs)
	if err != nil {
		return err
	}

	formatArgs := "dac.yml -i inventory --tag format"
	err = executeAnsiblePlaybook(dir, formatArgs)
	if err != nil {
		return err
	}

	// only delete if everything worked, to aid debugging
	os.RemoveAll(dir)
	return nil
}

func executeAnsibleMount(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	dir, err := setupAnsible(fsType, volume, brickAllocations)
	if err != nil {
		return err
	}

	startupArgs := "dac.yml -i inventory --tag client_mount"
	err = executeAnsiblePlaybook(dir, startupArgs)
	if err != nil {
		return err
	}

	os.RemoveAll(dir)
	return nil
}

func executeAnsibleUnmount(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	dir, err := setupAnsible(fsType, volume, brickAllocations)
	if err != nil {
		return err
	}

	stopArgs := "dac.yml -i inventory --tag client_unmount"
	err = executeAnsiblePlaybook(dir, stopArgs)
	if err != nil {
		return err
	}

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
