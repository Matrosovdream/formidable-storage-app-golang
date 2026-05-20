package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestSiteRepository_CRUD(t *testing.T) {
	db := helpers.DB(t)
	repo := repository.NewSiteRepository(db)
	ctx := context.Background()

	url := fmt.Sprintf("https://site-test-%d.example/", time.Now().UnixNano())
	s := &entity.Site{Name: "Repo Site", URL: url}
	require.NoError(t, repo.Create(ctx, nil, s))
	require.Greater(t, s.ID, int64(0))
	t.Cleanup(func() { _ = repo.Delete(ctx, nil, s.ID) })

	got, err := repo.FindByID(ctx, nil, s.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "Repo Site", got.Name)
	require.Equal(t, url, got.URL)

	s.Name = "Renamed"
	require.NoError(t, repo.Update(ctx, nil, s))

	got2, err := repo.FindByID(ctx, nil, s.ID)
	require.NoError(t, err)
	require.Equal(t, "Renamed", got2.Name)
}

func TestSiteRepository_NotFound(t *testing.T) {
	db := helpers.DB(t)
	repo := repository.NewSiteRepository(db)

	got, err := repo.FindByID(context.Background(), nil, -1)
	require.NoError(t, err)
	require.Nil(t, got)
}
