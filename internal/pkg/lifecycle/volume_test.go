package lifecycle

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVolumeLifecycleManager_Mount(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)

	volume := registry.Volume{
		Name: "vol1", SizeBricks: 3, State: registry.BricksProvisioned, JobName: "job1"}
	vlm := NewVolumeLifecycleManager(mockVolReg, nil, volume)
	hosts := []string{"host1", "host2"}

	mockVolReg.EXPECT().UpdateVolumeAttachments(volume.Name, []registry.Attachment{
		{Hostname: "host1", State: registry.RequestAttach, Job: "job1"},
		{Hostname: "host2", State: registry.RequestAttach, Job: "job1"},
	})
	fakeWait := func(volumeName registry.VolumeName, condition func(old *registry.Volume, new *registry.Volume) bool) error {
		old := &registry.Volume{}
		new := &registry.Volume{}
		assert.False(t, condition(old, new))
		new.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.Detached},
			{Hostname: "host1", Job: "job1", State: registry.Attached},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.True(t, condition(old, new))

		new.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.AttachmentError},
			{Hostname: "host1", Job: "job1", State: registry.Detached},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.False(t, condition(old, new))

		new.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.Attached},
			{Hostname: "host1", Job: "job1", State: registry.AttachmentError},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.True(t, condition(old, new))
		return nil
	}
	mockVolReg.EXPECT().WaitForCondition(volume.Name, gomock.Any()).DoAndReturn(fakeWait)

	err := vlm.Mount(hosts, "job1")
	assert.Equal(t, "unable to mount volume: vol1", err.Error())
}
