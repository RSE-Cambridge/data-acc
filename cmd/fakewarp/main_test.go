package main

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/golang/mock/gomock"
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

func TestRunCliAcceptsRequiredArgs(t *testing.T) {
	// TODO: has to be a better way to do this
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKeystore := keystoreregistry.NewMockKeystore(mockCtrl)
	mockKeystore.EXPECT().CleanPrefix("/buffers/a")
	mockKeystore.EXPECT().CleanPrefix("/buffers/a2")
	mockKeystore.EXPECT().AtomicAdd("/buffers/a", gomock.Any())
	mockKeystore.EXPECT().AtomicAdd("/buffers/p1", gomock.Any())
	mockKeystore.EXPECT().AtomicAdd("/buffers/p2", gomock.Any())
	mockKeystore.EXPECT().Close().AnyTimes()
	testKeystore = mockKeystore
	defer func() {
		testKeystore = nil
	}()

	if err := runCli([]string{"--function", "pools"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "show_instances"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "show_sessions"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "teardown", "--job", "a", "--token", "a"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "teardown", "--job", "a", "--token", "a2", "--hurry"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "job_process", "--job", "a"}); err != nil {
		t.Fatal(err)
	}
	setupArgs := strings.Split(
		"--function setup --token a --job b --caller c --user 1 --groupid 1 --capacity dw:1GiB", " ")
	if err := runCli(setupArgs); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "real_size", "--token", "a"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "data_in", "--token", "a", "--job", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "paths", "--token", "a", "--job", "b", "--pathfile", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "pre_run", "--token", "a", "--job", "b", "--nodehostnamefile", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "post_run", "--token", "a", "--job", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := runCli([]string{"--function", "data_out", "--token", "a", "--job", "b"}); err != nil {
		t.Fatal(err)
	}
	createPersistentArgs := strings.Split(
		"--function create_persistent --token p1 --caller c --user 1 --groupid 1 --capacity dw:1GiB "+
			"--access striped --type scratch", " ")
	if err := runCli(createPersistentArgs); err != nil {
		t.Fatal(err)
	}
	createPersistentArgs = strings.Split(
		"--function create_persistent -t p2 -c c -u 1 -g 1 -C dw:1GiB -a striped -T scratch", " ")
	if err := runCli(createPersistentArgs); err != nil {
		t.Fatal(err)
	}
}
