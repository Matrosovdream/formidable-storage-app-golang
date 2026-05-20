package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestQueueProducer_EnqueueIncrementsStats(t *testing.T) {
	rdb := helpers.Redis(t)
	ctx := context.Background()

	queue := "test-queue-" + time.Now().Format("150405.000000")
	statsKey := "test-stats-" + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		_ = rdb.Del(ctx, queue).Err()
		_ = rdb.Del(ctx, statsKey+":99").Err()
	})

	stats := usecase.NewQueueStatsUseCase(rdb, statsKey)
	p := gw.NewQueueProducer(rdb, queue, stats)

	payload := map[string]any{"hello": "world"}
	require.NoError(t, p.Enqueue(ctx, gw.JobUpdateFrmFields, 99, payload))

	// 1 message must exist on the list.
	n, err := rdb.LLen(ctx, queue).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	// Counter must have ticked up.
	counts, _ := stats.CountsForSite(ctx, 99)
	require.Equal(t, int64(1), counts["fields"])

	// Body must round-trip into JobEnvelope.
	raw, err := rdb.RPop(ctx, queue).Bytes()
	require.NoError(t, err)
	var env gw.JobEnvelope
	require.NoError(t, json.Unmarshal(raw, &env))
	require.Equal(t, gw.JobUpdateFrmFields, env.Name)
	require.Equal(t, int64(99), env.SiteID)

	var got map[string]any
	require.NoError(t, json.Unmarshal(env.Payload, &got))
	require.Equal(t, "world", got["hello"])
}
