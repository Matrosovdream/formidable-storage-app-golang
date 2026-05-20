package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/cache"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/stretchr/testify/require"
)

func TestMemoryDriver_PutGetForget(t *testing.T) {
	d := cache.NewMemoryDriver()
	ctx := context.Background()

	type payload struct{ V int }

	require.NoError(t, d.Put(ctx, "k", payload{V: 7}, time.Minute))
	var got payload
	found, err := d.Get(ctx, "k", &got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 7, got.V)

	require.NoError(t, d.Forget(ctx, "k"))
	found, _ = d.Get(ctx, "k", &got)
	require.False(t, found)
}

func TestCacheUseCase_RememberCallsFnOnlyOnce(t *testing.T) {
	d := cache.NewMemoryDriver()
	kb := cache.NewKeyBuilder("t:")
	uc := usecase.NewCacheUseCase(d, kb, time.Minute)

	calls := 0
	build := func() (any, error) {
		calls++
		return map[string]int{"a": 1}, nil
	}

	var first, second map[string]int
	require.NoError(t, uc.Remember(context.Background(), "key", time.Minute, &first, build))
	require.NoError(t, uc.Remember(context.Background(), "key", time.Minute, &second, build))
	require.Equal(t, 1, calls)
	require.Equal(t, first, second)
	require.Equal(t, 1, first["a"])
}

func TestCacheUseCase_RememberTracked_HitMiss(t *testing.T) {
	d := cache.NewMemoryDriver()
	kb := cache.NewKeyBuilder("t:")
	uc := usecase.NewCacheUseCase(d, kb, time.Minute)

	var v map[string]int
	hit, err := uc.RememberTracked(context.Background(), "k", time.Minute, &v, func() (any, error) {
		return map[string]int{"x": 1}, nil
	})
	require.NoError(t, err)
	require.False(t, hit)

	var v2 map[string]int
	hit, err = uc.RememberTracked(context.Background(), "k", time.Minute, &v2, func() (any, error) {
		return map[string]int{"x": 99}, nil
	})
	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, 1, v2["x"]) // came from cache, not the new fn
}
