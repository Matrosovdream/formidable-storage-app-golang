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

type FrmEpShipmentHistoryRepository struct {
	db *sqlx.DB
}

func NewFrmEpShipmentHistoryRepository(db *sqlx.DB) *FrmEpShipmentHistoryRepository {
	return &FrmEpShipmentHistoryRepository{db: db}
}

func (r *FrmEpShipmentHistoryRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const frmEpHistoryColumns = "id, easypost_shipment_id, user_id, site_id, change_type, description, created_at, updated_at"

type FrmEpShipmentHistoryInput struct {
	EasypostShipmentID sql.NullString
	UserID             sql.NullInt64
	ChangeType         sql.NullString
	Description        sql.NullString
}

func (r *FrmEpShipmentHistoryRepository) InsertMany(ctx context.Context, qu Querier, siteID int64, rows []FrmEpShipmentHistoryInput) error {
	if len(rows) == 0 {
		return nil
	}
	q := r.q(qu)
	const chunk = 200
	for _, batch := range ChunkRows(rows, chunk) {
		args := make([]any, 0, len(batch)*6)
		placeholders := make([]string, 0, len(batch))
		pos := 1
		now := time.Now().UTC()
		for _, in := range batch {
			marks := make([]string, 6)
			for i := 0; i < 6; i++ {
				marks[i] = "$" + strconv.Itoa(pos)
				pos++
			}
			placeholders = append(placeholders, "("+strings.Join(marks, ", ")+")")
			args = append(args, in.EasypostShipmentID, in.UserID, siteID, in.ChangeType, in.Description, now)
		}
		stmt := "INSERT INTO frm_easypost_shipment_history (easypost_shipment_id, user_id, site_id, change_type, description, created_at) VALUES " + strings.Join(placeholders, ", ")
		if _, err := q.ExecContext(ctx, stmt, args...); err != nil {
			return err
		}
	}
	return nil
}

func (r *FrmEpShipmentHistoryRepository) List(ctx context.Context, qu Querier, siteID int64, p ListParams) (ListResult[entity.FrmEpShipmentHistory], error) {
	p.Normalize(25, 200)

	where := []string{"site_id = $1"}
	args := []any{siteID}
	pos := 2

	for k, v := range p.Filters {
		switch k {
		case "easypost_shipment_id", "user_id", "change_type":
			where = append(where, fmt.Sprintf("%s = $%d", k, pos))
			args = append(args, v)
			pos++
		case "created_from":
			where = append(where, fmt.Sprintf("created_at >= $%d", pos))
			args = append(args, v)
			pos++
		case "created_to":
			where = append(where, fmt.Sprintf("created_at <= $%d", pos))
			args = append(args, v)
			pos++
		}
	}

	sortCol := "id"
	switch p.SortBy {
	case "id", "created_at", "change_type":
		sortCol = p.SortBy
	}

	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.q(qu).GetContext(ctx, &total, "SELECT COUNT(*) FROM frm_easypost_shipment_history WHERE "+whereSQL, args...); err != nil {
		return ListResult[entity.FrmEpShipmentHistory]{}, err
	}

	listArgs := append(args, p.PerPage, p.Offset())
	stmt := fmt.Sprintf(
		"SELECT %s FROM frm_easypost_shipment_history WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		frmEpHistoryColumns, whereSQL, sortCol, p.SortDir, pos, pos+1,
	)

	var data []entity.FrmEpShipmentHistory
	if err := r.q(qu).SelectContext(ctx, &data, stmt, listArgs...); err != nil {
		return ListResult[entity.FrmEpShipmentHistory]{}, err
	}

	last := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if last < 1 {
		last = 1
	}
	return ListResult[entity.FrmEpShipmentHistory]{Data: data, Total: total, Page: p.Page, PerPage: p.PerPage, LastPage: last}, nil
}
