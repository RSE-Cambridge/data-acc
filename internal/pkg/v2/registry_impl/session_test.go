package registry_impl

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionRegistry_GetSessionMutex(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	store := mock_store.NewMockKeystore(mockCtrl)
	registry := NewSessionRegistry(store)
	fakeErr := errors.New("fake")
	store.EXPECT().NewMutex("/session_lock/foo").Return(nil, fakeErr)

	mutex, err := registry.GetSessionMutex("foo")

	assert.Nil(t, mutex)
	assert.Equal(t, fakeErr, err)
}
