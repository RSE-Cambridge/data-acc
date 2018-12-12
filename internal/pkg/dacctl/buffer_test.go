package dacctl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatePerJobBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockPoolReg := mocks.NewMockPoolRegistry(mockCtrl)
	mockDisk := mocks.NewMockDisk(mockCtrl)
	mockDisk.EXPECT().Lines("jobfile")

	err := CreatePerJobBuffer(mockVolReg, mockPoolReg, mockDisk, "token",
		2, 2, "", "test", "jobfile")
	assert.Equal(t, "must format capacity correctly and include pool", err.Error())
}

func TestGetPerJobVolume(t *testing.T) {
	pool := registry.Pool{}
	summary := jobSummary{
		PerJobBuffer: &cmdPerJobBuffer{},
	}
	volume := getPerJobVolume("token", &pool, 3, 42, 42,
		"test", 20, summary)
	// TODO: lots more work to do here!
	assert.Equal(t, registry.VolumeName("token"), volume.Name)
	assert.True(t, volume.AttachGlobalNamespace)
	assert.False(t, volume.AttachPrivateNamespace)
}

func TestSetPaths(t *testing.T) {
	volume := registry.Volume{
		Name:                   "job1",
		UUID:                   "uuid1",
		AttachPrivateNamespace: true,
		AttachGlobalNamespace:  true,
	}
	job := registry.Job{
		Name: "job1",
		MultiJobVolumes: []registry.VolumeName{
			registry.VolumeName("multi1"),
			registry.VolumeName("multi2"),
		},
	}

	paths := setPaths(&volume, job)

	assert.Equal(t, 4, len(paths))
	assert.Equal(t,
		"/dac/job1/private",
		paths["DW_JOB_PRIVATE"])
	assert.Equal(t,
		"/dac/job1/global",
		paths["DW_JOB_STRIPED"])
	assert.Equal(t,
		"/dac/job1/persistent/multi1",
		paths["DW_PERSISTENT_STRIPED_multi1"])
	assert.Equal(t,
		"/dac/job1/persistent/multi2",
		paths["DW_PERSISTENT_STRIPED_multi2"])
}
