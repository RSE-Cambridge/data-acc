package actions

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockCliContext struct {
	capacity int
}

func (c *mockCliContext) String(name string) string {
	switch name {
	case "capacity":
		return fmt.Sprintf("pool1:%dGB", c.capacity)
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
	switch name {
	case "user":
		return 1001
	case "group":
		return 1001
	default:
		return 42 + len(name)
	}
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

	mockPool.EXPECT().AllocateBricksForVolume(gomock.Any())
	mockCtxt := &mockCliContext{}

	actions := NewDacctlActions(mockPool, mockObj, nil)

	err := actions.CreatePersistentBuffer(mockCtxt)
	assert.Nil(t, err)
}

func TestDacctlActions_PreRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockDisk := mocks.NewMockDisk(mockCtrl)
	mockCtxt := &mockCliContext{}
	actions := NewDacctlActions(nil, mockVolReg, mockDisk)
	testVLM = &mockVLM{}
	defer func() { testVLM = nil }()

	mockDisk.EXPECT().Lines("nodehostnamefile1").DoAndReturn(func(string) ([]string, error) {
		return []string{"host1", "host2"}, nil
	})
	mockVolReg.EXPECT().Job("token").DoAndReturn(
		func(name string) (registry.Job, error) {
			return registry.Job{
				Name:            "token",
				JobVolume:       registry.VolumeName("token"),
				MultiJobVolumes: []registry.VolumeName{registry.VolumeName("othervolume")},
			}, nil
		})
	mockVolReg.EXPECT().JobAttachHosts("token", []string{"host1", "host2"})
	mockVolReg.EXPECT().Volume(registry.VolumeName("token"))
	mockVolReg.EXPECT().Volume(registry.VolumeName("othervolume")).DoAndReturn(
		func(name registry.VolumeName) (registry.Volume, error) {
			return registry.Volume{Name: name}, nil
		})

	err := actions.PreRun(mockCtxt)
	assert.Nil(t, err)
}

func TestDacctlActions_Paths(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockDisk := mocks.NewMockDisk(mockCtrl)
	mockCtxt := &mockCliContext{}
	actions := NewDacctlActions(nil, mockVolReg, mockDisk)
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

func TestDacctlActions_CreatePerJobBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockPoolReg := mocks.NewMockPoolRegistry(mockCtrl)
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockDisk := mocks.NewMockDisk(mockCtrl)
	mockCtxt := &mockCliContext{capacity: 2}
	actions := NewDacctlActions(mockPoolReg, mockVolReg, mockDisk)

	mockDisk.EXPECT().Lines("jobfile").DoAndReturn(func(string) ([]string, error) {
		return []string{
			"#DW persistentdw name=mybuffer",
			"#DW jobdw capacity=2GB access_mode=striped,private type=scratch",
		}, nil
	})

	mockPoolReg.EXPECT().Pools().DoAndReturn(func() ([]registry.Pool, error) {
		return []registry.Pool{{Name: "pool1", GranularityGB: 200}}, nil
	})
	mockVolReg.EXPECT().Volume(registry.VolumeName("mybuffer")).DoAndReturn(
		func(name registry.VolumeName) (registry.Volume, error) {
			return registry.Volume{Name: name, MultiJob: true}, nil
		})
	expectedVolume := registry.Volume{
		Name:                   "token",
		MultiJob:               false,
		State:                  registry.Registered,
		Pool:                   "pool1",
		SizeBricks:             2,
		SizeGB:                 400,
		JobName:                "token",
		Owner:                  1001,
		Group:                  1001,
		CreatedBy:              "caller",
		CreatedAt:              uint(time.Now().Unix()), // TODO this is racey!
		AttachGlobalNamespace:  true,
		AttachPrivateNamespace: true,
		AttachAsSwapBytes:      0,
	}
	mockVolReg.EXPECT().AddVolume(expectedVolume)
	mockVolReg.EXPECT().AddJob(registry.Job{
		Name:      "token",
		Owner:     1001,
		CreatedAt: uint(time.Now().Unix()),
		Paths: map[string]string{
			"DW_PERSISTENT_STRIPED_mybuffer": "/dac/token_persistent_mybuffer",
			"DW_JOB_PRIVATE":                 "/dac/token_job_private",
			"DW_JOB_STRIPED":                 "/dac/token_job/global",
		},
		JobVolume:       registry.VolumeName("token"),
		MultiJobVolumes: []registry.VolumeName{"mybuffer"},
	})
	mockVolReg.EXPECT().Volume(registry.VolumeName("token")).DoAndReturn(
		func(name registry.VolumeName) (registry.Volume, error) {
			return registry.Volume{
				Name:       name,
				SizeBricks: 0, // TODO: skips ProvisionBricks logic
			}, nil
		})
	// TODO: sort out the volume passed here!
	mockPoolReg.EXPECT().AllocateBricksForVolume(gomock.Any())

	err := actions.CreatePerJobBuffer(mockCtxt)
	assert.Nil(t, err)
}

type mockVLM struct{}

func (*mockVLM) ProvisionBricks() error {
	panic("implement me")
}

func (*mockVLM) DataIn() error {
	panic("implement me")
}

func (*mockVLM) Mount(hosts []string, jobName string) error {
	return nil
}

func (*mockVLM) Unmount(hosts []string, jobName string) error {
	panic("implement me")
}

func (*mockVLM) DataOut() error {
	panic("implement me")
}

func (*mockVLM) Delete() error {
	panic("implement me")
}
