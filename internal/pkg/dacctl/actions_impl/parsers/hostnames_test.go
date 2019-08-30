package parsers

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_fileio"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetHostnamesFromFile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	disk := mock_fileio.NewMockDisk(mockCtrl)

	fakeHosts := []string{"test1", "test2"}
	disk.EXPECT().Lines("file").Return(fakeHosts, nil)

	hosts, err := GetHostnamesFromFile(disk, "file")
	assert.Nil(t, err)
	assert.Equal(t, fakeHosts, hosts)
}

func TestGetHostnamesFromFile_Empty(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	disk := mock_fileio.NewMockDisk(mockCtrl)

	disk.EXPECT().Lines("file").Return(nil, nil)

	hosts, err := GetHostnamesFromFile(disk, "file")
	assert.Nil(t, err)
	var fakeHosts []string
	assert.Equal(t, fakeHosts, hosts)

	fakeErr := errors.New("bob")
	disk.EXPECT().Lines("file").Return(nil, fakeErr)
	hosts, err = GetHostnamesFromFile(disk, "file")
	assert.Equal(t, fakeHosts, hosts)
	assert.Equal(t, "bob", err.Error())
}

func TestGetHostnamesFromFile_ErrorOnBadHostname(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	disk := mock_fileio.NewMockDisk(mockCtrl)

	fakeHosts := []string{"Test", "test", "test1", "test2.com", "bad hostname", "foo/bar", ""}
	disk.EXPECT().Lines("file").Return(fakeHosts, nil)

	hosts, err := GetHostnamesFromFile(disk, "file")
	assert.Nil(t, hosts)
	assert.Equal(t, "invalid hostname in: [bad hostname foo/bar ]", err.Error())
}
