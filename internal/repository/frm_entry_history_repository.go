package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type FrmEntryHistoryRepository struct {
	db *sqlx.DB
}

func NewFrmEntryHistoryRepository(db *sqlx.DB) *FrmEntryHistoryRepository {
	return &FrmEntryHistoryRepository{db: db}
}

func (r *FrmEntryHistoryRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const frmEntryHistoryColumns = "id, entry_id, site_id, field_id, user_id, update_type_id, old_value, new_value, change_date, created_at, updated_at"

type FrmEntryHistoryInput struct {
	EntryID      sql.NullInt64
	FieldID      int64
	UserID       sql.NullInt64
	UpdateTypeID sql.NullInt64
	OldValue     sql.NullString
	NewValue     sql.NullString
	ChangeDate   sql.NullTime
}

// InsertMany appends history rows in chunks of 200.
func (r *FrmEntryHistoryRepository) InsertMany(ctx context.Context, qu Querier, siteID int64, rows []FrmEntryHistoryInput) error {
	if len(rows) == 0 {
		return nil
	}
	q := r.q(qu)
	const chunk = 200
	for _, batch := range ChunkRows(rows, chunk) {
		args := make([]any, 0, len(batch)*9)
		placeholders := make([]string, 0, len(batch))
		pos := 1
		for _, in := range batch {
			marks := make([]string, 9)
			for i := 0; i < 9; i++ {
				marks[i] = "$" + strconv.Itoa(pos)
				pos++
			}
			placeholders = append(placeholders, "("+strings.Join(marks, ", ")+")")
			args = append(args, in.EntryID, siteID, in.FieldID, in.UserID, in.UpdateTypeID, in.OldValue, in.NewValue, in.ChangeDate, time.Now().UTC())
		}
		stmt := "INSERT INTO frm_entry_history (entry_id, site_id, field_id, user_id, update_type_id, old_value, new_value, change_date, created_at) VALUES " + strings.Join(placeholders, ", ")
		if _, err := q.ExecContext(ctx, stmt, args...); err != nil {
			return err
		}
	}
	return nil
}

func (r *FrmEntryHistoryRepository) FindByEntry(ctx context.Context, qu Querier, siteID, entryID int64) ([]entity.FrmEntryHistory, error) {
	var out []entity.FrmEntryHistory
	err := r.q(qu).SelectContext(ctx, &out,
		"SELECT "+frmEntryHistoryColumns+" FROM frm_entry_history WHERE site_id = $1 AND entry_id = $2 ORDER BY id ASC",
		siteID, entryID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

