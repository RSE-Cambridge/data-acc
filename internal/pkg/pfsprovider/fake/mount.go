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
			swapDir := path.Join(mountDir, "/swap")
			if err := mkdir(attachment.Hostname, swapDir); err != nil {
				return err
			}

			// TODO: swapmb := int(volume.AttachAsSwapBytes/1024)
			swapmb := 2
			swapFile := path.Join(swapDir, fmt.Sprintf("/%s", attachment.Hostname))
			if err := createSwap(attachment.Hostname, swapmb, swapFile); err != nil {
				return err
			}

			if err := swapOn(attachment.Hostname, swapFile); err != nil {
				return err
			}
		}

		if !volume.MultiJob && volume.AttachPrivateNamespace {
			privateDir := path.Join(mountDir, fmt.Sprintf("/private/%s", attachment.Hostname))
			if err := mkdir(attachment.Hostname, privateDir); err != nil {
				return err
			}
			chown(attachment.Hostname, volume.Owner, privateDir)
		}

		sharedDir := path.Join(mountDir, "/shared")
		if err := mkdir(attachment.Hostname, sharedDir); err != nil {
			return err
		}
		chown(attachment.Hostname, volume.Owner, sharedDir)
	}
	// TODO on error should we always call umount? maybe?
	// TODO move to ansible style automation or preamble?
	return nil
}

func umount(volume registry.Volume, brickAllocations []registry.BrickAllocation) error {
	log.Println("FAKE Umount for:", volume.Name)
	var mountDir = getMountDir(volume)
	for _, attachment := range volume.Attachments {
		if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
			swapFile := path.Join(mountDir, fmt.Sprintf("/swap/%s", attachment.Hostname)) // TODO share?
			if err := swapOff(attachment.Hostname, swapFile); err != nil {
				return err
			}
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

func createSwap(hostname string, swapMb int, filename string) error {
	cmd := fmt.Sprintf("dd if=/dev/zero of=%s bs=1024 count=%d && sudo chmod 0600 %s && sudo mkswap %s",
		filename, swapMb, filename, filename)
	return remoteExecuteCmd(hostname, cmd)
}

func swapOn(hostname string, filename string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("swapon %s", filename))
}

func swapOff(hostname string, filename string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("swapoff %s", filename))
}

func chown(hostname string, owner uint, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("chown %d %s", owner, directory))
}

func umountLustre(hostname string, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("umount -l %s", directory))
}

func removeSubtree(hostname string, directory string) error {
	return remoteExecuteCmd(hostname, fmt.Sprintf("rm -rf %s", directory))
}

func mountLustre(hostname string, mgtHost string, fsname string, directory string) error {
	if err := remoteExecuteCmd(hostname, "modprobe -v lustre"); err != nil {
		return err
	}
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
