package keystoreregistry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/oldregistry"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestBufferRegistry_AddBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	key := "/buffers/test"
	buff := oldregistry.Buffer{Name: "test"}
	mockObj := NewMockKeystore(mockCtrl)
	mockObj.EXPECT().AtomicAdd(key, gomock.Any())

	reg := NewBufferRegistry(mockObj)
	reg.AddBuffer(buff)
}

func TestBufferRegistry_RemoveBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	prefix := "/buffers/test"
	buff := oldregistry.Buffer{Name: "test"}
	mockObj := NewMockKeystore(mockCtrl)
	mockObj.EXPECT().CleanPrefix(prefix)

	reg := NewBufferRegistry(mockObj)
	reg.RemoveBuffer(buff)
}
