package fake

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"os/exec"
	"path"
)

func getMountDir(volume registry.Volume) string {
	// TODO: what about the environment variables that are being set? should share logic with here
	if volume.MultiJob {
		return fmt.Sprintf("/mnt/multi_job_buffer/%s", volume.UUID)
	}
	return fmt.Sprintf("/mnt/job_buffer/%s", volume.UUID)
}

func mount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
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

		if err := mkdir(attachment.Hostname, mountDir); err != nil {
			return err
		}
		if err := mountLustre(attachment.Hostname, primaryBrickHost, volume.UUID, mountDir); err != nil {
			return err
		}

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
		if err := umountLustre(attachment.Hostname, mountDir); err != nil {
			return err
		}
		if err := removeSubtree(attachment.Hostname, mountDir); err != nil {
			return err
		}
	}
	return nil
}

func umountLustre(hostname string, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("umount -l %s", directory))
}

func removeSubtree(hostname string, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("rm -rf %s", directory))
}

func mountLustre(hostname string, mgtHost string, fsname string, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf(
		"mount -t lustre %s:/%s %s", mgtHost, fsname, directory))
}

func mkdir(hostname string, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("mkdir -p %s", directory))
}

func remoteExecuteCmd(hostname string, cmdStr string) error {
	log.Println("SSH to:", hostname, "with command:", cmdStr)

	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", hostname, "sudo", cmdStr)
	output, err := cmd.CombinedOutput()

	if err == nil {
		log.Println("Completed remote ssh run:", cmdStr)
		log.Println(string(output))
		return nil
	} else {
		log.Println("Error in remove ssh run:", string(output))
		return err
	}
}
