package ansible

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeRunner struct {
	err       error
	calls     int
	hostnames []string
	cmdStrs   []string
}

func (f *fakeRunner) Execute(hostname string, cmdStr string) error {
	f.calls += 1
	f.hostnames = append(f.hostnames, hostname)
	f.cmdStrs = append(f.cmdStrs, cmdStr)
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

	err := mountLustre("host", "mgt", "fs", "dir")
	assert.Nil(t, err)
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "host", fake.hostnames[1])
	assert.Equal(t, "modprobe -v lustre", fake.cmdStrs[0])
	assert.Equal(t, "mount -t lustre mgt:/fs dir", fake.cmdStrs[1])

	fake = &fakeRunner{err: errors.New("expected")}
	runner = fake
	err = mountRemoteFilesystem(Lustre, "host", "", "", "")
	assert.Equal(t, "expected", err.Error())
	assert.Equal(t, "modprobe -v lustre", fake.cmdStrs[0])
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
	assert.Equal(t, "dd if=/dev/zero of=file bs=1024 count=3072 && sudo chmod 0600 file", fake.cmdStrs[0])
	assert.Equal(t, "losetup loopback file", fake.cmdStrs[1])
	assert.Equal(t, "mkswap loopback", fake.cmdStrs[2])
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
	assert.Equal(t, "chmod 770 dir", fake.cmdStrs[1])
}

func Test_Mount(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{}
	runner = fake
	volume := registry.Volume{
		Name: "asdf", JobName: "asdf",
		AttachGlobalNamespace:  true,
		AttachPrivateNamespace: true,
		AttachAsSwapBytes:      10000,
		Attachments: []registry.Attachment{
			{Hostname: "client1", Job: "job1", State: registry.RequestAttach},
			{Hostname: "client2", Job: "job1", State: registry.RequestAttach},
			{Hostname: "client3", Job: "job3", State: registry.Attached},
			{Hostname: "client3", Job: "job3", State: registry.RequestDetach},
			{Hostname: "client3", Job: "job3", State: registry.Detached},
			{Hostname: "client2", Job: "job2", State: registry.RequestAttach},
		},
		ClientPort: 42,
		Owner:      1001,
		Group:      1001,
	}

	assert.PanicsWithValue(t,
		"failed to find primary brick for volume: asdf",
		func() { mount(Lustre, volume, nil) })

	bricks := []registry.BrickAllocation{
		{Hostname: "host1"},
		{Hostname: "host2"},
	}
	err := mount(Lustre, volume, bricks)
	assert.Nil(t, err)
	assert.Equal(t, 51, fake.calls)

	assert.Equal(t, "client1", fake.hostnames[0])
	assert.Equal(t, "mkdir -p /dac/job1/job", fake.cmdStrs[0])
	assert.Equal(t, "modprobe -v lustre", fake.cmdStrs[1])
	assert.Equal(t, "mount -t lustre host1:/ /dac/job1/job", fake.cmdStrs[2])

	assert.Equal(t, "mkdir -p /dac/job1/job/swap", fake.cmdStrs[3])
	assert.Equal(t, "chown 0:0 /dac/job1/job/swap", fake.cmdStrs[4])
	assert.Equal(t, "chmod 770 /dac/job1/job/swap", fake.cmdStrs[5])
	assert.Equal(t, "dd if=/dev/zero of=/dac/job1/job/swap/client1 bs=1024 count=2048 && sudo chmod 0600 /dac/job1/job/swap/client1", fake.cmdStrs[6])
	assert.Equal(t, "losetup /dev/loop42 /dac/job1/job/swap/client1", fake.cmdStrs[7])
	assert.Equal(t, "mkswap /dev/loop42", fake.cmdStrs[8])
	assert.Equal(t, "swapon /dev/loop42", fake.cmdStrs[9])
	assert.Equal(t, "mkdir -p /dac/job1/job/private/client1", fake.cmdStrs[10])
	assert.Equal(t, "chown 1001:1001 /dac/job1/job/private/client1", fake.cmdStrs[11])
	assert.Equal(t, "chmod 770 /dac/job1/job/private/client1", fake.cmdStrs[12])
	assert.Equal(t, "ln -s /dac/job1/job/private/client1 /dac/asdf/job_private", fake.cmdStrs[13])

	assert.Equal(t, "mkdir -p /dac/job1/job/global", fake.cmdStrs[14])
	assert.Equal(t, "chown 1001:1001 /dac/job1/job/global", fake.cmdStrs[15])
	assert.Equal(t, "chmod 770 /dac/job1/job/global", fake.cmdStrs[16])

	assert.Equal(t, "client2", fake.hostnames[17])
	assert.Equal(t, "mkdir -p /dac/job1/job", fake.cmdStrs[17])

	assert.Equal(t, "client2", fake.hostnames[34])
	assert.Equal(t, "mkdir -p /dac/job2/job", fake.cmdStrs[34])
	assert.Equal(t, "client2", fake.hostnames[50])
	assert.Equal(t, "chmod 770 /dac/job2/job/global", fake.cmdStrs[50])
}
