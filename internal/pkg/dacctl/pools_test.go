package dacctl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPools(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockReg := mocks.NewMockPoolRegistry(mockCtrl)
	fakePools := func() ([]registry.Pool, error) {
		return []registry.Pool{{Name: "fake", GranularityGB: 1}}, nil
	}
	mockReg.EXPECT().Pools().DoAndReturn(fakePools)

	pools, _ := GetPools(mockReg)
	actual := pools.String()
	expected := `{"pools":[{"id":"fake","units":"bytes","granularity":1073741824,"quantity":0,"free":0}]}`
	assert.EqualValues(t, expected, actual)
}
