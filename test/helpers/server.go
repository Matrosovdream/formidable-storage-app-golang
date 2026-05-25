package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/app"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	apphttp "github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/route"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// TestServer is a fully-wired Fiber app backed by the live dev Postgres + a unique Redis logical DB.
type TestServer struct {
	App   *fiber.App
	DB    *sqlx.DB
	Redis *redis.Client
	Deps  *app.Deps
}

// NewTestServer builds a Fiber app with all routes registered.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	cfg := buildTestConfig(t)
	deps, err := app.Build(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(deps.Close)

	a := fiber.New(fiber.Config{
		AppName:      "fsa-test",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: apphttp.ErrorHandler(deps.Log, true),
	})
	a.Get("/health", apphttp.HealthHandler(deps.DB, deps.Redis))

	ctrl := route.Controllers{
		Auth:                    apphttp.NewAuthController(deps.Auth),
		Site:                    apphttp.NewSiteController(deps.Site, deps.FrmEmailLogUC, deps.FrmField, deps.FrmEntryHistoryUC),
		SiteGenerate:            apphttp.NewSiteGenerateController(deps.SiteGenerate),
		Data:                    apphttp.NewDataController(deps.Data),
		RestV1EntryHistory:      apphttp.NewRestV1EntryHistoryController(deps.FrmEntryHistoryUC, deps.QueueProducer),
		RestV1Fields:            apphttp.NewRestV1FieldsController(deps.QueueProducer),
		RestV1EmailsLog:         apphttp.NewRestV1EmailsLogController(deps.FrmEmailLogUC, deps.QueueProducer),
		RestV1EpShipmentHistory: apphttp.NewRestV1EpShipmentHistoryController(deps.FrmEpShipmentHistoryUC, deps.QueueProducer),
	}
	mw := route.Middleware{
		AuthSanctum:         middleware.AuthSanctum(deps.Auth),
		OptionalAuthSanctum: middleware.OptionalAuthSanctum(deps.Auth),
		RestToken:           middleware.RestToken(deps.Sites, deps.SiteTokens),
	}
	route.Register(a, ctrl, mw)

	return &TestServer{App: a, DB: deps.DB, Redis: deps.Redis, Deps: deps}
}

func buildTestConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.Name = "fsa-test"
	cfg.App.Env = "test"
	cfg.App.Debug = true
	cfg.App.Port = 0
	cfg.DB.Host = envOr("TEST_DB_HOST", "localhost")
	cfg.DB.Port = atoi(envOr("TEST_DB_PORT", "5432"))
	cfg.DB.Database = envOr("TEST_DB_DATABASE", "app")
	cfg.DB.Username = envOr("TEST_DB_USERNAME", "laravel")
	cfg.DB.Password = envOr("TEST_DB_PASSWORD", "secret")
	cfg.DB.SSLMode = envOr("TEST_DB_SSLMODE", "disable")
	cfg.DB.MaxOpenConns = 5
	cfg.DB.MaxIdleConns = 2
	cfg.Redis.Host = envOr("TEST_REDIS_HOST", "localhost")
	cfg.Redis.Port = atoi(envOr("TEST_REDIS_PORT", "6379"))
	cfg.Redis.DB = 2
	cfg.Cache.Driver = "memory"
	cfg.Cache.TTL = 60
	cfg.Cache.Prefix = "test:"
	cfg.Queue.Connection = "redis"
	cfg.Queue.Stream = "test-queues:" + t.Name()
	cfg.Queue.StatsKey = "test-stats:" + t.Name()
	cfg.Auth.BcryptCost = 4 // fast in tests
	cfg.Auth.TokenLifetimeMinutes = 0
	cfg.Log.Level = "warn"
	return cfg
}

// Do issues a request through Fiber's in-memory test transport and returns the response.
func (s *TestServer) Do(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, rdr)
	if rdr != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.App.Test(req, -1)
	require.NoError(t, err)
	return resp
}

// DecodeJSON drains and parses an HTTP response body as JSON into dest.
func DecodeJSON(t *testing.T, resp *http.Response, dest any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(dest))
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}
