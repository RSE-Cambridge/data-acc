package fake

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GenerateDataCopy(t *testing.T) {
	request := registry.DataCopyRequest{}

	cmd, err := generateDataCopyCmd(registry.VolumeName("asdf"), request)
	assert.Nil(t, err)
	assert.Empty(t, cmd)

	request.SourceType = registry.File
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateDataCopyCmd(registry.VolumeName("asdf"), request)
	assert.Nil(t, err)
	assert.Equal(t, "rsync source dest", cmd)

	request.SourceType = registry.Directory
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateDataCopyCmd(registry.VolumeName("asdf"), request)
	assert.Nil(t, err)
	assert.Equal(t, "rsync -r source dest", cmd)
}
