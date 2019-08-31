package filesystem_impl

import (
	"bytes"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

var conf = config.GetFilesystemConfig()

func getInventory(fsType FSType, fsUuid string, allBricks []datamodel.Brick) string {
	allocationByHost := make(map[datamodel.BrickHostName][]datamodel.BrickAllocation)
	for i, brick := range allBricks {
		allocationByHost[brick.BrickHostName] = append(allocationByHost[brick.BrickHostName], datamodel.BrickAllocation{
			Brick:          brick,
			AllocatedIndex: uint(i),
		})
	}

	// If we have more brick allocations than maxMDTs
	// assign at most one mdt per host.
	// While this may give us less MDTs than max MDTs,
	// but it helps spread MDTs across network connections
	oneMdtPerHost := len(allBricks) > int(conf.MaxMDTs)

	hosts := make(map[string]HostInfo)
	mgsnode := ""
	for host, allocations := range allocationByHost {
		osts := make(map[string]int)
		for _, allocation := range allocations {
			osts[allocation.Brick.Device] = int(allocation.AllocatedIndex)
		}

		mdts := make(map[string]int)
		if oneMdtPerHost {
			allocation := allocations[0]
			mdts[allocation.Brick.Device] = int(allocation.AllocatedIndex)
		} else {
			for _, allocation := range allocations {
				mdts[allocation.Brick.Device] = int(allocation.AllocatedIndex)
			}
		}

		hostInfo := HostInfo{MDTS: mdts, OSTS: osts}

		if allocations[0].AllocatedIndex == 0 {
			if fsType == Lustre {
				hostInfo.MGS = conf.MGSDevice
			} else {
				hostInfo.MGS = allocations[0].Brick.Device
			}
			mgsnode = string(host)
		}
		hosts[string(host)] = hostInfo
	}

	// TODO: add attachments?

	fsinfo := FSInfo{
		Vars: map[string]string{
			"mgsnode": mgsnode,
			//"client_port": fmt.Sprintf("%d", volume.ClientPort),
			"lnet_suffix": conf.LnetSuffix,
			"mdt_size":    fmt.Sprintf("%dm", conf.MDTSizeMB),
		},
		Hosts: hosts,
	}
	fsname := fmt.Sprintf("%s", fsUuid)
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

	strOut = strings.Replace(strOut, "all:", conf.HostGroup+":", -1)
	return strOut
}

func getPlaybook(fsType FSType, fsUuid string) string {
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
        fs_name: %s`, fsUuid, role, fsUuid)
}

func getAnsibleDir(suffix string) string {
	return path.Join(conf.AnsibleDir, suffix)
}

func setupAnsible(fsType FSType, internalName string, bricks []datamodel.Brick) (string, error) {
	if len(bricks) == 0 {
		log.Panicf("can't create filesystem with no bricks: %s", internalName)
	}

	dir, err := ioutil.TempDir("", fmt.Sprintf("fs%s_", internalName))
	if err != nil {
		return dir, err
	}
	log.Println("Using ansible tempdir:", dir)

	playbook := getPlaybook(fsType, internalName)
	tmpPlaybook := filepath.Join(dir, "dac.yml")
	if err := ioutil.WriteFile(tmpPlaybook, bytes.NewBufferString(playbook).Bytes(), 0666); err != nil {
		return dir, err
	}
	log.Println(playbook)

	inventory := getInventory(fsType, internalName, bricks)
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

func executeAnsibleSetup(internalName string, bricks []datamodel.Brick, doFormat bool) error {
	// TODO: restore beegfs support
	dir, err := setupAnsible(Lustre, internalName, bricks)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	// allow skip format when trying to rebuild
	if doFormat {
		formatArgs := "dac.yml -i inventory --tag format"
		err = executeAnsiblePlaybook(dir, formatArgs)
		if err != nil {
			return fmt.Errorf("error during server format: %s", err.Error())
		}
	}

	startupArgs := "dac.yml -i inventory --tag mount,create_mdt,create_mgs,create_osts,client_mount"
	err = executeAnsiblePlaybook(dir, startupArgs)
	if err != nil {
		return fmt.Errorf("error during create fs: %s", err.Error())
	}
	return nil
}

func executeAnsibleTeardown(internalName string, bricks []datamodel.Brick) error {
	dir, err := setupAnsible(Lustre, internalName, bricks)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	stopArgs := "dac.yml -i inventory --tag stop_all,unmount,client_unmount"
	err = executeAnsiblePlaybook(dir, stopArgs)
	if err != nil {
		return fmt.Errorf("error during server umount: %s", err.Error())
	}

	formatArgs := "dac.yml -i inventory --tag format"
	err = executeAnsiblePlaybook(dir, formatArgs)
	if err != nil {
		return fmt.Errorf("error during server clean: %s", err.Error())
	}
	return nil
}

func executeAnsiblePlaybook(dir string, args string) error {
	// TODO: downgrade debug log!
	cmdStr := fmt.Sprintf(`cd %s; . .venv/bin/activate; ansible-playbook %s;`, dir, args)
	log.Println("Requested ansible:", cmdStr)

	if conf.SkipAnsible {
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
			if err := cmd.Process.Kill(); err != nil {
				log.Panicf("error trying to kill process: %s", err.Error())
			}
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
