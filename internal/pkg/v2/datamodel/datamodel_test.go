package datamodel

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Session(t *testing.T) {
	session := Session{}

	sessionAsString, err := json.Marshal(session)

	assert.Nil(t, err)
	// TODO: not very human readable for Type and ActionType
	expected := `{"Name":"","Revision":0,"Owner":0,"Group":0,"CreatedAt":0,"VolumeRequest":{"MultiJob":false,"Caller":"","TotalCapacityBytes":0,"PoolName":"","Access":0,"Type":0,"SwapBytes":0},"Status":{"Error":null,"FileSystemCreated":false,"CopyDataInComplete":false,"CopyDataOutComplete":false,"DeleteRequested":false,"DeleteSkipCopyDataOut":false},"StageInRequests":null,"StageOutRequests":null,"MultiJobAttachments":null,"Paths":null,"ActualSizeBytes":0,"Allocations":null,"PrimaryBrickHost":"","RequestedAttachHosts":null,"FilesystemStatus":{"Error":null,"InternalName":"","InternalData":""},"CurrentAttachments":null}`
	assert.Equal(t, expected, string(sessionAsString))

	var sessionFromString Session
	data := []byte(expected)
	err = json.Unmarshal(data, &sessionFromString)

	assert.Nil(t, err)
	assert.Equal(t, session, sessionFromString)
}
