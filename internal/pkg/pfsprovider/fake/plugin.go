package fake

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
)

func GetPlugin() pfsprovider.Plugin {
	return &plugin{}
}

type plugin struct{}

func (*plugin) Mounter() pfsprovider.Mounter {
	return &mounter{}
}

func (*plugin) VolumeProvider() pfsprovider.VolumeProvider {
	return &volumeProvider{}
}

type volumeProvider struct{}

func (*volumeProvider) SetupVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("FAKE SetupVolume for:", volume.Name)
	log.Println(volume)
	log.Println(printLustreInfo(volume, brickAllocations))
	log.Println(printLustrePlaybook(volume))
	return nil
}

func (*volumeProvider) TeardownVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("FAKE TeardownVolume for:", volume.Name)
	log.Println(volume)
	log.Println(printLustreInfo(volume, brickAllocations))
	log.Println(printLustrePlaybook(volume))
	return nil
}

func (*volumeProvider) CopyDataIn(volume registry.Volume) error {
	log.Println("FAKE CopyDataIn for:", volume.Name)
	log.Println(volume.StageIn)
	return nil
}

func (*volumeProvider) CopyDataOut(volume registry.Volume) error {
	log.Println("FAKE CopyDataOut for:", volume.Name)
	log.Println(volume.StageOut)
	return nil
}

type mounter struct{}

func (*mounter) Mount(volume registry.Volume) error {
	log.Println("FAKE Mount for:", volume.Name, "with attachments:", volume.Attachments)
	return nil
}

func (*mounter) Unmount(volume registry.Volume) error {
	log.Println("FAKE Unmount for:", volume.Name, "with attachments:", volume.Attachments)
	return nil
}
