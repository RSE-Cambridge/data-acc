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
	fakeWait := func(volumeName registry.VolumeName, condition func(event *registry.VolumeChange) bool) error {
		event := &registry.VolumeChange{New: &registry.Volume{}}
		assert.False(t, condition(event))
		event.New.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.Detached},
			{Hostname: "host1", Job: "job1", State: registry.Attached},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.True(t, condition(event))

		event.New.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.AttachmentError},
			{Hostname: "host1", Job: "job1", State: registry.Detached},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.False(t, condition(event))

		event.New.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.Attached},
			{Hostname: "host1", Job: "job1", State: registry.AttachmentError},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.True(t, condition(event))
		return nil
	}
	mockVolReg.EXPECT().WaitForCondition(volume.Name, gomock.Any()).DoAndReturn(fakeWait)

	err := vlm.Mount(hosts, "job1")
	assert.Equal(t, "unable to mount volume: vol1", err.Error())
}

func TestVolumeLifecycleManager_Unmount(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockVolReg := mocks.NewMockVolumeRegistry(mockCtrl)

	volume := registry.Volume{
		Name: "vol1", SizeBricks: 3, State: registry.BricksProvisioned, JobName: "job1",
		Attachments: []registry.Attachment{
			{Hostname: "host1", Job: "job1", State: registry.Attached},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
			{Hostname: "host1", Job: "job2"},
		}}
	vlm := NewVolumeLifecycleManager(mockVolReg, nil, volume)
	hosts := []string{"host1", "host2"}

	mockVolReg.EXPECT().UpdateVolumeAttachments(volume.Name, []registry.Attachment{
		{Hostname: "host1", State: registry.RequestDetach, Job: "job1"},
		{Hostname: "host2", State: registry.RequestDetach, Job: "job1"},
	})
	fakeWait := func(volumeName registry.VolumeName, condition func(event *registry.VolumeChange) bool) error {
		event := &registry.VolumeChange{New: &registry.Volume{}}
		event.New.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2"},
			{Hostname: "host1", Job: "job1", State: registry.Detached},
			{Hostname: "host2", Job: "job1", State: registry.Detached},
		}
		assert.True(t, condition(event))

		event.New.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2", State: registry.AttachmentError},
			{Hostname: "host1", Job: "job1", State: registry.Detached},
			{Hostname: "host2", Job: "job1", State: registry.Attached},
		}
		assert.False(t, condition(event))

		event.New.Attachments = []registry.Attachment{
			{Hostname: "host1", Job: "job2"},
			{Hostname: "host1", Job: "job1", State: registry.AttachmentError},
			{Hostname: "host2", Job: "job1", State: registry.Detached},
		}
		assert.True(t, condition(event))
		return nil
	}
	mockVolReg.EXPECT().WaitForCondition(volume.Name, gomock.Any()).DoAndReturn(fakeWait)

	err := vlm.Unmount(hosts, "job2")
	assert.Equal(t, "attachment must be attached to do unmount for volume: vol1", err.Error())

	err = vlm.Unmount(hosts, "job3")
	assert.Equal(t, "can't find attachment for volume: vol1 host: host1 job: job3", err.Error())

	err = vlm.Unmount(hosts, "job1")
	assert.Equal(t, "unable to unmount volume: vol1 because: attachment for host host1 in error state", err.Error())
}
