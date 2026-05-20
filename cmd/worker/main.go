package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/app"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	delivery "github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/messaging"
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

	disp := delivery.NewDispatcher(deps.Redis, cfg.Queue.Stream, deps.QueueStats, deps.Log)
	delivery.RegisterHandlers(disp, delivery.HandlerDeps{
		Fields:           deps.FrmField,
		EntryHistory:     deps.FrmEntryHistoryUC,
		EmailLog:         deps.FrmEmailLogUC,
		ShipmentHistory:  deps.FrmEpShipmentHistoryUC,
	})

	deps.Log.Info("worker started")
	if err := disp.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		deps.Log.WithError(err).Error("worker exited")
	}
	deps.Log.Info("worker stopped")
}
