package fakewarp

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPools(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockObj := keystoreregistry.NewMockKeystore(mockCtrl)

	pools, _ := GetPools(keystoreregistry.NewPoolRegistry(mockObj))
	actual := pools.String()
	expected := `{"pools":[{"id":"fake","units":"bytes","granularity":214748364800,"quantity":400,"free":395}]}`
	assert.EqualValues(t, expected, actual[:len(actual)-1])
	assert.EqualValues(t, "\n", actual[len(actual)-1:])
}
