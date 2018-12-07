package ansible

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeRunner struct {
	err error
	calls int
	hostnames []string
	cmdStrs []string
}

func (f *fakeRunner) Execute(hostname string, cmdStr string) error {
	f.calls += 1
	if f.err != nil {
		return f.err
	}
	f.hostnames = append(f.hostnames, hostname)
	f.cmdStrs = append(f.cmdStrs, cmdStr)
	return nil
}

func Test_mkdir(t *testing.T) {
	defer func() {runner = &run{}}()
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
	defer func() {runner = &run{}}()
	fake := &fakeRunner{}
	runner = fake

	err := mountLustre("host", "mgt", "fs", "dir")
	assert.Nil(t, err)
	assert.Equal(t, "host", fake.hostnames[0])
	assert.Equal(t, "host", fake.hostnames[1])
	assert.Equal(t, "modprobe -v lustre", fake.cmdStrs[0])
	assert.Equal(t, "mount -t lustre mgt:/fs dir", fake.cmdStrs[1])

	runner = &fakeRunner{err: errors.New("expected")}
	err = mountLustre("", "", "", "")
	assert.Equal(t, "expected", err.Error())
}
