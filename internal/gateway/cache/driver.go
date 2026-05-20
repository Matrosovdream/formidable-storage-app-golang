package cache

import (
	"context"
	"time"
)

// Driver is the storage interface for the cache layer.
type Driver interface {
	Get(ctx context.Context, key string, dest any) (found bool, err error)
	Put(ctx context.Context, key string, value any, ttl time.Duration) error
	Forget(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	Name() string
}
