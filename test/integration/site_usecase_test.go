package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestSiteUseCase_CreateProducesTokenAndStats(t *testing.T) {
	db := helpers.DB(t)
	rdb := helpers.Redis(t)
	v := config.NewValidator()

	sites := repository.NewSiteRepository(db)
	siteTokens := repository.NewSiteTokenRepository(db)
	queueStats := usecase.NewQueueStatsUseCase(rdb, "test-stats-"+time.Now().Format("150405.000000"))
	uc := usecase.NewSiteUseCase(db, v, sites, siteTokens, queueStats)

	url := fmt.Sprintf("https://uc-test-%d.example/", time.Now().UnixNano())
	resp, err := uc.Create(context.Background(), model.CreateSiteRequest{Name: "UC Site", URL: url})
	require.NoError(t, err)
	require.Greater(t, resp.ID, int64(0))
	require.NotEmpty(t, resp.Token)
	require.NotNil(t, resp.Stats.Queue)
	t.Cleanup(func() { _ = sites.Delete(context.Background(), nil, resp.ID) })

	view, err := uc.View(context.Background(), resp.ID)
	require.NoError(t, err)
	require.Equal(t, url, view.URL)
	require.Equal(t, resp.Token, view.Token)
}

func TestSiteUseCase_CreateValidationError(t *testing.T) {
	db := helpers.DB(t)
	rdb := helpers.Redis(t)
	v := config.NewValidator()

	sites := repository.NewSiteRepository(db)
	siteTokens := repository.NewSiteTokenRepository(db)
	queueStats := usecase.NewQueueStatsUseCase(rdb, "test-stats-"+time.Now().Format("150405.000000"))
	uc := usecase.NewSiteUseCase(db, v, sites, siteTokens, queueStats)

	_, err := uc.Create(context.Background(), model.CreateSiteRequest{Name: "", URL: "not-a-url"})
	require.Error(t, err)
	_, ok := err.(*usecase.ValidationError)
	require.True(t, ok, "expected ValidationError, got %T: %v", err, err)
}
