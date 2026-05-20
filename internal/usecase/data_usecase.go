package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/jmoiron/sqlx"
)

type DataUseCase struct {
	db         *sqlx.DB
	history    *FrmEntryHistoryUseCase
	emails     *FrmEmailLogUseCase
}

func NewDataUseCase(db *sqlx.DB, history *FrmEntryHistoryUseCase, emails *FrmEmailLogUseCase) *DataUseCase {
	return &DataUseCase{db: db, history: history, emails: emails}
}

// ListEntries returns a paginated, joined view of per-entry summaries scoped to siteID.
func (u *DataUseCase) ListEntries(ctx context.Context, siteID int64, req model.DataEntriesRequest) (repository.ListResult[model.DataEntryRow], error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 25
	}
	if perPage > 200 {
		perPage = 200
	}

	sortCol := "last_update"
	switch req.SortBy {
	case "entry_id", "last_update", "email_count", "update_count":
		sortCol = req.SortBy
	}
	sortDir := "desc"
	if req.SortDir == "asc" {
		sortDir = "asc"
	}

	var entryFilter sql.NullInt64
	if req.EntryID != nil {
		entryFilter = sql.NullInt64{Int64: *req.EntryID, Valid: true}
	}

	const entriesCTE = `
		WITH entries AS (
			SELECT entry_id FROM frm_entry_history WHERE site_id = $1 AND entry_id IS NOT NULL
			UNION
			SELECT entry_id FROM frm_emails_log    WHERE site_id = $1 AND entry_id IS NOT NULL
		),
		summary AS (
			SELECT
				e.entry_id,
				$1::bigint AS site_id,
				(SELECT MAX(change_date) FROM frm_entry_history WHERE site_id = $1 AND entry_id = e.entry_id) AS last_update,
				(SELECT COUNT(*) FROM frm_emails_log    WHERE site_id = $1 AND entry_id = e.entry_id)::bigint AS email_count,
				(SELECT COUNT(*) FROM frm_entry_history WHERE site_id = $1 AND entry_id = e.entry_id)::bigint AS update_count
			FROM entries e
		)
		SELECT entry_id, site_id, last_update, email_count, update_count
		FROM summary
		WHERE ($2::bigint IS NULL OR entry_id = $2)
	`

	// total
	var total int64
	if err := u.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM ("+entriesCTE+") sub",
		siteID, entryFilter,
	); err != nil {
		return repository.ListResult[model.DataEntryRow]{}, err
	}

	// page
	stmt := entriesCTE + " ORDER BY " + sortCol + " " + sortDir + " NULLS LAST LIMIT $3 OFFSET $4"
	offset := (page - 1) * perPage

	type row struct {
		EntryID     int64        `db:"entry_id"`
		SiteID      int64        `db:"site_id"`
		LastUpdate  sql.NullTime `db:"last_update"`
		EmailCount  int64        `db:"email_count"`
		UpdateCount int64        `db:"update_count"`
	}
	var rs []row
	if err := u.db.SelectContext(ctx, &rs, stmt, siteID, entryFilter, perPage, offset); err != nil {
		return repository.ListResult[model.DataEntryRow]{}, err
	}

	data := make([]model.DataEntryRow, len(rs))
	for i, r := range rs {
		var lp *time.Time
		if r.LastUpdate.Valid {
			t := r.LastUpdate.Time
			lp = &t
		}
		data[i] = model.DataEntryRow{
			EntryID:     r.EntryID,
			SiteID:      r.SiteID,
			LastUpdate:  lp,
			EmailCount:  r.EmailCount,
			UpdateCount: r.UpdateCount,
		}
	}

	last := int((total + int64(perPage) - 1) / int64(perPage))
	if last < 1 {
		last = 1
	}
	return repository.ListResult[model.DataEntryRow]{Data: data, Total: total, Page: page, PerPage: perPage, LastPage: last}, nil
}

func (u *DataUseCase) EntryUpdates(ctx context.Context, siteID, entryID int64) (model.EntryUpdatesResponse, error) {
	return u.history.GetByEntry(ctx, siteID, entryID)
}

func (u *DataUseCase) EntryEmails(ctx context.Context, siteID, entryID int64) ([]model.EmailLogItem, error) {
	return u.emails.FindByEntry(ctx, siteID, entryID)
}
