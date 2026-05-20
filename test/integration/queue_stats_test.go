package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestQueueStats_IncrementAndCounts(t *testing.T) {
	rdb := helpers.Redis(t)
	// Use a key unique to this test to avoid collisions across runs.
	key := "test-queue-stats-" + time.Now().Format("150405.000000")
	q := usecase.NewQueueStatsUseCase(rdb, key)
	siteID := int64(42)
	ctx := context.Background()
	t.Cleanup(func() { _ = rdb.Del(ctx, key+":42").Err() })

	require.NoError(t, q.Increment(ctx, siteID, "fields"))
	require.NoError(t, q.Increment(ctx, siteID, "fields"))
	require.NoError(t, q.Increment(ctx, siteID, "emails"))

	counts, err := q.CountsForSite(ctx, siteID)
	require.NoError(t, err)
	require.Equal(t, int64(2), counts["fields"])
	require.Equal(t, int64(1), counts["emails"])
	require.Equal(t, int64(0), counts["entry_history"])

	require.NoError(t, q.Decrement(ctx, siteID, "fields"))
	counts, _ = q.CountsForSite(ctx, siteID)
	require.Equal(t, int64(1), counts["fields"])
}
