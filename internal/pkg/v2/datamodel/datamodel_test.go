package datamodel

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Volume(t *testing.T) {
	volume := Volume{}

	volumeAsString, err := json.Marshal(volume)

	assert.Nil(t, err)
	expected := `{"Name":"","UUID":"","MultiJob":false,"Pool":"","SizeBricks":0,"SizeGB":0,"JobName":"","Owner":0,"Group":0,"CreatedBy":"","CreatedAt":0,"Attachments":null,"AttachGlobalNamespace":false,"AttachPrivateNamespace":false,"AttachAsSwapBytes":0,"AttachPrivateCache":null,"StageInRequests":null,"StageOutRequests":null,"ClientPort":0,"HadBricksAssigned":false}`
	assert.Equal(t, expected, string(volumeAsString))

	var volumeFromString Volume
	data := []byte(expected)
	err = json.Unmarshal(data, &volumeFromString)

	assert.Nil(t, err)
	assert.Equal(t, volume, volumeFromString)
}

func Test_Session(t *testing.T) {
	session := Session{}

	sessionAsString, err := json.Marshal(session)

	assert.Nil(t, err)
	// TODO: not very human readable for Type and Action
	expected := `{"Name":"","Revision":0,"Owner":0,"Group":0,"CreatedAt":0,"VolumeRequest":{"MultiJob":false,"Caller":"","TotalCapacityBytes":0,"PoolName":"","Access":0,"Type":0,"SwapBytes":0},"Status":{"Error":null,"FileSystemCreated":false,"DeleteRequested":false,"DeleteSkipCopyDataOut":false},"StageInRequests":null,"StageOutRequests":null,"MultiJobAttachments":null,"Paths":null,"ActualSizeBytes":0,"Allocations":null,"PrimaryBrickHost":"","AttachHosts":null}`
	assert.Equal(t, expected, string(sessionAsString))

	var sessionFromString Session
	data := []byte(expected)
	err = json.Unmarshal(data, &sessionFromString)

	assert.Nil(t, err)
	assert.Equal(t, session, sessionFromString)
}
