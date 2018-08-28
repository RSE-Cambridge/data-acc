package ansible

import (
	"bytes"
	"encoding/json"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func GetPlugin(fsType FSType) pfsprovider.Plugin {
	return &plugin{FSType: fsType}
}

type FSType int

const (
	BeegFS FSType = iota
	Lustre
)

var fsTypeStrings = map[FSType]string{
	BeegFS: "BeegFS",
	Lustre: "Lustre",
}
var stringToFSType = map[string]FSType{
	"":       BeegFS,
	"BeegFS": BeegFS,
	"Lustre": Lustre,
}

func (fsType FSType) String() string {
	return fsTypeStrings[fsType]
}

func (fsType FSType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(fsTypeStrings[fsType])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (fsType *FSType) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	*fsType = stringToFSType[str]
	return nil
}

type plugin struct {
	FSType FSType
}

func (plugin *plugin) Mounter() pfsprovider.Mounter {
	return &mounter{FSType: plugin.FSType}
}

func (plugin *plugin) VolumeProvider() pfsprovider.VolumeProvider {
	return &volumeProvider{FSType: plugin.FSType}
}

type volumeProvider struct {
	FSType FSType
}

func (*volumeProvider) SetupVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("SetupVolume for:", volume.Name)
	return executeTempAnsible(volume, brickAllocations, false)
}

func (*volumeProvider) TeardownVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("TeardownVolume for:", volume.Name)
	return executeTempAnsible(volume, brickAllocations, true)
}

func (*volumeProvider) CopyDataIn(volume registry.Volume) error {
	// TODO we should support multiple stagein commands! oops!
	log.Println("CopyDataIn for:", volume.Name)
	return processDataCopy(volume.Name, volume.StageIn)
}

func (*volumeProvider) CopyDataOut(volume registry.Volume) error {
	log.Println("CopyDataOut for:", volume.Name)
	return processDataCopy(volume.Name, volume.StageOut)
}

type mounter struct {
	FSType FSType
}

func (*mounter) Mount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("Mount for:", volume.Name)
	return mount(volume, brickAllocations)
}

func (*mounter) Unmount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("Umount for:", volume.Name)
	return umount(volume, brickAllocations)
}
