package dacctl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
)

type pool struct {
	Id          string `json:"id"`
	Units       string `json:"units"`
	Granularity uint   `json:"granularity"`
	Quantity    uint   `json:"quantity"`
	Free        uint   `json:"free"`
}

type pools []pool

func (list *pools) String() string {
	message := map[string]pools{"pools": *list}
	return toJson(message)
}

const GbInBytes = 1073741824

func GetPools(registry registry.PoolRegistry) (*pools, error) {
	pools := pools{}
	regPools, err := registry.Pools()
	if err != nil {
		return &pools, err
	}

	for _, regPool := range regPools {
		free := len(regPool.AvailableBricks)
		quantity := free + len(regPool.AllocatedBricks)
		pools = append(pools, pool{
			Id:          regPool.Name,
			Units:       "bytes",
			Granularity: regPool.GranularityGB * GbInBytes,
			Quantity:    uint(quantity),
			Free:        uint(free),
		})
	}
	return &pools, nil
}
