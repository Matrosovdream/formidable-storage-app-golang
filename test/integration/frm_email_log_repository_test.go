package integration_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestFrmEmailLogRepository_UpsertAndList(t *testing.T) {
	db := helpers.DB(t)
	siteID := helpers.SeedSite(t, db)
	repo := repository.NewFrmEmailLogRepository(db)
	ctx := context.Background()

	rows := []repository.FrmEmailLogInput{
		{
			MessageID:    sql.NullString{String: "msg-1", Valid: true},
			Subject:      sql.NullString{String: "Hello", Valid: true},
			EmailFrom:    sql.NullString{String: "a@example.com", Valid: true},
			EmailTo:      sql.NullString{String: "b@example.com", Valid: true},
			ContentPlain: sql.NullString{String: "plain", Valid: true},
			ContentHTML:  sql.NullString{String: "<p>html</p>", Valid: true},
			Status:       1,
			DateSent:     sql.NullTime{Time: time.Now().UTC(), Valid: true},
		},
		{
			MessageID:    sql.NullString{String: "msg-2", Valid: true},
			Subject:      sql.NullString{String: "World", Valid: true},
			EmailTo:      sql.NullString{String: "c@example.com", Valid: true},
			ContentPlain: sql.NullString{String: "p2", Valid: true},
			ContentHTML:  sql.NullString{String: "<p>h2</p>", Valid: true},
			Status:       0,
		},
	}
	require.NoError(t, repo.UpsertMany(ctx, nil, siteID, rows))

	// Upsert again with mutated subject for msg-1 → should update, not insert.
	rows[0].Subject = sql.NullString{String: "Hello v2", Valid: true}
	require.NoError(t, repo.UpsertMany(ctx, nil, siteID, rows))

	res, err := repo.List(ctx, nil, siteID, repository.ListParams{})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)
	require.Len(t, res.Data, 2)

	// Filter by subject ILIKE
	res, err = repo.List(ctx, nil, siteID, repository.ListParams{
		Filters: map[string]any{"subject": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Total)
	require.Equal(t, "Hello v2", res.Data[0].Subject.String)

	// Filter by status
	res, err = repo.List(ctx, nil, siteID, repository.ListParams{Filters: map[string]any{"status": 0}})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Total)
}
