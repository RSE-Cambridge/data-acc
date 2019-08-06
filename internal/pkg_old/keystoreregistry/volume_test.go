package keystoreregistry

import (
	"context"
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
		{Job: "job1", Hostname: "foo1"}, {Job: "job2", Hostname: "foo1"}, {Job: "job2", Hostname: "foo2"},
	}

	attachment, ok := findAttachment(nil, "", "")
	assert.Nil(t, attachment)
	assert.False(t, ok)

	attachment, ok = findAttachment(attachments, "foo2", "job1")
	assert.Nil(t, attachment)
	assert.False(t, ok)

	attachment, ok = findAttachment(attachments, "foo1", "job1")
	assert.True(t, ok)
	assert.Equal(t, registry.Attachment{Job: "job1", Hostname: "foo1"}, *attachment)

	attachment, ok = findAttachment(attachments, "foo1", "job2")
	assert.True(t, ok)
	assert.Equal(t, registry.Attachment{Job: "job2", Hostname: "foo1"}, *attachment)
}

func TestVolumeRegistry_MergeAttachments(t *testing.T) {
	oldAttachments := []registry.Attachment{
		{Job: "job1", Hostname: "foo1"}, {Job: "job2", Hostname: "foo1"}, {Job: "job2", Hostname: "foo2"},
	}

	assert.Nil(t, mergeAttachments(nil, nil))
	assert.Equal(t, oldAttachments, mergeAttachments(oldAttachments, nil))
	assert.Equal(t, oldAttachments, mergeAttachments(nil, oldAttachments))

	// add new
	result := mergeAttachments(oldAttachments, []registry.Attachment{{Job: "foo", Hostname: "foo"}})
	assert.Equal(t, 4, len(result))
	assert.Equal(t, registry.Attachment{Job: "foo", Hostname: "foo"}, result[0])
	assert.Equal(t, oldAttachments[0], result[1])
	assert.Equal(t, oldAttachments[1], result[2])
	assert.Equal(t, oldAttachments[2], result[3])

	// in place update
	updates := []registry.Attachment{
		{Job: "job2", Hostname: "foo1", State: registry.Attached},
		{Job: "job2", Hostname: "foo2", State: registry.Attached},
	}
	result = mergeAttachments(oldAttachments, updates)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, updates[0], result[0])
	assert.Equal(t, updates[1], result[1])
	assert.Equal(t, oldAttachments[0], result[2])
}

func TestVolumeRegistry_GetVolumeChanges_nil(t *testing.T) {
	volReg := volumeRegistry{keystore: fakeKeystore{watchChan: nil, t: t, key: "/volume/vol1/"}}

	changes := volReg.GetVolumeChanges(context.TODO(), registry.Volume{Name: "vol1"})

	_, ok := <-changes
	assert.False(t, ok)
}

func TestVolumeRegistry_GetVolumeChanges(t *testing.T) {
	raw := make(chan KeyValueUpdate)
	volReg := volumeRegistry{keystore: fakeKeystore{
		watchChan: raw, t: t, key: "/volume/vol1/", withPrefix: false,
	}}

	changes := volReg.GetVolumeChanges(context.TODO(), registry.Volume{Name: "vol1"})

	vol := &registry.Volume{Name: "test1"}

	go func() {
		raw <- KeyValueUpdate{
			New: &KeyValueVersion{Key: "asdf", Value: toJson(vol)},
			Old: &KeyValueVersion{Key: "asdf", Value: toJson(vol)},
		}
		raw <- KeyValueUpdate{
			IsDelete: true,
			New:      nil,
			Old:      &KeyValueVersion{Key: "asdf", Value: toJson(vol)},
		}
		raw <- KeyValueUpdate{
			New: &KeyValueVersion{Key: "asdf", Value: "asdf"},
			Old: nil,
		}
		close(raw)
	}()

	ch1 := <-changes
	assert.Nil(t, ch1.Err)
	assert.False(t, ch1.IsDelete)
	assert.Equal(t, vol, ch1.Old)
	assert.Equal(t, vol, ch1.New)

	ch2 := <-changes
	assert.Nil(t, ch2.Err)
	assert.True(t, ch2.IsDelete)
	assert.Nil(t, ch2.New)
	assert.Equal(t, vol, ch1.Old)

	ch3 := <-changes
	assert.Equal(t, "invalid character 'a' looking for beginning of value", ch3.Err.Error())
	assert.False(t, ch3.IsDelete)
	assert.Nil(t, ch3.Old)
	assert.Nil(t, ch3.New)

	_, ok := <-changes
	assert.False(t, ok)
	_, ok = <-raw
	assert.False(t, ok)
}
