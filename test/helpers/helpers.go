// Package helpers provides test harness utilities backed by the live dev DB + Redis.
package helpers

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *sqlx.DB {
	t.Helper()
	host := envOr("TEST_DB_HOST", "localhost")
	portS := envOr("TEST_DB_PORT", "5432")
	port, _ := strconv.Atoi(portS)
	cfg := config.DBConfig{
		Host:     host,
		Port:     port,
		Database: envOr("TEST_DB_DATABASE", "app"),
		Username: envOr("TEST_DB_USERNAME", "laravel"),
		Password: envOr("TEST_DB_PASSWORD", "secret"),
		SSLMode:  envOr("TEST_DB_SSLMODE", "disable"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := config.OpenDB(ctx, cfg)
	require.NoError(t, err, "open test db")
	return db
}

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	host := envOr("TEST_REDIS_HOST", "localhost")
	portS := envOr("TEST_REDIS_PORT", "6379")
	port, _ := strconv.Atoi(portS)
	cfg := config.RedisConfig{Host: host, Port: port, DB: 1}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := config.OpenRedis(ctx, cfg)
	require.NoError(t, err, "open test redis")
	return c
}

// DB returns a *sqlx.DB closed at end of test.
func DB(t *testing.T) *sqlx.DB {
	db := testDB(t)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// Redis returns a *redis.Client with a per-test FLUSHDB cleanup on a dedicated DB index.
func Redis(t *testing.T) *redis.Client {
	c := testRedis(t)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

var siteCounter uint64

// SeedSite creates a site with a unique URL and returns it. Site is deleted on test cleanup.
func SeedSite(t *testing.T, db *sqlx.DB) (siteID int64) {
	t.Helper()
	n := atomic.AddUint64(&siteCounter, 1)
	url := fmt.Sprintf("https://test-%d-%d.example/", time.Now().UnixNano(), rand.Int())
	_ = n
	const stmt = "INSERT INTO sites (name, url) VALUES ($1, $2) RETURNING id"
	require.NoError(t, db.QueryRowx(stmt, "Test Site", url).Scan(&siteID))
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sites WHERE id = $1", siteID)
	})
	return siteID
}

// UniqueTokenPrefix returns a per-test key prefix usable for Redis isolation.
func UniqueTokenPrefix(t *testing.T) string {
	return fmt.Sprintf("test:%s:%d:", t.Name(), time.Now().UnixNano())
}

func envOr(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}
