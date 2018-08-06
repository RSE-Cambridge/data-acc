package fake

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"path"
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
		// TODO: what about the environment variables that are being set? should share logic with here

		if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
			swapDir := path.Join(mountDir, fmt.Sprintf("/swap/%s", attachment.Hostname))
			log.Printf("dd if=/dev/zero of=%s bs=1024 count=%d && chmod 0600 %s && mkswap %s",
				swapDir, int(volume.AttachAsSwapBytes/1024), swapDir, swapDir)
			log.Printf("swapon %s", swapDir)
		}

		if !volume.MultiJob && volume.AttachPrivateNamespace {
			privateDir := path.Join(mountDir, fmt.Sprintf("/private/%s", attachment.Hostname))
			log.Printf("FAKE mkdir -p %s", privateDir)
			log.Printf("FAKE chown %d %s", volume.Owner, privateDir)
		}

		sharedDir := path.Join(mountDir, "/shared")
		log.Printf("FAKE mkdir -p %s", sharedDir)
		log.Printf("FAKE chown %d %s", volume.Owner, sharedDir)
	}
	return nil
}

func (*mounter) Unmount(volume registry.Volume) error {
	log.Println("FAKE Umount for:", volume.Name)
	mountDir := fmt.Sprintf("/mnt/lustre/%s", volume.Name) // TODO fix to match above in func
	for _, attachment := range volume.Attachments {
		if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
			swapDir := path.Join(mountDir, fmt.Sprintf("/swap/%s", attachment.Hostname))
			log.Printf("FAKE swapoff %s", swapDir)
		}
		log.Printf("FAKE ssh %s umount %s", attachment.Hostname, mountDir)
	}
	return nil
}
