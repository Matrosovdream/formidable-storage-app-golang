package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDriver struct {
	client *redis.Client
}

func NewRedisDriver(client *redis.Client) *RedisDriver {
	return &RedisDriver{client: client}
}

func (d *RedisDriver) Get(ctx context.Context, key string, dest any) (bool, error) {
	raw, err := d.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if dest == nil {
		return true, nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (d *RedisDriver) Put(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return d.client.Set(ctx, key, raw, ttl).Err()
}

func (d *RedisDriver) Forget(ctx context.Context, key string) error {
	return d.client.Del(ctx, key).Err()
}

func (d *RedisDriver) Has(ctx context.Context, key string) (bool, error) {
	n, err := d.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (d *RedisDriver) Name() string { return "redis" }
