package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
)

func notEqual(a, b []string) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil || b == nil {
		return true
	}
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if a[i] != b[i] {
			return true
		}
	}
	return false
}

func TestStripFunctionArg(t *testing.T) {
	if v := stripFunctionArg([]string{"asdf", "--function", "foo"}); notEqual([]string{"asdf", "foo"}, v) {
		t.Fatalf("Expected 'foo' in list but got %s", v)
	}

	if v := stripFunctionArg([]string{}); notEqual([]string{}, v) {
		t.Fatalf("Expected empty list but got %s", v)
	}
}

func TestCreatePersistentBuffer(t *testing.T) {
	testActions = &stubDacctlActions{}
	testKeystore = &stubKeystore{}
	defer func() {
		testActions = nil
		testKeystore = nil
	}()

	createPersistentArgs := strings.Split(
		"--function create_persistent -t p2 -c c -u 1 -g 1 -C dw:1GiB -a striped -T scratch", " ")
	err := runCli(createPersistentArgs)
	assert.Equal(t, "CreatePersistentBuffer p2", err.Error())

	createPersistentArgs = strings.Split(
		"--function create_persistent --token p1 --caller c --user 1 --groupid 1 --capacity dw:1GiB "+
			"--access striped --type scratch", " ")
	err = runCli(createPersistentArgs)
	assert.Equal(t, "CreatePersistentBuffer p1", err.Error())
}

func TestDeleteBuffer(t *testing.T) {
	testActions = &stubDacctlActions{}
	testKeystore = &stubKeystore{}
	defer func() {
		testActions = nil
		testKeystore = nil
	}()

	err := runCli([]string{"--function", "teardown", "--job", "a", "--token", "a"})
	assert.Equal(t, "DeleteBuffer a", err.Error())

	err = runCli([]string{"--function", "teardown", "--job", "b", "--token", "a2", "--hurry"})
	assert.Equal(t, "DeleteBuffer a2", err.Error())
}

func TestCreatePerJobBuffer(t *testing.T) {
	testActions = &stubDacctlActions{}
	testKeystore = &stubKeystore{}
	defer func() {
		testActions = nil
		testKeystore = nil
	}()

	setupArgs := strings.Split(
		"--function setup --token a --job b --caller c --user 1 --groupid 1 --capacity dw:1GiB --nodehostnamefile asdf", " ")
	err := runCli(setupArgs)
	assert.Equal(t, "CreatePerJobBuffer", err.Error())

}

func TestShow(t *testing.T) {
	testActions = &stubDacctlActions{}
	testKeystore = &stubKeystore{}
	defer func() {
		testActions = nil
		testKeystore = nil
	}()

	err := runCli([]string{"--function", "pools"})
	assert.Equal(t, "ListPools", err.Error())

	err = runCli([]string{"--function", "show_instances"})
	assert.Equal(t, "ShowInstances", err.Error())

	err = runCli([]string{"--function", "show_sessions"})
	assert.Equal(t, "ShowSessions", err.Error())

	err = runCli([]string{"--function", "show_configurations"})
	assert.Equal(t, "ShowConfigurations", err.Error())
}

func TestFlow(t *testing.T) {
	testActions = &stubDacctlActions{}
	testKeystore = &stubKeystore{}
	defer func() {
		testActions = nil
		testKeystore = nil
	}()

	err := runCli([]string{"--function", "job_process", "--job", "a"})
	assert.Equal(t, "ValidateJob", err.Error())

	err = runCli([]string{"--function", "real_size", "--token", "a"})
	assert.Equal(t, "RealSize", err.Error())

	err = runCli([]string{"--function", "data_in", "--token", "a", "--job", "b"})
	assert.Equal(t, "DataIn", err.Error())

	err = runCli([]string{"--function", "paths", "--token", "a", "--job", "b", "--pathfile", "c"})
	assert.Equal(t, "Paths", err.Error())

	err = runCli([]string{"--function", "pre_run", "--token", "a", "--job", "b", "--nodehostnamefile", "c"})
	assert.Equal(t, "PreRun", err.Error())

	err = runCli([]string{"--function", "post_run", "--token", "a", "--job", "b"})
	assert.Equal(t, "PostRun", err.Error())

	err = runCli([]string{"--function", "data_out", "--token", "a", "--job", "b"})
	assert.Equal(t, "DataOut", err.Error())
}

type stubKeystore struct{}

func (*stubKeystore) Close() error {
	log.Println("closed")
	return nil
}

func (*stubKeystore) Create(key string, value []byte) (int64, error) {
	panic("implement me")
}

func (*stubKeystore) Update(key string, value []byte, modRevision int64) (int64, error) {
	panic("implement me")
}

func (*stubKeystore) Delete(key string, modRevision int64) error {
	panic("implement me")
}

func (*stubKeystore) DeleteAllKeysWithPrefix(keyPrefix string) (int64, error) {
	panic("implement me")
}

func (*stubKeystore) GetAll(keyPrefix string) ([]store.KeyValueVersion, error) {
	panic("implement me")
}

func (*stubKeystore) Get(key string) (store.KeyValueVersion, error) {
	panic("implement me")
}

func (*stubKeystore) IsExist(key string) (bool, error) {
	panic("implement me")
}

func (*stubKeystore) Watch(ctxt context.Context, key string, withPrefix bool) store.KeyValueUpdateChan {
	panic("implement me")
}

func (*stubKeystore) KeepAliveKey(ctxt context.Context, key string) error {
	panic("implement me")
}

func (*stubKeystore) NewMutex(lockKey string) (store.Mutex, error) {
	panic("implement me")
}

type stubDacctlActions struct{}

func (*stubDacctlActions) CreatePersistentBuffer(c dacctl.CliContext) error {
	return fmt.Errorf("CreatePersistentBuffer %s", c.String("token"))
}

func (*stubDacctlActions) DeleteBuffer(c dacctl.CliContext) error {
	return fmt.Errorf("DeleteBuffer %s", c.String("token"))
}

func (*stubDacctlActions) CreatePerJobBuffer(c dacctl.CliContext) error {
	return errors.New("CreatePerJobBuffer")
}

func (*stubDacctlActions) ShowInstances() (string, error) {
	return "", errors.New("ShowInstances")
}

func (*stubDacctlActions) ShowSessions() (string, error) {
	return "", errors.New("ShowSessions")
}

func (*stubDacctlActions) ListPools() (string, error) {
	return "", errors.New("ListPools")
}

func (*stubDacctlActions) ShowConfigurations() (string, error) {
	return "", errors.New("ShowConfigurations")
}

func (*stubDacctlActions) ValidateJob(c dacctl.CliContext) error {
	return errors.New("ValidateJob")
}

func (*stubDacctlActions) RealSize(c dacctl.CliContext) (string, error) {
	return "", errors.New("RealSize")
}

func (*stubDacctlActions) DataIn(c dacctl.CliContext) error {
	return errors.New("DataIn")
}

func (*stubDacctlActions) Paths(c dacctl.CliContext) error {
	return errors.New("Paths")
}

func (*stubDacctlActions) PreRun(c dacctl.CliContext) error {
	return errors.New("PreRun")
}

func (*stubDacctlActions) PostRun(c dacctl.CliContext) error {
	return errors.New("PostRun")
}

func (*stubDacctlActions) DataOut(c dacctl.CliContext) error {
	return errors.New("DataOut")
}
