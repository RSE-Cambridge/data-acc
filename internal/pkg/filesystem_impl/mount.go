package filesystem_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"log"
	"os/exec"
	"path"
	"time"
)

func getMountDir(sourceName datamodel.SessionName, isMultiJob bool, attachingForSession datamodel.SessionName) string {
	// TODO: what about the environment variables that are being set? should share logic with here
	if isMultiJob {
		return fmt.Sprintf("/dac/%s_persistent_%s", attachingForSession, sourceName)
	}
	return fmt.Sprintf("/dac/%s_job", sourceName)
}

func mount(fsType FSType, sessionName datamodel.SessionName, isMultiJob bool, internalName string,
	primaryBrickHost datamodel.BrickHostName, attachment datamodel.AttachmentSession,
	owner uint, group uint, setInitialPermissions bool) error {
	log.Println("Mount for:", sessionName)

	if primaryBrickHost == "" {
		log.Panicf("failed to find primary brick for volume: %s", sessionName)
	}
	if len(attachment.Hosts) < 1 {
		log.Println("Skip mount as no hosts given")
		return nil
	}
	if fsType == BeegFS {
		// Write out the config needed, and do the mount using ansible
		// TODO: Move Lustre mount here that is done below
		//executeAnsibleMount(fsType, volume, brickAllocations)
	}
	var mountDir = getMountDir(sessionName, isMultiJob, attachment.SessionName)

	for _, attachHost := range attachment.Hosts {
		log.Printf("Mounting %s on host: %s for session: %s", sessionName, attachHost,
			attachment.SessionName)

		if err := mkdir(attachHost, mountDir); err != nil {
			return err
		}
		if err := mountRemoteFilesystem(fsType, attachHost, conf.LnetSuffix,
			string(primaryBrickHost), internalName, mountDir); err != nil {
			return err
		}
	}

	// Create global and private directories, with correct permissions
	if setInitialPermissions {
		// use first mounted host
		attachHost := attachment.Hosts[0]

		// make a directory users can write into
		sharedDir := path.Join(mountDir, "/global")
		// TODO: would install be better here?
		if err := mkdir(attachHost, sharedDir); err != nil {
			return err
		}
		if err := fixUpOwnership(attachHost, owner, group, sharedDir); err != nil {
			return err
		}

		if !isMultiJob {
			// base private dir similar to global dir
			privateDir := path.Join(mountDir, "/private")
			// TODO would install be better here?
			if err := mkdir(attachHost, privateDir); err != nil {
				return err
			}
			if err := fixUpOwnership(attachHost, owner, group, privateDir); err != nil {
				return err
			}

			// Swap is owned by root
			swapDir := path.Join(mountDir, "/swap")
			if err := mkdir(attachHost, swapDir); err != nil {
				return err
			}
			if err := fixUpOwnership(attachHost, owner, group, privateDir); err != nil {
				return err
			}
		}
	}

	// add sym link to a private directory as needed
	if attachment.PrivateMount {
		for _, attachHost := range attachment.Hosts {
			privateDir := path.Join(mountDir, fmt.Sprintf("/private/%s", attachHost))
			if err := mkdir(attachHost, privateDir); err != nil {
				return err
			}
			if err := fixUpOwnership(attachHost, owner, group, privateDir); err != nil {
				return err
			}
			// need a consistent symlink for shared environment variables across all hosts
			privateSymLinkDir := fmt.Sprintf("/dac/%s_job_private", sessionName)
			if err := createSymbolicLink(attachHost, privateDir, privateSymLinkDir); err != nil {
				return err
			}
		}
	}

	// TODO: swap!
	//if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
	//	swapDir := path.Join(mountDir, "/swap")
	//	if err := mkdir(attachment.Hostname, swapDir); err != nil {
	//		return err
	//	}
	//	if err := fixUpOwnership(attachment.Hostname, 0, 0, swapDir); err != nil {
	//		return err
	//	}
	//	swapSizeMB := int(volume.AttachAsSwapBytes / (1024 * 1024))
	//	swapFile := path.Join(swapDir, fmt.Sprintf("/%s", attachment.Hostname))
	//	loopback := fmt.Sprintf("/dev/loop%d", volume.ClientPort)
	//	if err := createSwap(attachment.Hostname, swapSizeMB, swapFile, loopback); err != nil {
	//		return err
	//	}
	//	if err := swapOn(attachment.Hostname, loopback); err != nil {
	//		return err
	//	}
	//}
	// TODO on error should we always call umount? maybe?
	// TODO move to ansible style automation or preamble?
	return nil
}

func unmount(fsType FSType, sessionName datamodel.SessionName, isMultiJob bool, internalName string,
	primaryBrickHost datamodel.BrickHostName, attachment datamodel.AttachmentSession) error {
	log.Println("Umount for:", sessionName)

	for _, attachHost := range attachment.Hosts {
		log.Printf("Unmounting %s on host: %s for session: %s", sessionName, attachHost,
			attachment.SessionName)

		var mountDir = getMountDir(sessionName, isMultiJob, attachment.SessionName)
		// TODO: swap!
		//if !volume.MultiJob && volume.AttachAsSwapBytes > 0 {
		//	swapFile := path.Join(mountDir, fmt.Sprintf("/swap/%s", attachment.Hostname)) // TODO share?
		//	loopback := fmt.Sprintf("/dev/loop%d", volume.ClientPort)                     // TODO share?
		//	if err := swapOff(attachment.Hostname, loopback); err != nil {
		//		log.Printf("Warn: failed to swap off %+v", attachment)
		//	}
		//	if err := detachLoopback(attachment.Hostname, loopback); err != nil {
		//		log.Printf("Warn: failed to detach loopback %+v", attachment)
		//	}
		//	if err := removeSubtree(attachment.Hostname, swapFile); err != nil {
		//		return err
		//	}
		//}
		if attachment.PrivateMount {
			privateSymLinkDir := fmt.Sprintf("/dac/%s_job_private", sessionName)
			if err := removeSubtree(attachHost, privateSymLinkDir); err != nil {
				return err
			}
		}

		if fsType == Lustre {
			if err := umountLustre(attachHost, mountDir); err != nil {
				return err
			}
			if err := removeSubtree(attachHost, mountDir); err != nil {
				return err
			}
		}
	}

	if fsType == BeegFS {
		// TODO: Move Lustre unmount here that is done below
		// executeAnsibleUnmount(fsType, volume, brickAllocations)
		// TODO: this makes copy out much harder in its current form :(
	}
	return nil
}

