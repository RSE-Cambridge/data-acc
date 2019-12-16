package filesystem_impl

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeRunner struct {
	err       error
	calls     int
	hostnames []string
	cmdStrs   []string
}

func (f *fakeRunner) Execute(hostname string, asRoot bool, cmdStr string) error {
	f.calls += 1
	f.hostnames = append(f.hostnames, hostname)
	f.cmdStrs = append(f.cmdStrs, cmdStr)
	if cmdStr == "grep /mnt/dac/job1_job /etc/mtab" {
		return errors.New("trigger mount")
	}
	if cmdStr == "grep /mnt/dac/job1_persistent_asdf /etc/mtab" {
		return errors.New("trigger mount")
	}
	return f.err
}

func Test_mkdir(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	err := mkdir("host", "dir")
	assert.Nil(t, err)
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "mkdir -p dir", fake.cmdStrs[0])

	runner = &fakeRunner{err: errors.New("expected")}
	err = mkdir("", "")
	assert.Equal(t, "expected", err.Error())
}

func Test_mountLustre(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	err := mountLustre("host", "-opa@o2ib1", "mgt", "fs", "/mnt/dac/job1_job")
	assert.Nil(t, err)
	assert.Equal(t, 2, fake.calls)
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "host", fake.hostnames[1])
	assert.Equal(t, "grep /mnt/dac/job1_job /etc/mtab", fake.cmdStrs[0])
	assert.Equal(t, "mount -t lustre -o flock,nodev,nosuid mgt-opa@o2ib1:/fs /mnt/dac/job1_job", fake.cmdStrs[1])

	fake = &fakeRunner{err: errors.New("expected")}
	runner = fake
	err = mountRemoteFilesystem(Lustre, "host", "", "mgt", "fs", "asdf")
	assert.Equal(t, "expected", err.Error())
	assert.Equal(t, 2, fake.calls)
	assert.Equal(t, "grep asdf /etc/mtab", fake.cmdStrs[0])
	assert.Equal(t, "mount -t lustre -o flock,nodev,nosuid mgt:/fs asdf", fake.cmdStrs[1])
}

func Test_createSwap(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	err := createSwap("host", 3, "file", "loopback")
	assert.Nil(t, err)
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "host", fake.hostnames[1])
	assert.Equal(t, "host", fake.hostnames[2])
	assert.Equal(t, 4, len(fake.cmdStrs))
	assert.Equal(t, "dd if=/dev/zero of=file bs=1024 count=3072", fake.cmdStrs[0])
	assert.Equal(t, "chmod 0600 file", fake.cmdStrs[1])
	assert.Equal(t, "losetup loopback file", fake.cmdStrs[2])
	assert.Equal(t, "mkswap loopback", fake.cmdStrs[3])
}

func Test_fixUpOwnership(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	err := fixUpOwnership("host", 10, 11, "dir")
	assert.Nil(t, err)

	assert.Equal(t, 2, fake.calls)
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "chown 10:11 dir", fake.cmdStrs[0])
	assert.Equal(t, "host", fake.hostnames[1])
	assert.Equal(t, "chmod 700 dir", fake.cmdStrs[1])
}

func Test_Mount(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	sessionName := datamodel.SessionName("job1")
	internalName := "fsuuid"
	primaryBrickHost := datamodel.BrickHostName("host1")
	owner := uint(1001)
	group := uint(1002)
	attachment := datamodel.AttachmentSession{

		SessionName: "job2", // changed to prove this is not used
		Hosts:       []string{"client1", "client2"},

		GlobalMount:  true,
		PrivateMount: true,
		SwapBytes:    1024 * 1024, // 1 MiB
	}
	err := mount(Lustre, sessionName, false,
		internalName, primaryBrickHost, attachment,
		owner, group, true)
	assert.Nil(t, err)
	assert.Equal(t, 17, fake.calls)

	assert.Equal(t, "client1", fake.hostnames[0])
	assert.Equal(t, "mkdir -p /mnt/dac/job1_job", fake.cmdStrs[0])
	assert.Equal(t, "grep /mnt/dac/job1_job /etc/mtab", fake.cmdStrs[1])
	assert.Equal(t, "mount -t lustre -o flock,nodev,nosuid host1:/fsuuid /mnt/dac/job1_job", fake.cmdStrs[2])

	assert.Equal(t, "client2", fake.hostnames[3])
	assert.Equal(t, "mkdir -p /mnt/dac/job1_job", fake.cmdStrs[3])
	assert.Equal(t, "grep /mnt/dac/job1_job /etc/mtab", fake.cmdStrs[4])
	assert.Equal(t, "mount -t lustre -o flock,nodev,nosuid host1:/fsuuid /mnt/dac/job1_job", fake.cmdStrs[5])

	assert.Equal(t, "client1", fake.hostnames[6])
	assert.Equal(t, "mkdir -p /mnt/dac/job1_job/global", fake.cmdStrs[6])
	assert.Equal(t, "chown 1001:1002 /mnt/dac/job1_job/global", fake.cmdStrs[7])
	assert.Equal(t, "chmod 700 /mnt/dac/job1_job/global", fake.cmdStrs[8])

	assert.Equal(t, "client1", fake.hostnames[9])
	assert.Equal(t, "mkdir -p /mnt/dac/job1_job/private/client1", fake.cmdStrs[9])
	assert.Equal(t, "chown 1001:1002 /mnt/dac/job1_job/private/client1", fake.cmdStrs[10])
	assert.Equal(t, "chmod 700 /mnt/dac/job1_job/private/client1", fake.cmdStrs[11])
	assert.Equal(t, "ln -s /mnt/dac/job1_job/private/client1 /mnt/dac/job1_job_private", fake.cmdStrs[12])

	assert.Equal(t, "client2", fake.hostnames[13])
	assert.Equal(t, "mkdir -p /mnt/dac/job1_job/private/client2", fake.cmdStrs[13])

	assert.Equal(t, "client2", fake.hostnames[16])
	assert.Equal(t, "ln -s /mnt/dac/job1_job/private/client2 /mnt/dac/job1_job_private", fake.cmdStrs[16])
}

