package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type QueueStatsIncrementer interface {
	Increment(ctx context.Context, siteID int64, t string) error
}

type QueueProducer struct {
	rdb   *redis.Client
	queue string
	stats QueueStatsIncrementer
}

func NewQueueProducer(rdb *redis.Client, queue string, stats QueueStatsIncrementer) *QueueProducer {
	if queue == "" {
		queue = "queues:default"
	}
	return &QueueProducer{rdb: rdb, queue: queue, stats: stats}
}

func (p *QueueProducer) Enqueue(ctx context.Context, name string, siteID int64, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	env := JobEnvelope{
		ID:         uuid.NewString(),
		Name:       name,
		SiteID:     siteID,
		Payload:    body,
		Attempts:   0,
		EnqueuedAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	if err := p.rdb.LPush(ctx, p.queue, raw).Err(); err != nil {
		return err
	}
	if p.stats != nil {
		_ = p.stats.Increment(ctx, siteID, StatTypeFor(name))
	}
	return nil
}
