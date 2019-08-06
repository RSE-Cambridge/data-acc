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
	log.Println("SetupVolume for:", volume.Name)
	return nil
}

func (*volumeProvider) TeardownVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("TeardownVolume for:", volume.Name)
	return nil
}

func (*volumeProvider) CopyDataIn(volume registry.Volume) error {
	log.Println("CopyDataIn for:", volume.Name)
	return nil
}

func (*volumeProvider) CopyDataOut(volume registry.Volume) error {
	log.Println("CopyDataOut for:", volume.Name)
	return nil
}

type mounter struct{}

func (*mounter) Mount(volume registry.Volume, brickAllocations []registry.BrickAllocation, attachments []registry.Attachment) error {
	log.Println("Mount for:", volume.Name)
	return nil
}

func (*mounter) Unmount(volume registry.Volume, brickAllocations []registry.BrickAllocation, attachments []registry.Attachment) error {
	log.Println("Umount for:", volume.Name)
	return nil
}
