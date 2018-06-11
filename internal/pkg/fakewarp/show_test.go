package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
)

func assertNewline(t *testing.T, actual string) {
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}

func TestGetInstances(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockKeystore(mockCtrl)
	mockReg := keystoreregistry.NewVolumeRegistry(mockObj)

	instances, err := GetInstances(mockReg)
	if err != nil {
		log.Fatal(err)
	}
	actual := instances.String()

	expected := `{"instances":[{"id":"fakebuffer","capacity":{"bytes":3,"nodes":40},"links":{"session":"fakebuffer"}}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assertNewline(t, actual)
}

func TestGetSessions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockKeystore(mockCtrl)
	mockReg := keystoreregistry.NewVolumeRegistry(mockObj)

	sessions, err := GetSessions(mockReg)
	if err != nil {
		log.Fatal(err)
	}
	actual := sessions.String()

	expected := `{"sessions":[{"id":"fakebuffer","created":1234567890,"owner":1001,"token":"fakebuffer"}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assertNewline(t, actual)
}

func TestGetConfigurations(t *testing.T) {
	actual := GetConfigurations().String()
	expected := `{"configurations":[]}`
	assert.EqualValues(t, expected, actual)
}
