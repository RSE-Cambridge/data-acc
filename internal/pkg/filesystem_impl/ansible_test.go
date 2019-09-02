package filesystem_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlugin_GetInventory(t *testing.T) {
	brickAllocations := []datamodel.Brick{
		{BrickHostName: "dac1", Device: "nvme1n1"},
		{BrickHostName: "dac1", Device: "nvme2n1"},
		{BrickHostName: "dac1", Device: "nvme3n1"},
		{BrickHostName: "dac2", Device: "nvme2n1"},
		{BrickHostName: "dac2", Device: "nvme3n1"},
	}
	fsUuid := "abcdefgh"
	result := getInventory(BeegFS, fsUuid, brickAllocations)
	expected := `dacs:
  children:
    abcdefgh:
      hosts:
        dac1:
          mgs: nvme1n1
          mdts: {nvme1n1: 0, nvme2n1: 1, nvme3n1: 2}
          osts: {nvme1n1: 0, nvme2n1: 1, nvme3n1: 2}
        dac2:
          mdts: {nvme2n1: 3, nvme3n1: 4}
          osts: {nvme2n1: 3, nvme3n1: 4}
      vars:
        fs_name: abcdefgh
        lnet_suffix: ""
        mdt_size_mb: "20480"
        mgsnode: dac1
`
	assert.Equal(t, expected, result)
}

func TestPlugin_GetInventory_withNoOstOnOneHost(t *testing.T) {
	brickAllocations := []datamodel.Brick{
		{BrickHostName: "dac1", Device: "nvme1n1"},
		{BrickHostName: "dac2", Device: "nvme2n1"},
		{BrickHostName: "dac2", Device: "nvme3n1"},
	}
	fsUuid := "abcdefgh"
	result := getInventory(Lustre, fsUuid, brickAllocations)
	expected := `dacs:
  children:
    abcdefgh:
      hosts:
        dac1:
          mgs: sdb
          mdts: {nvme1n1: 0}
          osts: {nvme1n1: 0}
        dac2:
          mdts: {nvme2n1: 1, nvme3n1: 2}
          osts: {nvme2n1: 1, nvme3n1: 2}
      vars:
        fs_name: abcdefgh
        lnet_suffix: ""
        mdt_size_mb: "20480"
        mgsnode: dac1
`
	assert.Equal(t, expected, result)
}

func TestPlugin_GetPlaybook_beegfs(t *testing.T) {
	fsUuid := "abcdefgh"
	result := getPlaybook(BeegFS, fsUuid)
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
	fsUuid := "abcdefgh"
	result := getPlaybook(Lustre, fsUuid)
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

func TestPlugin_GetInventory_MaxMDT(t *testing.T) {
	var brickAllocations []datamodel.Brick
	for i := 1; i <= 26; i = i + 2 {
		brickAllocations = append(brickAllocations, datamodel.Brick{
			BrickHostName: datamodel.BrickHostName(fmt.Sprintf("dac%d", i)),
			Device:        "nvme1n1",
		})
		brickAllocations = append(brickAllocations, datamodel.Brick{
			BrickHostName: datamodel.BrickHostName(fmt.Sprintf("dac%d", i)),
			Device:        "nvme2n1",
		})
	}

	fsUuid := "abcdefgh"
	result := getInventory(Lustre, fsUuid, brickAllocations)
	expected := `dacs:
  children:
    abcdefgh:
      hosts:
        dac1:
          mgs: sdb
          mdts: {nvme1n1: 0}
          osts: {nvme1n1: 0, nvme2n1: 1}
        dac3:
          mdts: {nvme1n1: 2}
          osts: {nvme1n1: 2, nvme2n1: 3}
        dac5:
          mdts: {nvme1n1: 4}
          osts: {nvme1n1: 4, nvme2n1: 5}
        dac7:
          mdts: {nvme1n1: 6}
          osts: {nvme1n1: 6, nvme2n1: 7}
        dac9:
          mdts: {nvme1n1: 8}
          osts: {nvme1n1: 8, nvme2n1: 9}
        dac11:
          mdts: {nvme1n1: 10}
          osts: {nvme1n1: 10, nvme2n1: 11}
        dac13:
          mdts: {nvme1n1: 12}
          osts: {nvme1n1: 12, nvme2n1: 13}
        dac15:
          mdts: {nvme1n1: 14}
          osts: {nvme1n1: 14, nvme2n1: 15}
        dac17:
          mdts: {nvme1n1: 16}
          osts: {nvme1n1: 16, nvme2n1: 17}
        dac19:
          mdts: {nvme1n1: 18}
          osts: {nvme1n1: 18, nvme2n1: 19}
        dac21:
          mdts: {nvme1n1: 20}
          osts: {nvme1n1: 20, nvme2n1: 21}
        dac23:
          mdts: {nvme1n1: 22}
          osts: {nvme1n1: 22, nvme2n1: 23}
        dac25:
          mdts: {nvme1n1: 24}
          osts: {nvme1n1: 24, nvme2n1: 25}
      vars:
        fs_name: abcdefgh
        lnet_suffix: ""
        mdt_size_mb: "20480"
        mgsnode: dac1
`
	assert.Equal(t, expected, result)
}
