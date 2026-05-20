package usecase

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	QueueStatFields       = "fields"
	QueueStatEmails       = "emails"
	QueueStatEntryHistory = "entry_history"
	QueueStatShipmentHist = "shipment_history"
)

type QueueStatsUseCase struct {
	rdb     *redis.Client
	keyBase string
}

func NewQueueStatsUseCase(rdb *redis.Client, keyBase string) *QueueStatsUseCase {
	if keyBase == "" {
		keyBase = "queue-stats"
	}
	return &QueueStatsUseCase{rdb: rdb, keyBase: keyBase}
}

func (q *QueueStatsUseCase) key(siteID int64) string {
	return q.keyBase + ":" + strconv.FormatInt(siteID, 10)
}

func (q *QueueStatsUseCase) Increment(ctx context.Context, siteID int64, t string) error {
	return q.rdb.HIncrBy(ctx, q.key(siteID), t, 1).Err()
}

func (q *QueueStatsUseCase) Decrement(ctx context.Context, siteID int64, t string) error {
	return q.rdb.HIncrBy(ctx, q.key(siteID), t, -1).Err()
}

func (q *QueueStatsUseCase) CountsForSite(ctx context.Context, siteID int64) (map[string]int64, error) {
	out := map[string]int64{
		QueueStatFields:       0,
		QueueStatEmails:       0,
		QueueStatEntryHistory: 0,
		QueueStatShipmentHist: 0,
	}
	raw, err := q.rdb.HGetAll(ctx, q.key(siteID)).Result()
	if err != nil {
		return out, err
	}
	for k, v := range raw {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			out[k] = n
		}
	}
	return out, nil
}
