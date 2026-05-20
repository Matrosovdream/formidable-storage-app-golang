package usecase

import (
	"context"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/cache"
)

type CacheUseCase struct {
	driver     cache.Driver
	Keys       *cache.KeyBuilder
	defaultTTL time.Duration
}

func NewCacheUseCase(driver cache.Driver, keys *cache.KeyBuilder, defaultTTL time.Duration) *CacheUseCase {
	return &CacheUseCase{driver: driver, Keys: keys, defaultTTL: defaultTTL}
}

func (c *CacheUseCase) Driver() cache.Driver { return c.driver }

func (c *CacheUseCase) Get(ctx context.Context, key string, dest any) (bool, error) {
	return c.driver.Get(ctx, key, dest)
}

func (c *CacheUseCase) Put(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	return c.driver.Put(ctx, key, value, ttl)
}

func (c *CacheUseCase) Forget(ctx context.Context, key string) error {
	return c.driver.Forget(ctx, key)
}

// Remember returns the cached value if present, otherwise invokes fn, caches its result, and stores into dest.
// fn MUST return a value JSON-compatible with dest.
func (c *CacheUseCase) Remember(ctx context.Context, key string, ttl time.Duration, dest any, fn func() (any, error)) error {
	_, err := c.RememberTracked(ctx, key, ttl, dest, fn)
	return err
}

// RememberTracked is Remember plus a boolean indicating whether the value came from cache (true=hit).
func (c *CacheUseCase) RememberTracked(ctx context.Context, key string, ttl time.Duration, dest any, fn func() (any, error)) (hit bool, err error) {
	found, err := c.driver.Get(ctx, key, dest)
	if err != nil {
		return false, err
	}
	if found {
		return true, nil
	}
	value, err := fn()
	if err != nil {
		return false, err
	}
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	if err := c.driver.Put(ctx, key, value, ttl); err != nil {
		return false, err
	}
	// Round-trip into dest via the driver to keep dest populated with the cached form.
	if _, err := c.driver.Get(ctx, key, dest); err != nil {
		return false, err
	}
	return false, nil
}

func (c *CacheUseCase) RememberEntryMeta(ctx context.Context, siteID, entryID int64, dest any, fn func() (any, error)) error {
	return c.Remember(ctx, c.Keys.EntryMeta(siteID, entryID), c.defaultTTL, dest, fn)
}

func (c *CacheUseCase) ForgetEntryMeta(ctx context.Context, siteID, entryID int64) error {
	return c.driver.Forget(ctx, c.Keys.EntryMeta(siteID, entryID))
}
