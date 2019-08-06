package dacctl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func assertNewline(t *testing.T, actual string) {
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}

func TestGetInstances(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockReg := mocks.NewMockVolumeRegistry(mockCtrl)
	fakeGetVolumes := func() ([]registry.Volume, error) {
		return []registry.Volume{
			{Name: "fake1", Pool: "pool1", SizeGB: 2},
		}, nil
	}
	mockReg.EXPECT().AllVolumes().DoAndReturn(fakeGetVolumes)

	instances, err := GetInstances(mockReg)
	if err != nil {
		log.Fatal(err)
	}
	actual := instances.String()

	// TODO need to return sessions correctly... i.e. job
	expected := `{"instances":[{"id":"fake1","capacity":{"bytes":2147483648,"nodes":0},"links":{"session":""}}]}`
	assert.EqualValues(t, expected, actual)
}

func TestGetSessions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockReg := mocks.NewMockVolumeRegistry(mockCtrl)
	mockJobs := func() ([]registry.Job, error) {
		return []registry.Job{{Name: "fake1", CreatedAt: 42, Owner: 1001}}, nil

	}
	mockReg.EXPECT().Jobs().DoAndReturn(mockJobs)
	sessions, err := GetSessions(mockReg)
	if err != nil {
		log.Fatal(err)
	}
	actual := sessions.String()

	expected := `{"sessions":[{"id":"fake1","created":42,"owner":1001,"token":"fake1"}]}`
	assert.EqualValues(t, expected, actual)
}

func TestGetConfigurations(t *testing.T) {
	actual := GetConfigurations().String()
	expected := `{"configurations":[]}`
	assert.EqualValues(t, expected, actual)
}
