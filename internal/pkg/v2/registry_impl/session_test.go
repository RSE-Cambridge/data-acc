package registry_impl

import (
	"encoding/json"
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_store"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

var exampleSessionString = []byte(`{"Name":"foo","Revision":0,"Owner":0,"Group":0,"CreatedAt":0,"VolumeRequest":{"MultiJob":false,"Caller":"","TotalCapacityBytes":0,"PoolName":"","Access":0,"Type":0,"SwapBytes":0},"Status":{"Error":"","FileSystemCreated":false,"CopyDataInComplete":false,"CopyDataOutComplete":false,"DeleteRequested":false,"DeleteSkipCopyDataOut":false,"UnmountComplete":false,"MountComplete":false},"StageInRequests":null,"StageOutRequests":null,"MultiJobAttachments":null,"Paths":null,"ActualSizeBytes":0,"AllocatedBricks":null,"PrimaryBrickHost":"host1","RequestedAttachHosts":null,"FilesystemStatus":{"Error":"","InternalName":"","InternalData":""},"CurrentAttachments":null}`)
var exampleSession = datamodel.Session{Name: "foo", PrimaryBrickHost: "host1"}

func TestExampleString(t *testing.T) {
	exampleStr, err := json.Marshal(exampleSession)
	assert.Nil(t, err)
	assert.Equal(t, string(exampleSessionString), string(exampleStr))

	var unmarshalSession datamodel.Session
	err = json.Unmarshal(exampleStr, &unmarshalSession)
	assert.Nil(t, err)
	assert.Equal(t, unmarshalSession, exampleSession)

	sessionWithError := datamodel.Session{
		Name: "foo", PrimaryBrickHost: "host1",
		Status: datamodel.SessionStatus{Error: "fake_error"},
	}
	sessionWithErrorStr, err := json.Marshal(sessionWithError)
	assert.Nil(t, err)
	assert.Contains(t, string(sessionWithErrorStr), "fake_error")
}

func TestSessionRegistry_GetSessionMutex(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	fakeErr := errors.New("fake")
	keystore.EXPECT().NewMutex("/session/foo").Return(nil, fakeErr)

	mutex, err := registry.GetSessionMutex("foo")
	assert.Nil(t, mutex)
	assert.Equal(t, fakeErr, err)

	assert.PanicsWithValue(t, "invalid session name: 'foo/bar'", func() {
		registry.GetSessionMutex("foo/bar")
	})
}

func TestSessionRegistry_CreateSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().Create("/session/foo", exampleSessionString).Return(int64(42), nil)

	session, err := registry.CreateSession(exampleSession)
	assert.Nil(t, err)
	assert.Equal(t, int64(42), session.Revision)

	assert.PanicsWithValue(t, "invalid session name: 'foo/bar'", func() {
		registry.CreateSession(datamodel.Session{Name: "foo/bar", PrimaryBrickHost: "host1"})
	})
	assert.PanicsWithValue(t, "session must have allocations before being created: foo", func() {
		registry.CreateSession(datamodel.Session{Name: "foo", ActualSizeBytes: 1024, PrimaryBrickHost: "host1"})
	})
	assert.PanicsWithValue(t, "allocations out of sync with ActualSizeBytes: foo", func() {
		registry.CreateSession(datamodel.Session{
			Name:             "foo",
			AllocatedBricks:  []datamodel.Brick{{}},
			PrimaryBrickHost: "host1",
		})
	})
	assert.PanicsWithValue(t, "PrimaryBrickHost must be set before creating session: foo", func() {
		registry.CreateSession(datamodel.Session{Name: "foo", PrimaryBrickHost: ""})
	})
}

func TestSessionRegistry_GetSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().Get("/session/foo").Return(store.KeyValueVersion{
		ModRevision: 42,
		Value:       exampleSessionString,
	}, nil)

	session, err := registry.GetSession("foo")

	assert.Nil(t, err)
	assert.Equal(t, datamodel.Session{Name: "foo", Revision: 42, PrimaryBrickHost: "host1"}, session)

	assert.PanicsWithValue(t, "invalid session name: 'foo/bar'", func() {
		registry.GetSession("foo/bar")
	})

	fakeErr := errors.New("fake")
	keystore.EXPECT().Get("/session/foo").Return(store.KeyValueVersion{}, fakeErr)
	session, err = registry.GetSession("foo")
	assert.NotNil(t, err)
	assert.Equal(t, "unable to get session due to: fake", err.Error())

	keystore.EXPECT().Get("/session/foo").Return(store.KeyValueVersion{}, nil)
	assert.PanicsWithValue(t,
		"unable parse session from store due to: unexpected end of JSON input",
		func() {
			registry.GetSession("foo")
		})
}

func TestSessionRegistry_GetAllSessions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().GetAll("/session/").Return([]store.KeyValueVersion{{
		ModRevision: 42,
		Value:       exampleSessionString,
	}}, nil)

	sessions, err := registry.GetAllSessions()
	assert.Nil(t, err)
	assert.Equal(t, []datamodel.Session{{Name: "foo", Revision: 42, PrimaryBrickHost: "host1"}}, sessions)

	fakeErr := errors.New("fake")
	keystore.EXPECT().GetAll("/session/").Return(nil, fakeErr)
	sessions, err = registry.GetAllSessions()
	assert.Nil(t, sessions)
	assert.NotNil(t, err)
	assert.Equal(t, "unable to get all sessions due to: fake", err.Error())

	keystore.EXPECT().GetAll("/session/").Return(nil, nil)
	sessions, err = registry.GetAllSessions()
	assert.Nil(t, err)
	assert.Nil(t, sessions)

	keystore.EXPECT().GetAll("/session/").Return([]store.KeyValueVersion{{}}, nil)
	assert.PanicsWithValue(t,
		"unable parse session from store due to: unexpected end of JSON input",
		func() { registry.GetAllSessions() })
}

func TestSessionRegistry_UpdateSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().Update("/session/foo", exampleSessionString, int64(0)).Return(int64(44), nil)

	session, err := registry.UpdateSession(datamodel.Session{Name: "foo", PrimaryBrickHost: "host1", Revision: 0})

	assert.Nil(t, err)
	assert.Equal(t, datamodel.Session{Name: "foo", PrimaryBrickHost: "host1", Revision: 44}, session)
}

func TestSessionRegistry_DeleteSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	fakeErr := errors.New("fake")
	keystore.EXPECT().Delete("/session/foo", int64(40)).Return(fakeErr)

	err := registry.DeleteSession(datamodel.Session{Name: "foo", Revision: 40})

	assert.Equal(t, fakeErr, err)
}
