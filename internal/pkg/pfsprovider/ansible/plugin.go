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

func (volProvider *volumeProvider) SetupVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	return executeTempAnsible(volProvider.FSType, volume, brickAllocations, false)
}

func (volProvider *volumeProvider) TeardownVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	return executeTempAnsible(volProvider.FSType, volume, brickAllocations, true)
}

func (*volumeProvider) CopyDataIn(volume registry.Volume) error {
	// TODO we should support multiple stagein commands! oops!
	//return processDataCopy(volume.Name, volume.StageIn)
	log.Println("FAKE CopyIn:", volume.Name, volume.StageIn)
	return nil
}

func (*volumeProvider) CopyDataOut(volume registry.Volume) error {
	//return processDataCopy(volume.Name, volume.StageOut)
	log.Println("FAKE CopyOut:", volume.Name, volume.StageOut)
	return nil
}

type mounter struct {
	FSType FSType
}

func (*mounter) Mount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	// return mount(volume, brickAllocations)
	log.Println("FAKE Mount:", volume.Name, volume.Attachments)
	return nil
}

func (*mounter) Unmount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	// return umount(volume, brickAllocations)
	log.Println("FAKE Unmount:", volume.Name, volume.Attachments)
	return nil
}
