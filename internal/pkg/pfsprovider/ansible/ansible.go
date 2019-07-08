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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type HostInfo struct {
	MGS  string         `yaml:"mgs,omitempty"`
	MDTS map[string]int `yaml:"mdts,omitempty,flow"`
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

var DefaultHostGroup = "dac-prod"
var DefaultMaxMDTs uint = 24

func getInventory(fsType FSType, volume registry.Volume, brickAllocations []registry.BrickAllocation) string {
	// NOTE: only used by lustre
	mgsDevice := os.Getenv("DAC_MGS_DEV")
	if mgsDevice == "" {
		mgsDevice = "sdb"
	}
	maxMDTs := DefaultMaxMDTs
	maxMDTsConf, err := strconv.ParseUint(os.Getenv("DAC_MAX_MDT_COUNT"), 10, 32)
	if err == nil && maxMDTsConf > 0 {
		maxMDTs = uint(maxMDTsConf)
	}

	allocationsByHost := make(map[string][]registry.BrickAllocation)
	for _, allocation := range brickAllocations {
		allocationsByHost[allocation.Hostname] = append(allocationsByHost[allocation.Hostname], allocation)
	}

	// If we have more brick allocations than maxMDTs
	// assign at most one mdt per host.
	// While this may give us less MDTs than max MDTs,
	// but it helps spread MDTs across network connections
	oneMdtPerHost := len(brickAllocations) > int(maxMDTs)

	hosts := make(map[string]HostInfo)
	mgsnode := ""
	for host, allocations := range allocationsByHost {
		osts := make(map[string]int)
		for _, allocation := range allocations {
			osts[allocation.Device] = int(allocation.AllocatedIndex)
		}

		mdts := make(map[string]int)
		if oneMdtPerHost {
			allocation := allocations[0]
			mdts[allocation.Device] = int(allocation.AllocatedIndex)
		} else {
			for _, allocation := range allocations {
				mdts[allocation.Device] = int(allocation.AllocatedIndex)
			}
		}

		hostInfo := HostInfo{MDTS: mdts, OSTS: osts}

		if allocations[0].AllocatedIndex == 0 {
			if fsType == Lustre {
				hostInfo.MGS = mgsDevice
			} else {
				hostInfo.MGS = allocations[0].Device
			}
			mgsnode = host
		}
		hosts[host] = hostInfo
	}

	// for beegfs, mount clients via ansible
	if fsType == BeegFS {
		// TODO: this can't work now, as we need to also pass the job name
		for _, attachment := range volume.Attachments {
		  hosts[attachment.Hostname] = HostInfo{}
		}
	}


	fsinfo := FSInfo{
		Vars: map[string]string{
			"mgsnode":     mgsnode,
			"client_port": fmt.Sprintf("%d", volume.ClientPort),
			"lnet_suffix": getLnetSuffix(),
			"mdt_size":    fmt.Sprintf("%dm", getMdtSizeMB()),
		},
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
	strOut = strings.Replace(strOut, " mdts:", fmt.Sprintf(" %s_mdts:", fsname), -1)
	strOut = strings.Replace(strOut, " osts:", fmt.Sprintf(" %s_osts:", fsname), -1)
	strOut = strings.Replace(strOut, " mgsnode:", fmt.Sprintf(" %s_mgsnode:", fsname), -1)
	strOut = strings.Replace(strOut, " client_port:", fmt.Sprintf(" %s_client_port:", fsname), -1)
	strOut = strings.Replace(strOut, " mdt_size:", fmt.Sprintf(" %s_mdt_size:", fsname), -1)

	hostGroup := os.Getenv("DAC_HOST_GROUP")
	if hostGroup == "" {
		hostGroup = DefaultHostGroup
	}
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
		ansibleDir = "/var/lib/data-acc/fs-ansible/"
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

	formatArgs := "dac.yml -i inventory --tag clean"
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

	skipAnsible := os.Getenv("DAC_SKIP_ANSIBLE")
	if skipAnsible == "True" {
		log.Println("Skip as DAC_SKIP_ANSIBLE=True")
		time.Sleep(time.Millisecond * 200)
		return nil
	}

	var err error
	for i := 1; i <= 3; i++ {
		log.Println("Attempt", i, "of ansible:", cmdStr)
		cmd := exec.Command("bash", "-c", cmdStr)

		timer := time.AfterFunc(time.Minute*5, func() {
			log.Println("Time up, waited more than 5 mins to complete.")
			cmd.Process.Kill()
		})
		output, currentErr := cmd.CombinedOutput()
		timer.Stop()

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
