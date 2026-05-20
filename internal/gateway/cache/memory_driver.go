package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type memoryEntry struct {
	value     []byte
	expiresAt time.Time
}

type MemoryDriver struct {
	mu    sync.RWMutex
	items map[string]memoryEntry
}

func NewMemoryDriver() *MemoryDriver {
	return &MemoryDriver{items: make(map[string]memoryEntry)}
}

func (d *MemoryDriver) Get(_ context.Context, key string, dest any) (bool, error) {
	d.mu.RLock()
	e, ok := d.items[key]
	d.mu.RUnlock()
	if !ok {
		return false, nil
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		d.mu.Lock()
		delete(d.items, key)
		d.mu.Unlock()
		return false, nil
	}
	if dest == nil {
		return true, nil
	}
	if err := json.Unmarshal(e.value, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (d *MemoryDriver) Put(_ context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	d.mu.Lock()
	d.items[key] = memoryEntry{value: raw, expiresAt: exp}
	d.mu.Unlock()
	return nil
}

func (d *MemoryDriver) Forget(_ context.Context, key string) error {
	d.mu.Lock()
	delete(d.items, key)
	d.mu.Unlock()
	return nil
}

func (d *MemoryDriver) Has(ctx context.Context, key string) (bool, error) {
	ok, err := d.Get(ctx, key, nil)
	return ok, err
}

func (d *MemoryDriver) Name() string { return "memory" }
