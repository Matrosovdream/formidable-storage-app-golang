package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/app"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	apphttp "github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/route"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/docs"
	"github.com/gofiber/fiber/v2"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	deps, err := app.Build(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer deps.Close()

	a := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		ErrorHandler: apphttp.ErrorHandler(deps.Log, cfg.App.Debug),
	})

	a.Get("/health", apphttp.HealthHandler(deps.DB, deps.Redis))
	docs.Register(a)

	ctrl := route.Controllers{
		Auth:                    apphttp.NewAuthController(deps.Auth),
		Site:                    apphttp.NewSiteController(deps.Site),
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

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	deps.Log.Infof("web listening on %s", addr)
	if err := a.Listen(addr); err != nil {
		deps.Log.WithError(err).Fatal("listen failed")
	}
}
