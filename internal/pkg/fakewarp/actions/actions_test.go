package actions

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type mockCliContext struct{}

func (c *mockCliContext) String(name string) string {
	switch name {
	case "capacity":
		return "pool1:0"
	case "token":
		return "token"
	case "caller":
		return "caller"
	case "user":
		return "user"
	case "access":
		return "access"
	case "type":
		return "type"
	case "job":
		return "jobfile"
	case "nodehostnamefile":
		return "nodehostnamefile1"
	case "pathfile":
		return "pathfile1"
	default:
		return ""
	}
}

func (c *mockCliContext) Int(name string) int {
	return 42 + len(name)
}

func TestCreatePersistentBufferReturnsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockVolumeRegistry(mockCtrl)
	mockObj.EXPECT().AddVolume(gomock.Any()) // TODO
	mockObj.EXPECT().AddJob(gomock.Any())
	mockPool := mocks.NewMockPoolRegistry(mockCtrl)
	mockPool.EXPECT().Pools().DoAndReturn(func() ([]registry.Pool, error) {
		return []registry.Pool{{Name: "pool1", GranularityGB: 1}}, nil
	})
	mockCtxt := &mockCliContext{}

	actions := NewFakewarpActions(mockPool, mockObj, nil)

	if err := actions.CreatePersistentBuffer(mockCtxt); err != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(err))
		t.Fatal("expected success")
	}
}

func TestFakewarpActions_PreRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockDisk := mocks.NewMockDisk(mockCtrl)
	mockCtxt := &mockCliContext{}
	actions := NewFakewarpActions(nil, mockVolReg, mockDisk)
	testVLM = &mockVLM{}
	defer func() { testVLM = nil }()

	mockDisk.EXPECT().Lines("nodehostnamefile1").DoAndReturn(func(string) ([]string, error) {
		return []string{"host1", "host2"}, nil
	})
	mockVolReg.EXPECT().Job("token").DoAndReturn(
		func(name string) (registry.Job, error) {
			return registry.Job{JobVolume: registry.VolumeName("token")}, nil
		})
	mockVolReg.EXPECT().Volume(registry.VolumeName("token"))

	err := actions.PreRun(mockCtxt)
	assert.EqualValues(t, "host1,host2", err.Error())
}

func TestFakewarpActions_Paths(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockDisk := mocks.NewMockDisk(mockCtrl)
	mockCtxt := &mockCliContext{}
	actions := NewFakewarpActions(nil, mockVolReg, mockDisk)
	testVLM = &mockVLM{}
	defer func() { testVLM = nil }()

	mockVolReg.EXPECT().Job("token").DoAndReturn(
		func(name string) (registry.Job, error) {
			return registry.Job{JobVolume: registry.VolumeName("token"), Paths: map[string]string{"a": "A"}}, nil
		})
	mockDisk.EXPECT().Write("pathfile1", []string{"a=A"})

	err := actions.Paths(mockCtxt)
	assert.Nil(t, err)
}

type mockVLM struct{}

func (*mockVLM) ProvisionBricks(pool registry.Pool) error {
	panic("implement me")
}

func (*mockVLM) DataIn() error {
	panic("implement me")
}

func (*mockVLM) Mount(hosts []string) error {
	return errors.New(strings.Join(hosts, ","))
}

func (*mockVLM) Unmount() error {
	panic("implement me")
}

func (*mockVLM) DataOut() error {
	panic("implement me")
}

func (*mockVLM) Delete() error {
	panic("implement me")
}
