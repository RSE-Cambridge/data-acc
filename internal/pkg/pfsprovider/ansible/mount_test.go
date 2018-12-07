package ansible

import (
	"errors"
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

func Test_chown(t *testing.T) {
	defer func() { runner = &run{} }()
	fake := &fakeRunner{err: errors.New("expected")}
	runner = fake

	err := chown("host", 10, 11, "dir")
	assert.Equal(t, "expected", err.Error())
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "chown 10:11 dir", fake.cmdStrs[0])
}
