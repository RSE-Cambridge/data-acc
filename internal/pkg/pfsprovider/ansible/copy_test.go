package ansible

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GenerateDataCopy(t *testing.T) {
	testVolume := registry.Volume{
		Name:  registry.VolumeName("asdf"),
		Owner: 1001,
	}
	request := registry.DataCopyRequest{}

	cmd, err := generateDataCopyCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Empty(t, cmd)

	request.SourceType = registry.File
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateDataCopyCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Equal(t, "sudo su `getent passwd 1001 | cut -d: -f1` rsync source dest", cmd)

	request.SourceType = registry.List
	request.Source = "list_filename"
	cmd, err = generateDataCopyCmd(testVolume, request)
	assert.Equal(t, "", cmd)
	assert.Equal(t, "unsupported source type list for volume: asdf", err.Error())

}

func Test_GenerateRsyncCmd(t *testing.T) {
	testVolume := registry.Volume{
		Name: registry.VolumeName("asdf"),
	}
	request := registry.DataCopyRequest{}

	cmd, err := generateRsyncCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Empty(t, cmd)

	request.SourceType = registry.File
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateRsyncCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Equal(t, "rsync source dest", cmd)

	request.SourceType = registry.Directory
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateRsyncCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Equal(t, "rsync -r source dest", cmd)

	request.SourceType = registry.List
	request.Source = "list_filename"
	cmd, err = generateRsyncCmd(testVolume, request)
	assert.Equal(t, "", cmd)
	assert.Equal(t, "unsupported source type list for volume: asdf", err.Error())
}
