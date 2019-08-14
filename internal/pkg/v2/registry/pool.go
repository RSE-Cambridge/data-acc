package registry

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
)

type PoolRegistry interface {
	// Get all registered pools
	GetPools() ([]datamodel.Pool, error)

	// Creates the pool if it doesn't exist
	// error if the granularity doesn't match and existing pool
	EnsurePoolCreated(poolName datamodel.PoolName, granularityGB int) (datamodel.Pool, error)
}
