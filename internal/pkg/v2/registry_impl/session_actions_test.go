package registry_impl

import (
	"context"
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_store"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionActions_SendSessionAction(t *testing.T) {
	// TODO: need way more testing here
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	brickHost := mock_registry.NewMockBrickHostRegistry(mockCtrl)
	keystore := mock_store.NewMockKeystore(mockCtrl)
	actions := sessionActions{brickHostRegistry: brickHost, store: keystore}
	session := datamodel.Session{Name:"foo", PrimaryBrickHost:"host1"}
	brickHost.EXPECT().IsBrickHostAlive(session.PrimaryBrickHost).Return(true, nil)
	keystore.EXPECT().Watch(context.TODO(), gomock.Any(), false).Return(nil)
	fakeErr := errors.New("fake")
	keystore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(store.KeyValueVersion{}, fakeErr)

	channel, err := actions.SendSessionAction(context.TODO(), datamodel.SessionCreateFilesystem, session)

	assert.Nil(t, channel)
	assert.Equal(t, "unable to send session action due to: fake", err.Error())
}
