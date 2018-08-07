package fake

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"fmt"
	"log"
	"path"
)

func getMountDir(volume registry.Volume) string {
	if volume.MultiJob {
		return fmt.Sprintf("/mnt/multi_job_buffer/%s", volume.UUID)
	}
	return fmt.Sprintf("/mnt/job_buffer/%s", volume.UUID)
}

func mount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("FAKE Mount for:", volume.Name)

	var primaryBrickHost string
	for _, allocation := range brickAllocations {
		if allocation.AllocatedIndex == 0 {
			primaryBrickHost = allocation.Hostname
			break
		}
	}

	if primaryBrickHost == "" {
		log.Panicf("failed to find primary brick for volume: %s", volume.Name)
	}

	var mountDir = getMountDir(volume)
	for _, attachment := range volume.Attachments {
		log.Printf("FAKE ssh %s mount -t lustre %s:/%s %s",
			attachment.Hostname, primaryBrickHost, volume.UUID, mountDir)
		// TODO: what about the environment variables that are being set? should share logic with here

		if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
			swapDir := path.Join(mountDir, fmt.Sprintf("/swap/%s", attachment.Hostname))
			log.Printf("FAKE dd if=/dev/zero of=%s bs=1024 count=%d && chmod 0600 %s && mkswap %s",
				swapDir, int(volume.AttachAsSwapBytes/1024), swapDir, swapDir)
			log.Printf("FAKE swapon %s", swapDir)
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

func umount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("FAKE Umount for:", volume.Name)
	var mountDir = getMountDir(volume)
	for _, attachment := range volume.Attachments {
		if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
			swapDir := path.Join(mountDir, fmt.Sprintf("/swap/%s", attachment.Hostname))
			log.Printf("FAKE swapoff %s", swapDir)
		}
		log.Printf("FAKE ssh %s umount %s", attachment.Hostname, mountDir)
	}
	return nil
}