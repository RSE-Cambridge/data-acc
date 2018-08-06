package fake

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"fmt"
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
	return executeTempAnsible(volume, brickAllocations, false)
}

func (*volumeProvider) TeardownVolume(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("FAKE TeardownVolume for:", volume.Name)
	return executeTempAnsible(volume, brickAllocations, true)
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
	log.Println("FAKE Mount for:", volume.Name)
	var mountDir string
	if volume.MultiJob {
		mountDir = fmt.Sprintf("/mnt/multi_job_buffer/%s", volume.UUID)
	} else {
		mountDir = fmt.Sprintf("/mnt/job_buffer/%s", volume.UUID)
	}
	for _, attachment := range volume.Attachments {
		// TODO: need brick directory
		log.Printf("Fake ssh %s mount -t lustre TODO:/%s %s", attachment.Hostname, volume.UUID, mountDir)
		log.Printf("fake update permissions, owner: %s", volume.Owner)
		// TODO: create swap file per compute? etc?
	}
	return nil
}

func (*mounter) Unmount(volume registry.Volume) error {
	log.Println("FAKE Mount for:", volume.Name)
	mountDir := fmt.Sprintf("/mnt/lustre/%s", volume.Name)
	for _, attachment := range volume.Attachments {
		log.Printf("Fake ssh %s umount %s", attachment.Hostname, mountDir)
	}
	return nil
}
