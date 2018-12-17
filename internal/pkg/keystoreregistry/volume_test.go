package keystoreregistry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVolumeRegistry_DeleteVolumeAttachments(t *testing.T) {
	vol := registry.Volume{
		Attachments: []registry.Attachment{
			{Hostname: "host1", Job: "job1"},
			{Hostname: "host2", Job: "job1"},
			{Hostname: "host2", Job: "job2"},
		},
	}
	numRemoved := removeAttachments(&vol, "job1", []string{"host1", "host2"})
	assert.Equal(t, 2, numRemoved)
	assert.Equal(t, 1, len(vol.Attachments))
}

func TestVolumeRegistry_FindAttachment(t *testing.T) {
	attachments := []registry.Attachment{
		{Job:"job1", Hostname:"foo1"}, {Job:"job2", Hostname:"foo1"}, {Job:"job2", Hostname:"foo2"},
	}

	attachment, ok := findAttachment(nil, "", "")
	assert.Nil(t, attachment)
	assert.False(t, ok)

	attachment, ok = findAttachment(attachments, "foo2", "job1")
	assert.Nil(t, attachment)
	assert.False(t, ok)

	attachment, ok = findAttachment(attachments, "foo1", "job1")
	assert.True(t, ok)
	assert.Equal(t, registry.Attachment{Job: "job1", Hostname:"foo1"}, *attachment)

	attachment, ok = findAttachment(attachments, "foo1", "job2")
	assert.True(t, ok)
	assert.Equal(t, registry.Attachment{Job: "job2", Hostname:"foo1"}, *attachment)
}

func TestVolumeRegistry_MergeAttachments(t *testing.T) {
	oldAttachments := []registry.Attachment{
		{Job:"job1", Hostname:"foo1"}, {Job:"job2", Hostname:"foo1"}, {Job:"job2", Hostname:"foo2"},
	}

	assert.Nil(t, mergeAttachments(nil, nil))
	assert.Equal(t, oldAttachments, mergeAttachments(oldAttachments, nil))
	assert.Equal(t, oldAttachments, mergeAttachments(nil, oldAttachments))

	// add new
	result := mergeAttachments(oldAttachments, []registry.Attachment{{Job:"foo", Hostname:"foo"}})
	assert.Equal(t, 4, len(result))
	assert.Equal(t, registry.Attachment{Job:"foo", Hostname:"foo"}, result[0])
	assert.Equal(t, oldAttachments[0], result[1])
	assert.Equal(t, oldAttachments[1], result[2])
	assert.Equal(t, oldAttachments[2], result[3])

	// in place update
	updates := []registry.Attachment{
		{Job:"job2", Hostname:"foo1", State:registry.Attached},
		{Job:"job2", Hostname:"foo2", State:registry.Attached},
	}
	result = mergeAttachments(oldAttachments, updates)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, updates[0], result[0])
	assert.Equal(t, updates[1], result[1])
	assert.Equal(t, oldAttachments[0], result[2])
}