package repository

import (
	"context"
	"database/sql"
	"fmt"
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

// List performs filter/sort/pagination over frm_entry_history scoped to siteID.
// Filter keys (whitelisted): entry_id, field_id, user_id, update_type_id (=);
// change_date_from, change_date_to (range).
// Sort whitelist: id, entry_id, field_id, change_date, created_at.
func (r *FrmEntryHistoryRepository) List(ctx context.Context, qu Querier, siteID int64, p ListParams) (ListResult[entity.FrmEntryHistory], error) {
	p.Normalize(25, 200)

	where := []string{"site_id = $1"}
	args := []any{siteID}
	pos := 2

	eqKeys := map[string]bool{"entry_id": true, "field_id": true, "user_id": true, "update_type_id": true}

	for k, v := range p.Filters {
		switch {
		case eqKeys[k]:
			where = append(where, fmt.Sprintf("%s = $%d", k, pos))
			args = append(args, v)
			pos++
		case k == "change_date_from":
			where = append(where, fmt.Sprintf("change_date >= $%d", pos))
			args = append(args, v)
			pos++
		case k == "change_date_to":
			where = append(where, fmt.Sprintf("change_date <= $%d", pos))
			args = append(args, v)
			pos++
		}
	}

	sortCol := "id"
	switch p.SortBy {
	case "id", "entry_id", "field_id", "change_date", "created_at":
		sortCol = p.SortBy
	}

	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.q(qu).GetContext(ctx, &total, "SELECT COUNT(*) FROM frm_entry_history WHERE "+whereSQL, args...); err != nil {
		return ListResult[entity.FrmEntryHistory]{}, err
	}

	listArgs := append(args, p.PerPage, p.Offset())
	stmt := fmt.Sprintf(
		"SELECT %s FROM frm_entry_history WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		frmEntryHistoryColumns, whereSQL, sortCol, p.SortDir, pos, pos+1,
	)

	var data []entity.FrmEntryHistory
	if err := r.q(qu).SelectContext(ctx, &data, stmt, listArgs...); err != nil {
		return ListResult[entity.FrmEntryHistory]{}, err
	}

	last := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if last < 1 {
		last = 1
	}
	return ListResult[entity.FrmEntryHistory]{Data: data, Total: total, Page: p.Page, PerPage: p.PerPage, LastPage: last}, nil
}

