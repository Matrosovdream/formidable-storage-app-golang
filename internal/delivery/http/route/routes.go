package route

import (
	"time"

	apphttp "github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	"github.com/gofiber/fiber/v2"
)

// Controllers bundles all HTTP controllers so route registration stays a pure function.
type Controllers struct {
	Auth                   *apphttp.AuthController
	Site                   *apphttp.SiteController
	SiteGenerate           *apphttp.SiteGenerateController
	Data                   *apphttp.DataController
	RestV1EntryHistory     *apphttp.RestV1EntryHistoryController
	RestV1Fields           *apphttp.RestV1FieldsController
	RestV1EmailsLog        *apphttp.RestV1EmailsLogController
	RestV1EpShipmentHistory *apphttp.RestV1EpShipmentHistoryController
}

// Middleware bundles the per-app middleware constructors.
type Middleware struct {
	AuthSanctum         fiber.Handler
	OptionalAuthSanctum fiber.Handler
	RestToken           fiber.Handler
}

// Register wires every route. Health is registered by the caller (cmd/web).
func Register(app *fiber.App, ctrl Controllers, mw Middleware) {
	api := app.Group("/api")
	api.Post("/login", ctrl.Auth.Login)
	api.Post("/register", ctrl.Auth.Register)
	api.Get("/user", mw.OptionalAuthSanctum, ctrl.Auth.Me)
	api.Post("/logout", mw.AuthSanctum, ctrl.Auth.Logout)

	sites := api.Group("/sites", mw.AuthSanctum)
	sites.Get("/list", ctrl.Site.List).Name("api-sites-list")
	sites.Get("/view/:site_id", ctrl.Site.View).Name("api-sites-view")
	sites.Get("/view/:site_id/emails", ctrl.Site.ViewEmails).Name("api-sites-view-emails")
	sites.Get("/view/:site_id/fields", ctrl.Site.ViewFields).Name("api-sites-view-fields")
	sites.Get("/view/:site_id/entry-updates", ctrl.Site.ViewEntryUpdates).Name("api-sites-view-entry-updates")
	sites.Get("/create", ctrl.Site.Create).Name("api-sites-create")
	sites.Post("/store", ctrl.Site.Store).Name("api-sites-store")
	sites.Delete("/delete/:site_id", ctrl.Site.Delete).Name("api-sites-delete")

	sites.Post("/generate/:site_id/emails", ctrl.SiteGenerate.Emails).Name("api-sites-generate-emails")
	sites.Post("/generate/:site_id/fields", ctrl.SiteGenerate.Fields).Name("api-sites-generate-fields")
	sites.Post("/generate/:site_id/entry-updates", ctrl.SiteGenerate.EntryUpdates).Name("api-sites-generate-entry-updates")

	data := api.Group("/data", mw.AuthSanctum)
	data.Get("/entries/:site_id", ctrl.Data.Entries).Name("api-data-entries")
	data.Get("/entries/:site_id/:entry_id/updates", ctrl.Data.EntryUpdates).Name("api-data-entry-updates")
	data.Get("/entries/:site_id/:entry_id/emails", ctrl.Data.EntryEmails).Name("api-data-entry-emails")

	// REST v1 (server-to-server) — gated by site bearer token.
	rest := app.Group("/rest/v1", mw.RestToken)
	rest.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"version":   "1.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	eh := rest.Group("/entry/history")
	eh.Post("/update", ctrl.RestV1EntryHistory.Update)
	eh.Post("/view/:id", ctrl.RestV1EntryHistory.GetEntryHistory)

	f := rest.Group("/fields")
	f.Post("/update-all", ctrl.RestV1Fields.UpdateAll)

	el := rest.Group("/emailslog")
	el.Post("/update-all", ctrl.RestV1EmailsLog.UpdateAll)
	el.Post("/update-all/raw", ctrl.RestV1EmailsLog.UpdateAllRaw)
	el.Post("/list", ctrl.RestV1EmailsLog.List)

	ep := rest.Group("/ep-shipment-history")
	ep.Post("/update-all", ctrl.RestV1EpShipmentHistory.UpdateAll)
	ep.Post("/list", ctrl.RestV1EpShipmentHistory.List)
}

// Use re-exports middleware so the route layer can refer to ctx helpers without circular import.
var (
	_ = middleware.SiteFromCtx
	_ = middleware.UserFromCtx
)
