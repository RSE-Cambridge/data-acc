package ansible

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlugin_GetInventory(t *testing.T) {
	volume := registry.Volume{
		Name: "1", UUID: "abcdefgh", ClientPort: 10002,
		Attachments: []registry.Attachment{
			{Hostname: "cpu1"},
			{Hostname: "cpu2"},
		},
	}
	brickAllocations := []registry.BrickAllocation{
		{Hostname: "dac1", Device: "nvme1n1", AllocatedIndex: 0},
		{Hostname: "dac1", Device: "nvme2n1", AllocatedIndex: 1},
		{Hostname: "dac1", Device: "nvme3n1", AllocatedIndex: 2},
		{Hostname: "dac2", Device: "nvme2n1", AllocatedIndex: 3},
		{Hostname: "dac2", Device: "nvme3n1", AllocatedIndex: 4},
	}
	result := getInventory(BeegFS, volume, brickAllocations)
	expected := `dac-prod:
  children:
    abcdefgh:
      hosts:
        cpu1: {}
        cpu2: {}
        dac1:
          abcdefgh_mgs: nvme1n1
          abcdefgh_mdts: {nvme1n1: 0, nvme2n1: 1, nvme3n1: 2}
          abcdefgh_osts: {nvme1n1: 0, nvme2n1: 1, nvme3n1: 2}
        dac2:
          abcdefgh_mdts: {nvme2n1: 3, nvme3n1: 4}
          abcdefgh_osts: {nvme2n1: 3, nvme3n1: 4}
      vars:
        abcdefgh_client_port: "10002"
        lnet_suffix: ""
        abcdefgh_mgsnode: dac1
`
	assert.Equal(t, expected, result)
}

func TestPlugin_GetInventory_withNoOstOnOneHost(t *testing.T) {
	volume := registry.Volume{Name: "1", UUID: "abcdefgh", ClientPort: 10002}
	brickAllocations := []registry.BrickAllocation{
		{Hostname: "dac1", Device: "nvme1n1", AllocatedIndex: 0},
		{Hostname: "dac2", Device: "nvme2n1", AllocatedIndex: 1},
	}
	result := getInventory(Lustre, volume, brickAllocations)
	expected := `dac-prod:
  children:
    abcdefgh:
      hosts:
        dac1:
          abcdefgh_mgs: sdb
          abcdefgh_mdts: {nvme1n1: 0}
          abcdefgh_osts: {nvme1n1: 0}
        dac2:
          abcdefgh_mdts: {nvme2n1: 1}
          abcdefgh_osts: {nvme2n1: 1}
      vars:
        abcdefgh_client_port: "10002"
        lnet_suffix: ""
        abcdefgh_mgsnode: dac1
`
	assert.Equal(t, expected, result)
}

func TestPlugin_GetPlaybook_beegfs(t *testing.T) {
	volume := registry.Volume{Name: "1", UUID: "abcdefgh"}
	result := getPlaybook(BeegFS, volume)
	assert.Equal(t, `---
- name: Setup FS
  hosts: abcdefgh
  any_errors_fatal: true
  become: yes
  roles:
    - role: beegfs
      vars:
        fs_name: abcdefgh`, result)
}

func TestPlugin_GetPlaybook_lustre(t *testing.T) {
	volume := registry.Volume{Name: "1", UUID: "abcdefgh"}
	result := getPlaybook(Lustre, volume)
	assert.Equal(t, `---
- name: Setup FS
  hosts: abcdefgh
  any_errors_fatal: true
  become: yes
  roles:
    - role: lustre
      vars:
        fs_name: abcdefgh`, result)
}
