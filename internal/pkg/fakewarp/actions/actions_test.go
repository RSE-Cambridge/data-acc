package actions

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
		return "nodehostnamefile"
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
	mockCtxt := &mockCliContext{}
	actions := NewFakewarpActions(nil, mockVolReg, nil)
	testVLM = &mockVLM{}
	defer func() { testVLM = nil }()

	mockVolReg.EXPECT().Volume(registry.VolumeName("token"))

	err := actions.PreRun(mockCtxt)
	assert.EqualValues(t, "mount", err.Error())
}

type mockVLM struct{}

func (*mockVLM) ProvisionBricks(pool registry.Pool) error {
	panic("implement me")
}

func (*mockVLM) DataIn() error {
	panic("implement me")
}

func (*mockVLM) Mount() error {
	return errors.New("mount")
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
