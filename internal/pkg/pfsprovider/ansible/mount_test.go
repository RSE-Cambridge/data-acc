package ansible

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeRunner struct {
	t *testing.T
	hostname string
	cmdStr string
	err error
}

func (f *fakeRunner) Execute(hostname string, cmdStr string) error {
	if f.err != nil {
		return f.err
	}
	assert.Equal(f.t, f.hostname, hostname)
	assert.Equal(f.t, f.cmdStr, cmdStr)
	return nil
}

func Test_mkdir(t *testing.T) {
	defer func() {runner = &run{}}()
	runner = &fakeRunner{t: t, hostname:"host", cmdStr:"mkdir -p dir"}
	err := mkdir("host", "dir")
	assert.Nil(t, err)

	runner = &fakeRunner{err: errors.New("expected")}
	err = mkdir("", "")
	assert.Equal(t, "expected", err.Error())
}
