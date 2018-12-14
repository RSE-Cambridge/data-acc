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
