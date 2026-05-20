// Package docs serves the OpenAPI spec + Swagger UI.
//
// Two HTTP routes are exposed by Handlers():
//
//	GET /openapi.yaml  → the raw OpenAPI 3.1 spec
//	GET /docs          → Swagger UI loading the spec above
//
// Both assets are embedded so the docs work from the distroless prod image
// without any source files on disk.
package docs

import (
	"embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed openapi.yaml docs.html
var assets embed.FS

// Register attaches the docs routes to the Fiber app.
func Register(app *fiber.App) {
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		raw, err := assets.ReadFile("openapi.yaml")
		if err != nil {
			return err
		}
		c.Type("yaml")
		return c.Send(raw)
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		raw, err := assets.ReadFile("docs.html")
		if err != nil {
			return err
		}
		c.Type("html")
		return c.Send(raw)
	})
}
