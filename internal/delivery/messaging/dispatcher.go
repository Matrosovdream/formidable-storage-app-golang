package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Handler runs a job. Return nil for success, error for failure (will be retried until maxAttempts).
type Handler func(ctx context.Context, env gw.JobEnvelope) error

type QueueStatsDecrementer interface {
	Decrement(ctx context.Context, siteID int64, t string) error
}

type Dispatcher struct {
	rdb         *redis.Client
	queue       string
	failedQueue string
	maxAttempts int
	handlers    map[string]Handler
	stats       QueueStatsDecrementer
	log         *logrus.Logger
}

func NewDispatcher(rdb *redis.Client, queue string, stats QueueStatsDecrementer, log *logrus.Logger) *Dispatcher {
	if queue == "" {
		queue = "queues:default"
	}
	return &Dispatcher{
		rdb:         rdb,
		queue:       queue,
		failedQueue: queue + ":failed",
		maxAttempts: 3,
		handlers:    make(map[string]Handler),
		stats:       stats,
		log:         log,
	}
}

func (d *Dispatcher) Register(name string, h Handler) {
	d.handlers[name] = h
}

// Run consumes jobs until ctx is cancelled.
func (d *Dispatcher) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		res, err := d.rdb.BRPop(ctx, 5*time.Second, d.queue).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				continue
			}
			d.log.WithError(err).Warn("queue poll failed; backing off")
			time.Sleep(time.Second)
			continue
		}
		if len(res) < 2 {
			continue
		}

		var env gw.JobEnvelope
		if err := json.Unmarshal([]byte(res[1]), &env); err != nil {
			d.log.WithError(err).Error("invalid job envelope; dropping")
			continue
		}

		if err := d.dispatch(ctx, env); err != nil {
			env.Attempts++
			if env.Attempts < d.maxAttempts {
				d.log.WithFields(logrus.Fields{"job": env.Name, "id": env.ID, "attempts": env.Attempts}).Warn("retrying job")
				raw, _ := json.Marshal(env)
				_ = d.rdb.LPush(ctx, d.queue, raw).Err()
				continue
			}
			d.log.WithError(err).WithFields(logrus.Fields{"job": env.Name, "id": env.ID}).Error("job failed permanently")
			raw, _ := json.Marshal(env)
			_ = d.rdb.LPush(ctx, d.failedQueue, raw).Err()
			d.decrement(ctx, env)
			continue
		}
		d.decrement(ctx, env)
	}
}

func (d *Dispatcher) dispatch(ctx context.Context, env gw.JobEnvelope) error {
	h, ok := d.handlers[env.Name]
	if !ok {
		return fmt.Errorf("no handler for job %q", env.Name)
	}
	return h(ctx, env)
}

func (d *Dispatcher) decrement(ctx context.Context, env gw.JobEnvelope) {
	if d.stats == nil {
		return
	}
	_ = d.stats.Decrement(ctx, env.SiteID, gw.StatTypeFor(env.Name))
}
