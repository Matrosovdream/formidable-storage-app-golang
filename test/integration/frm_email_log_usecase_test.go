package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestFrmEmailLogUseCase_UpdateAndList(t *testing.T) {
	db := helpers.DB(t)
	siteID := helpers.SeedSite(t, db)
	uc := usecase.NewFrmEmailLogUseCase(repository.NewFrmEmailLogRepository(db))

	now := time.Now().UTC().Format(time.RFC3339)
	mk := func(i int) model.FrmEmailLogInput {
		mid := fmt.Sprintf("uc-msg-%d", i)
		sub := fmt.Sprintf("Subject %d", i)
		from := "noreply@example.com"
		to := fmt.Sprintf("dest%d@example.com", i)
		cp := "plain"
		ch := "<p>html</p>"
		ds := now
		return model.FrmEmailLogInput{
			MessageID:    &mid,
			Subject:      &sub,
			EmailFrom:    &from,
			EmailTo:      &to,
			ContentPlain: &cp,
			ContentHTML:  &ch,
			Status:       int16(i % 3),
			DateSent:     &ds,
		}
	}

	rows := []model.FrmEmailLogInput{mk(1), mk(2), mk(3)}
	n, err := uc.UpdateAll(context.Background(), siteID, rows)
	require.NoError(t, err)
	require.Equal(t, 3, n)

	listed, err := uc.List(context.Background(), siteID, model.ListEmailsLogRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(3), listed.Total)
	require.Len(t, listed.Data, 3)
}
