package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// HealthHandler returns a Fiber handler that probes DB + Redis and returns Laravel-style JSON.
func HealthHandler(db *sqlx.DB, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		checks := fiber.Map{}

		dbErr := db.PingContext(c.Context())
		checks["database"] = fiber.Map{"ok": dbErr == nil, "error": errStr(dbErr)}

		rErr := rdb.Ping(c.Context()).Err()
		checks["redis"] = fiber.Map{"ok": rErr == nil, "error": errStr(rErr)}

		body := fiber.Map{
			"status":    "ok",
			"checks":    checks,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if dbErr != nil || rErr != nil {
			body["status"] = "degraded"
			return c.Status(fiber.StatusServiceUnavailable).JSON(body)
		}
		return c.JSON(body)
	}
}

func errStr(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}
