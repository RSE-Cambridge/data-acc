package filesystem_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/config"
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
	for i := 1; i <= 4; i = i + 2 {
		brickAllocations = append(brickAllocations, datamodel.Brick{
			BrickHostName: datamodel.BrickHostName(fmt.Sprintf("dac%d", i)),
			Device:        "nvme3n1",
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
          osts: {nvme1n1: 0, nvme2n1: 1, nvme3n1: 26}
        dac3:
          mdts: {nvme1n1: 1}
          osts: {nvme1n1: 2, nvme2n1: 3, nvme3n1: 27}
        dac5:
          mdts: {nvme1n1: 2}
          osts: {nvme1n1: 4, nvme2n1: 5}
        dac7:
          mdts: {nvme1n1: 3}
          osts: {nvme1n1: 6, nvme2n1: 7}
        dac9:
          mdts: {nvme1n1: 4}
          osts: {nvme1n1: 8, nvme2n1: 9}
        dac11:
          mdts: {nvme1n1: 5}
          osts: {nvme1n1: 10, nvme2n1: 11}
        dac13:
          mdts: {nvme1n1: 6}
          osts: {nvme1n1: 12, nvme2n1: 13}
        dac15:
          mdts: {nvme1n1: 7}
          osts: {nvme1n1: 14, nvme2n1: 15}
        dac17:
          mdts: {nvme1n1: 8}
          osts: {nvme1n1: 16, nvme2n1: 17}
        dac19:
          mdts: {nvme1n1: 9}
          osts: {nvme1n1: 18, nvme2n1: 19}
        dac21:
          mdts: {nvme1n1: 10}
          osts: {nvme1n1: 20, nvme2n1: 21}
        dac23:
          mdts: {nvme1n1: 11}
          osts: {nvme1n1: 22, nvme2n1: 23}
        dac25:
          mdts: {nvme1n1: 12}
          osts: {nvme1n1: 24, nvme2n1: 25}
      vars:
        fs_name: abcdefgh
        lnet_suffix: ""
        mdt_size_mb: "20480"
        mgsnode: dac1
`
	assert.Equal(t, expected, result)
}

func TestPlugin_GetFSInfo_MaxMDT_lessHosts(t *testing.T) {
	var brickAllocations []datamodel.Brick
	for i := 1; i <= 5; i++ {
		for j := 1; j <= 6; j++ {
			brickAllocations = append(brickAllocations, datamodel.Brick{
				BrickHostName: datamodel.BrickHostName(fmt.Sprintf("dac%d", i)),
				Device:        fmt.Sprintf("nvme%dn1", j),
			})
		}
	}
	for i := 1; i <= 2; i++ {
		brickAllocations = append(brickAllocations, datamodel.Brick{
			BrickHostName: datamodel.BrickHostName(fmt.Sprintf("dac%d", i)),
			Device:        "nvme11n1",
		})
	}

	fsUuid := "abcdefgh"
	conf := config.GetFilesystemConfig()
	conf.MGSHost = "dac5"
	conf.MGSDevice = "loop0"
	result := getFSInfo(Lustre, fsUuid, brickAllocations, conf)
	resultStr := fmt.Sprintf("%+v", result.Hosts)
	expected := `map[` +
		`dac1:{hostName:dac1 MGS: MDTS:map[nvme1n1:0 nvme2n1:1 nvme3n1:2 nvme4n1:3] ` +
		`OSTS:map[nvme11n1:30 nvme1n1:0 nvme2n1:1 nvme3n1:2 nvme4n1:3 nvme5n1:4 nvme6n1:5]} ` +
		`dac2:{hostName:dac2 MGS: MDTS:map[nvme1n1:4 nvme2n1:5 nvme3n1:6 nvme4n1:7] ` +
		`OSTS:map[nvme11n1:31 nvme1n1:6 nvme2n1:7 nvme3n1:8 nvme4n1:9 nvme5n1:10 nvme6n1:11]} ` +
		`dac3:{hostName:dac3 MGS: MDTS:map[nvme1n1:8 nvme2n1:9 nvme3n1:10 nvme4n1:11] ` +
		`OSTS:map[nvme1n1:12 nvme2n1:13 nvme3n1:14 nvme4n1:15 nvme5n1:16 nvme6n1:17]} ` +
		`dac4:{hostName:dac4 MGS: MDTS:map[nvme1n1:12 nvme2n1:13 nvme3n1:14 nvme4n1:15] ` +
		`OSTS:map[nvme1n1:18 nvme2n1:19 nvme3n1:20 nvme4n1:21 nvme5n1:22 nvme6n1:23]} ` +
		`dac5:{hostName:dac5 MGS:loop0 MDTS:map[nvme1n1:16 nvme2n1:17 nvme3n1:18 nvme4n1:19] ` +
		`OSTS:map[nvme1n1:24 nvme2n1:25 nvme3n1:26 nvme4n1:27 nvme5n1:28 nvme6n1:29]}]`
	assert.Equal(t, expected, resultStr)

	conf.MGSHost = "slurmmaster1"
	result2 := getFSInfo(Lustre, fsUuid, brickAllocations, conf)
	resultStr2 := fmt.Sprintf("%+v", result2.Hosts)
	expected2 := `map[` +
		`dac1:{hostName:dac1 MGS: MDTS:map[nvme1n1:0 nvme2n1:1 nvme3n1:2 nvme4n1:3] ` +
		`OSTS:map[nvme11n1:30 nvme1n1:0 nvme2n1:1 nvme3n1:2 nvme4n1:3 nvme5n1:4 nvme6n1:5]} ` +
		`dac2:{hostName:dac2 MGS: MDTS:map[nvme1n1:4 nvme2n1:5 nvme3n1:6 nvme4n1:7] ` +
		`OSTS:map[nvme11n1:31 nvme1n1:6 nvme2n1:7 nvme3n1:8 nvme4n1:9 nvme5n1:10 nvme6n1:11]} ` +
		`dac3:{hostName:dac3 MGS: MDTS:map[nvme1n1:8 nvme2n1:9 nvme3n1:10 nvme4n1:11] ` +
		`OSTS:map[nvme1n1:12 nvme2n1:13 nvme3n1:14 nvme4n1:15 nvme5n1:16 nvme6n1:17]} ` +
		`dac4:{hostName:dac4 MGS: MDTS:map[nvme1n1:12 nvme2n1:13 nvme3n1:14 nvme4n1:15] ` +
		`OSTS:map[nvme1n1:18 nvme2n1:19 nvme3n1:20 nvme4n1:21 nvme5n1:22 nvme6n1:23]} ` +
		`dac5:{hostName:dac5 MGS: MDTS:map[nvme1n1:16 nvme2n1:17 nvme3n1:18 nvme4n1:19] ` +
		`OSTS:map[nvme1n1:24 nvme2n1:25 nvme3n1:26 nvme4n1:27 nvme5n1:28 nvme6n1:29]} `+
		`slurmmaster1:{hostName:slurmmaster1 MGS:loop0 MDTS:map[] OSTS:map[]}]`
	assert.Equal(t, expected2, resultStr2)
}
