// Package app wires the dependency graph used by both cmd/web and cmd/worker.
package app

import (
	"context"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/cache"
	gwmsg "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Deps struct {
	Cfg      *config.Config
	Log      *logrus.Logger
	DB       *sqlx.DB
	Redis    *redis.Client
	Validate *validator.Validate

	// Repositories
	Users                 *repository.UserRepository
	Sites                 *repository.SiteRepository
	SiteTokens            *repository.SiteTokenRepository
	PersonalAccessTokens  *repository.PersonalAccessTokenRepository
	FrmFields             *repository.FrmFieldRepository
	FrmEntryHistory       *repository.FrmEntryHistoryRepository
	FrmEntryUpdateTypes   *repository.FrmEntryUpdateTypeRepository
	FrmEmailLog           *repository.FrmEmailLogRepository
	FrmEpShipment         *repository.FrmEpShipmentRepository
	FrmEpShipmentHistory  *repository.FrmEpShipmentHistoryRepository

	// Cache
	Cache *usecase.CacheUseCase

	// Use cases
	Auth                    *usecase.AuthUseCase
	Site                    *usecase.SiteUseCase
	FrmField                *usecase.FrmFieldUseCase
	FrmEntryHistoryUC       *usecase.FrmEntryHistoryUseCase
	FrmEmailLogUC           *usecase.FrmEmailLogUseCase
	FrmEpShipmentHistoryUC  *usecase.FrmEpShipmentHistoryUseCase
	Data                    *usecase.DataUseCase
	QueueStats              *usecase.QueueStatsUseCase

	// Messaging
	QueueProducer *gwmsg.QueueProducer
}

func Build(ctx context.Context, cfg *config.Config) (*Deps, error) {
	log := config.NewLogger(cfg.Log)

	db, err := config.OpenDB(ctx, cfg.DB)
	if err != nil {
		return nil, err
	}
	rdb, err := config.OpenRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}
	validate := config.NewValidator()

	// Repos
	users := repository.NewUserRepository(db)
	sites := repository.NewSiteRepository(db)
	siteTokens := repository.NewSiteTokenRepository(db)
	pats := repository.NewPersonalAccessTokenRepository(db)
	frmFields := repository.NewFrmFieldRepository(db)
	frmEntryHistory := repository.NewFrmEntryHistoryRepository(db)
	frmEntryUpdateTypes := repository.NewFrmEntryUpdateTypeRepository(db)
	frmEmailLog := repository.NewFrmEmailLogRepository(db)
	frmEpShipment := repository.NewFrmEpShipmentRepository(db)
	frmEpShipmentHistory := repository.NewFrmEpShipmentHistoryRepository(db)

	// Cache
	var cacheDriver cache.Driver
	if cfg.Cache.Driver == "memory" {
		cacheDriver = cache.NewMemoryDriver()
	} else {
		cacheDriver = cache.NewRedisDriver(rdb)
	}
	cacheKeys := cache.NewKeyBuilder(cfg.Cache.Prefix)
	cacheTTL := time.Duration(cfg.Cache.TTL) * time.Second
	cacheUC := usecase.NewCacheUseCase(cacheDriver, cacheKeys, cacheTTL)

	// Queue stats
	queueStats := usecase.NewQueueStatsUseCase(rdb, cfg.Queue.StatsKey)
	queueProducer := gwmsg.NewQueueProducer(rdb, cfg.Queue.Stream, queueStats)

	// Use cases
	authUC := usecase.NewAuthUseCase(db, validate, users, pats, cfg.Auth.BcryptCost, cfg.Auth.TokenLifetimeMinutes)
	siteUC := usecase.NewSiteUseCase(db, validate, sites, siteTokens, queueStats)
	fieldUC := usecase.NewFrmFieldUseCase(frmFields, cacheUC)
	historyUC := usecase.NewFrmEntryHistoryUseCase(db, frmEntryHistory, frmFields, fieldUC, frmEntryUpdateTypes, cacheUC)
	emailUC := usecase.NewFrmEmailLogUseCase(frmEmailLog)
	epHistoryUC := usecase.NewFrmEpShipmentHistoryUseCase(frmEpShipmentHistory)
	dataUC := usecase.NewDataUseCase(db, historyUC, emailUC)

	return &Deps{
		Cfg:      cfg,
		Log:      log,
		DB:       db,
		Redis:    rdb,
		Validate: validate,

		Users:                 users,
		Sites:                 sites,
		SiteTokens:            siteTokens,
		PersonalAccessTokens:  pats,
		FrmFields:             frmFields,
		FrmEntryHistory:       frmEntryHistory,
		FrmEntryUpdateTypes:   frmEntryUpdateTypes,
		FrmEmailLog:           frmEmailLog,
		FrmEpShipment:         frmEpShipment,
		FrmEpShipmentHistory:  frmEpShipmentHistory,

		Cache: cacheUC,

		Auth:                    authUC,
		Site:                    siteUC,
		FrmField:                fieldUC,
		FrmEntryHistoryUC:       historyUC,
		FrmEmailLogUC:           emailUC,
		FrmEpShipmentHistoryUC:  epHistoryUC,
		Data:                    dataUC,
		QueueStats:              queueStats,

		QueueProducer: queueProducer,
	}, nil
}

func (d *Deps) Close() {
	if d.DB != nil {
		_ = d.DB.Close()
	}
	if d.Redis != nil {
		_ = d.Redis.Close()
	}
}
