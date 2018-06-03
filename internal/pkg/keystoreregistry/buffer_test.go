package keystoreregistry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_keystoneregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestBufferRegistry_AddBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	key := "/buffers/test"
	value := "foo"
	buff := registry.Buffer{Name: "test", Owner: value}
	mockObj := mock_keystoreregistry.NewMockKeystore(mockCtrl)
	mockObj.EXPECT().AtomicAdd(key, value)

	reg := NewBufferRegistry(mockObj)
	reg.AddBuffer(buff)
}

func TestBufferRegistry_RemoveBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	prefix := "/buffers/test"
	buff := registry.Buffer{Name: "test"}
	mockObj := mock_keystoreregistry.NewMockKeystore(mockCtrl)
	mockObj.EXPECT().CleanPrefix(prefix)

	reg := NewBufferRegistry(mockObj)
	reg.RemoveBuffer(buff)
}
