package filesystem_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/filesystem"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

type HostInfo struct {
	hostName string
	MGS      string         `yaml:"mgs,omitempty"`
	MDTS     map[string]int `yaml:"mdts,omitempty,flow"`
	OSTS     map[string]int `yaml:"osts,omitempty,flow"`
}

type FSInfo struct {
	Hosts map[string]HostInfo `yaml:"hosts"`
	Vars  map[string]string   `yaml:"vars"`
}

type FileSystems struct {
	Children map[string]FSInfo `yaml:"children"`
}

type Wrapper struct {
	Dacs FileSystems
}

func NewAnsible() filesystem.Ansible {
	return &ansibleImpl{}
}

type ansibleImpl struct {
}

func (*ansibleImpl) CreateEnvironment(session datamodel.Session) (string, error) {
	return setupAnsible(Lustre, session.FilesystemStatus.InternalName, session.AllocatedBricks)
}

var conf = config.GetFilesystemConfig()

func getFSInfo(fsType FSType, fsUuid string, allBricks []datamodel.Brick) FSInfo {
	// give all bricks an index, using the random ordering of allBricks
	var allAllocations []datamodel.BrickAllocation
	for i, brick := range allBricks {
		allAllocations = append(allAllocations, datamodel.BrickAllocation{
			Brick:          brick,
			AllocatedIndex: uint(i),
		})
	}

	// group allocations by host
	allocationByHost := make(map[datamodel.BrickHostName][]datamodel.BrickAllocation)
	var orderedHostNames []datamodel.BrickHostName
	for _, allocation := range allAllocations {
		brickHostName := allocation.Brick.BrickHostName
		if _, ok := allocationByHost[brickHostName]; !ok {
			orderedHostNames = append(orderedHostNames, brickHostName)
		}
		allocationByHost[brickHostName] = append(allocationByHost[brickHostName], allocation)
	}

	// If we have more brick allocations than maxMDTs,
	// spread out the mdts between the hosts,
	// and go for the same number of mdts on each host
	maxMdtsPerHost := len(allBricks)
	if len(allBricks) > int(conf.MaxMDTs) {
		maxMdtsPerHost = int(conf.MaxMDTs) / len(orderedHostNames)
	}

	// create the HostInfo object for each host, with correct OSTS and MDTS
	hosts := make(map[string]HostInfo)
	mgsnode := ""
	mdtIndex := 0
	for _, host := range orderedHostNames {
		allocations := allocationByHost[host]
		osts := make(map[string]int)
		for _, allocation := range allocations {
			osts[allocation.Brick.Device] = int(allocation.AllocatedIndex)
		}

		mdts := make(map[string]int)
		for i, allocation := range allocations {
			// don't go over maxMdtsPerHost
			if i >= maxMdtsPerHost {
				break
			}
			mdts[allocation.Brick.Device] = mdtIndex
			mdtIndex++
		}

		hostInfo := HostInfo{hostName: string(host), OSTS: osts, MDTS: mdts}

		isPrimaryBrick := allocations[0].AllocatedIndex == 0
		if isPrimaryBrick {
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
			"mdt_size_mb": fmt.Sprintf("%d", conf.MDTSizeMB),
			"fs_name":     fsUuid,
		},
		Hosts: hosts,
	}
	return fsinfo
}

func getInventory(fsType FSType, fsUuid string, allBricks []datamodel.Brick) string {
	fsinfo := getFSInfo(fsType, fsUuid, allBricks)
	fsname := fmt.Sprintf("%s", fsUuid)
	data := Wrapper{Dacs: FileSystems{Children: map[string]FSInfo{fsname: fsinfo}}}

	output, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	strOut := string(output)
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

	inventory := getInventory(fsType, internalName, bricks)
	tmpInventory := filepath.Join(dir, "inventory")
	if err := ioutil.WriteFile(tmpInventory, []byte(inventory), 0666); err != nil {
		return dir, err
	}
	log.Println(inventory)

	cmd := exec.Command("cp", "-r", getAnsibleDir("roles"), dir)
	output, err := cmd.CombinedOutput()
	log.Println("copy roles", string(output))
	if err != nil {
		return dir, err
	}

	for _, playbook := range []string{"create.yml", "delete.yml", "restore.yml"} {
		cmd = exec.Command("cp", getAnsibleDir(playbook), dir)
		output, err = cmd.CombinedOutput()
		log.Println("copy playbooks", playbook, string(output))
		if err != nil {
			return dir, err
		}
	}

	cmd = exec.Command("cp", "-r", getAnsibleDir(".venv"), dir)
	output, err = cmd.CombinedOutput()
	log.Println("copy venv", string(output))
	return dir, err
}

func executeAnsibleSetup(internalName string, bricks []datamodel.Brick, doFormat bool) error {
	// TODO: restore beegfs support
	dir, err := setupAnsible(Lustre, internalName, bricks)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("error removing %s due to %s\n", dir, err)
		}
	}()

	// allow skip format when trying to rebuild
	if doFormat {
		formatArgs := "create.yml -i inventory"
		err = executeAnsiblePlaybook(dir, formatArgs)
		if err != nil {
			return fmt.Errorf("error during ansible create: %s", err.Error())
		}
	} else {
		formatArgs := "restore.yml -i inventory"
		err = executeAnsiblePlaybook(dir, formatArgs)
		if err != nil {
			return fmt.Errorf("error during ansible create: %s", err.Error())
		}
	}
	return nil
}

func executeAnsibleTeardown(internalName string, bricks []datamodel.Brick) error {
	dir, err := setupAnsible(Lustre, internalName, bricks)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("error removing %s due to %s\n", dir, err)
		}
	}()

	formatArgs := "delete.yml -i inventory"
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

		timer := time.AfterFunc(time.Minute*10, func() {
			log.Println("Time up, waited more than 10 mins to complete.")
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
