package filesystem_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GenerateDataCopy(t *testing.T) {
	session := datamodel.Session{
		Name:             "asdf",
		Owner:            1001,
		Group:            1002,
		FilesystemStatus: datamodel.FilesystemStatus{InternalName: "fsuuid"},
		Paths: map[string]string{
			"DW_JOB_STRIPED": "/mnt/lustre/fsuuid/global",
		},
	}
	request := datamodel.DataCopyRequest{}

	cmd, err := generateDataCopyCmd(session, request)
	assert.Nil(t, err)
	assert.Empty(t, cmd)

	request.SourceType = datamodel.File
	request.Source = "$DW_JOB_STRIPED/source"
	request.Destination = "dest"
	cmd, err = generateDataCopyCmd(session, request)
	assert.Nil(t, err)
	assert.Equal(t, "bash -c \"export DW_JOB_STRIPED='/mnt/lustre/fsuuid/global' && sudo -g '#1002' -u '#1001' rsync -ospgu --stats \\$DW_JOB_STRIPED/source dest\"", cmd)

	cmd, err = generateDataCopyCmd(session, request)
	assert.Nil(t, err)
	assert.Equal(t, "bash -c \"export DW_JOB_STRIPED='/mnt/lustre/fsuuid/global' && sudo -g '#1002' -u '#1001' rsync -ospgu --stats \\$DW_JOB_STRIPED/source dest\"", cmd)

	request.SourceType = datamodel.List
	request.Source = "list_filename"
	cmd, err = generateDataCopyCmd(session, request)
	assert.Equal(t, "", cmd)
	assert.Equal(t, "unsupported source type list for volume: asdf", err.Error())

}

func Test_GenerateRsyncCmd(t *testing.T) {
	testVolume := datamodel.Session{
		Name: "asdf",
	}
	request := datamodel.DataCopyRequest{}

	cmd, err := generateRsyncCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Empty(t, cmd)

	request.SourceType = datamodel.File
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateRsyncCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Equal(t, "rsync -ospgu --stats source dest", cmd)

	request.SourceType = datamodel.Directory
	request.Source = "source"
	request.Destination = "dest"
	cmd, err = generateRsyncCmd(testVolume, request)
	assert.Nil(t, err)
	assert.Equal(t, "rsync -r -ospgu --stats source dest", cmd)

	request.SourceType = datamodel.List
	request.Source = "list_filename"
	cmd, err = generateRsyncCmd(testVolume, request)
	assert.Equal(t, "", cmd)
	assert.Equal(t, "unsupported source type list for volume: asdf", err.Error())
}
