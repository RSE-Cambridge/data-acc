package registry_impl

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_store"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptySessionString = []byte(`{"Name":"foo","Revision":0,"Owner":0,"Group":0,"CreatedAt":0,"VolumeRequest":{"MultiJob":false,"Caller":"","TotalCapacityBytes":0,"PoolName":"","Access":0,"Type":0,"SwapBytes":0},"Status":{"Error":null,"FileSystemCreated":false,"CopyDataInComplete":false,"CopyDataOutComplete":false,"DeleteRequested":false,"DeleteSkipCopyDataOut":false},"StageInRequests":null,"StageOutRequests":null,"MultiJobAttachments":null,"Paths":null,"ActualSizeBytes":0,"Allocations":null,"PrimaryBrickHost":"","RequestedAttachHosts":null,"FilesystemStatus":{"Error":null,"InternalName":"","InternalData":""},"CurrentAttachments":null}`)

func TestSessionRegistry_GetSessionMutex(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	fakeErr := errors.New("fake")
	keystore.EXPECT().NewMutex("/lock/session/foo").Return(nil, fakeErr)

	mutex, err := registry.GetSessionMutex("foo")
	assert.Nil(t, mutex)
	assert.Equal(t, fakeErr, err)

	mutex, err = registry.GetSessionMutex("foo/bar")
	assert.Nil(t, mutex)
	assert.Equal(t, "invalid session name foo/bar", err.Error())
}

func TestSessionRegistry_CreateSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().Create("/session/foo", emptySessionString).Return(store.KeyValueVersion{ModRevision: 42}, nil)

	session, err := registry.CreateSession(datamodel.Session{Name: "foo"})
	assert.Nil(t, err)
	assert.Equal(t, int64(42), session.Revision)

	session, err = registry.CreateSession(datamodel.Session{Name: "foo/bar"})
	assert.NotNil(t, err)
	assert.Equal(t, "invalid session name foo/bar", err.Error())

	_, err = registry.CreateSession(datamodel.Session{Name: "foo", ActualSizeBytes: 1024})
	assert.NotNil(t, err)
	assert.Equal(t, "session must have allocations before being created", err.Error())

	_, err = registry.CreateSession(datamodel.Session{
		Name:            "foo",
		ActualSizeBytes: 1024,
		Allocations:     []datamodel.BrickAllocation{{}},
	})
	assert.NotNil(t, err)
	assert.Equal(t, "session must have a primary brick host set", err.Error())

	_, err = registry.CreateSession(datamodel.Session{
		Name:        "foo",
		Allocations: []datamodel.BrickAllocation{{}},
	})
	assert.NotNil(t, err)
	assert.Equal(t, "allocations out of sync with ActualSizeBytes", err.Error())

	_, err = registry.CreateSession(datamodel.Session{Name: "foo", PrimaryBrickHost: "foo"})
	assert.NotNil(t, err)
	assert.Equal(t, "PrimaryBrickHost should be empty if no bricks assigned", err.Error())
}

func TestSessionRegistry_GetSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().Get("/session/foo").Return(store.KeyValueVersion{
		ModRevision: 42,
		Value:       emptySessionString,
	}, nil)

	session, err := registry.GetSession("foo")

	assert.Nil(t, err)
	assert.Equal(t, datamodel.Session{Name: "foo", Revision: 42}, session)

	session, err = registry.GetSession("foo/bar")
	assert.NotNil(t, err)
	assert.Equal(t, "invalid session name foo/bar", err.Error())

	fakeErr := errors.New("fake")
	keystore.EXPECT().Get("/session/foo").Return(store.KeyValueVersion{}, fakeErr)
	session, err = registry.GetSession("foo")
	assert.NotNil(t, err)
	assert.Equal(t, "unable to get session due to: fake", err.Error())

	keystore.EXPECT().Get("/session/foo").Return(store.KeyValueVersion{}, nil)
	session, err = registry.GetSession("foo")
	assert.NotNil(t, err)
	assert.Equal(t, "unable parse session from store due to: unexpected end of JSON input", err.Error())
}

func TestSessionRegistry_GetAllSessions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	keystore := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(keystore)
	keystore.EXPECT().GetAll("/session/").Return([]store.KeyValueVersion{{
		ModRevision: 42,
		Value:       emptySessionString,
	}}, nil)

	sessions, err := registry.GetAllSessions()
	assert.Nil(t, err)
	assert.Equal(t, []datamodel.Session{{Name: "foo", Revision: 42}}, sessions)

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
	keystore.EXPECT().Update("/session/foo", emptySessionString, int64(0)).Return(store.KeyValueVersion{
		ModRevision: 44,
		Value:       emptySessionString,
	}, nil)

	session, err := registry.UpdateSession(datamodel.Session{Name: "foo", Revision: 0})

	assert.Nil(t, err)
	assert.Equal(t, datamodel.Session{Name: "foo", Revision: 44}, session)
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
