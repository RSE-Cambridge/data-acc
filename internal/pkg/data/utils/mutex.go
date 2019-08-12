package utils

import "context"

type Mutex interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}