func createSwap(hostname string, swapMB int, filename string, loopback string) error {
	file := fmt.Sprintf("dd if=/dev/zero of=%s bs=1024 count=%d", filename, swapMB*1024)
	if err := runner.Execute(hostname, file); err != nil {
		return err
	}
	if err := runner.Execute(hostname, fmt.Sprintf("chmod 0600 %s", filename)); err != nil {
		return err
	}
	device := fmt.Sprintf("losetup %s %s", loopback, filename)
	if err := runner.Execute(hostname, device); err != nil {
		return err
	}
	swap := fmt.Sprintf("mkswap %s", loopback)
	return runner.Execute(hostname, swap)
}

func swapOn(hostname string, loopback string) error {
	return runner.Execute(hostname, fmt.Sprintf("swapon %s", loopback))
}

func swapOff(hostname string, loopback string) error {
	return runner.Execute(hostname, fmt.Sprintf("swapoff %s", loopback))
}

func detachLoopback(hostname string, loopback string) error {
	return runner.Execute(hostname, fmt.Sprintf("losetup -d %s", loopback))
}

func fixUpOwnership(hostname string, owner uint, group uint, directory string) error {
	if err := runner.Execute(hostname, fmt.Sprintf("chown %d:%d %s", owner, group, directory)); err != nil {
		return err
	}
	return runner.Execute(hostname, fmt.Sprintf("chmod 700 %s", directory))
}

func umountLustre(hostname string, directory string) error {
	// only unmount if already mounted
	if err := runner.Execute(hostname, fmt.Sprintf("grep %s /etc/mtab", directory)); err == nil {
		// Don't add -l so we can spot when this fails
		if err := runner.Execute(hostname, fmt.Sprintf("umount %s", directory)); err != nil {
			return err
		}
	} else {
		// TODO: we should really just avoid this being possible?
		log.Println("skip umount, as not currently mounted.")
	}
	return nil
}

func removeSubtree(hostname string, directory string) error {
	return runner.Execute(hostname, fmt.Sprintf("rm -df %s", directory))
}

func createSymbolicLink(hostname string, src string, dest string) error {
	return runner.Execute(hostname, fmt.Sprintf("ln -s %s %s", src, dest))
}

func mountRemoteFilesystem(fsType FSType, hostname string, lnetSuffix string, mgtHost string, fsname string, directory string) error {
	if fsType == Lustre {
		return mountLustre(hostname, lnetSuffix, mgtHost, fsname, directory)
	} else if fsType == BeegFS {
		return mountBeegFS(hostname, mgtHost, fsname, directory)
	}
	return fmt.Errorf("mount unsuported by filesystem type %s", fsType)
}

func mountLustre(hostname string, lnetSuffix string, mgtHost string, fsname string, directory string) error {
	// We assume modprobe -v lustre is already done
	// First check if we are mounted already
	if err := runner.Execute(hostname, fmt.Sprintf("grep %s /etc/mtab", directory)); err != nil {
		if err := runner.Execute(hostname, fmt.Sprintf(
			"mount -t lustre -o flock,nodev,nosuid %s%s:/%s %s",
			mgtHost, lnetSuffix, fsname, directory)); err != nil {
			return err
		}
	}
	return nil
}

func mountBeegFS(hostname string, mgtHost string, fsname string, directory string) error {
	// Ansible mounts beegfs at /mnt/beegfs/<fsname>, link into above location here
	// First remove the directory, then replace with a symbolic link
	if err := removeSubtree(hostname, directory); err != nil {
		return err
	}
	return runner.Execute(hostname, fmt.Sprintf("ln -s /mnt/beegfs/%s %s", fsname, directory))
}

func mkdir(hostname string, directory string) error {
	return runner.Execute(hostname, fmt.Sprintf("mkdir -p %s", directory))
}

type Run interface {
	Execute(name string, cmd string) error
}

type run struct {
}

// TODO: need some code sharing here!!!
func (*run) Execute(hostname string, cmdStr string) error {
	log.Println("SSH to:", hostname, "with command:", cmdStr)

	if conf.SkipAnsible {
		log.Println("Skip as DAC_SKIP_ANSIBLE=True")
		time.Sleep(time.Millisecond * 200)
		return nil
	}

	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", hostname, "sudo", cmdStr)

	timer := time.AfterFunc(time.Minute, func() {
		log.Println("Time up, waited more than 5 mins to complete.")
		if err := cmd.Process.Kill(); err != nil {
			log.Panicf("error trying to kill process: %s", err.Error())
		}
	})

	output, err := cmd.CombinedOutput()
	timer.Stop()

	if err == nil {
		log.Println("Completed remote ssh run:", cmdStr)
		log.Println(string(output))
		return nil
	} else {
		log.Println("Error in remove ssh run:", string(output))
		return err
	}
}

var runner Run = &run{}