func Test_Umount(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	sessionName := datamodel.SessionName("job4")
	internalName := "fsuuid"
	primaryBrickHost := datamodel.BrickHostName("host1")
	attachment := datamodel.AttachmentSession{

		SessionName: "job2",
		Hosts:       []string{"client1", "client2"},

		GlobalMount:  true,
		PrivateMount: true,
		SwapBytes:    1024 * 1024, // 1 MiB
	}
	err := unmount(Lustre, sessionName, false,
		internalName, primaryBrickHost, attachment)
	assert.Nil(t, err)
	assert.Equal(t, 8, fake.calls)

	assert.Equal(t, "client1", fake.hostnames[0])
	assert.Equal(t, "rm -df /mnt/dac/job4_job_private", fake.cmdStrs[0])
	assert.Equal(t, "grep /mnt/dac/job4_job /etc/mtab", fake.cmdStrs[1])
	assert.Equal(t, "umount /mnt/dac/job4_job", fake.cmdStrs[2])
	assert.Equal(t, "rm -df /mnt/dac/job4_job", fake.cmdStrs[3])

	assert.Equal(t, "client2", fake.hostnames[7])
	assert.Equal(t, "rm -df /mnt/dac/job4_job", fake.cmdStrs[7])
}

func Test_Umount_multi(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	sessionName := datamodel.SessionName("asdf2")
	internalName := "uuidasdf"
	primaryBrickHost := datamodel.BrickHostName("host1")
	attachment := datamodel.AttachmentSession{
		SessionName:  "job1",
		Hosts:        []string{"client1"},
		GlobalMount:  true,
		PrivateMount: false,
		SwapBytes:    0,
	}
	err := unmount(Lustre, sessionName, true,
		internalName, primaryBrickHost, attachment)

	assert.Nil(t, err)
	assert.Equal(t, 3, fake.calls)

	assert.Equal(t, "client1", fake.hostnames[0])
	assert.Equal(t, "grep /mnt/dac/job1_persistent_asdf2 /etc/mtab", fake.cmdStrs[0])
	assert.Equal(t, "umount /mnt/dac/job1_persistent_asdf2", fake.cmdStrs[1])
	assert.Equal(t, "rm -df /mnt/dac/job1_persistent_asdf2", fake.cmdStrs[2])
}

func Test_Mount_multi(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake

	sessionName := datamodel.SessionName("asdf")
	internalName := "uuidasdf"
	primaryBrickHost := datamodel.BrickHostName("host1")
	owner := uint(1001)
	group := uint(1002)
	attachment := datamodel.AttachmentSession{
		SessionName:  "job1",
		Hosts:        []string{"client1"},
		GlobalMount:  true,
		PrivateMount: false,
		SwapBytes:    0,
	}
	err := mount(Lustre, sessionName, true,
		internalName, primaryBrickHost, attachment,
		owner, group, false)

	assert.Nil(t, err)
	assert.Equal(t, 3, fake.calls)

	assert.Equal(t, "client1", fake.hostnames[0])
	assert.Equal(t, "mkdir -p /mnt/dac/job1_persistent_asdf", fake.cmdStrs[0])
	assert.Equal(t, "client1", fake.hostnames[1])
	assert.Equal(t, "grep /mnt/dac/job1_persistent_asdf /etc/mtab", fake.cmdStrs[1])
	assert.Equal(t, "client1", fake.hostnames[2])
	assert.Equal(t, "mount -t lustre -o flock,nodev,nosuid host1:/uuidasdf /mnt/dac/job1_persistent_asdf", fake.cmdStrs[2])
}
