package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type FrmFieldRepository struct {
	db *sqlx.DB
}

func NewFrmFieldRepository(db *sqlx.DB) *FrmFieldRepository {
	return &FrmFieldRepository{db: db}
}

func (r *FrmFieldRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const frmFieldColumns = "id, field_id, site_id, key, type, label, created_at, updated_at"

func (r *FrmFieldRepository) FindBySite(ctx context.Context, qu Querier, siteID int64) ([]entity.FrmField, error) {
	var out []entity.FrmField
	err := r.q(qu).SelectContext(ctx, &out, "SELECT "+frmFieldColumns+" FROM frm_fields WHERE site_id = $1 ORDER BY id ASC", siteID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FrmFieldInput is the payload row used by UpsertMany.
type FrmFieldInput struct {
	FieldID int64
	Key     string
	Type    string
	Label   string
}

// UpsertMany bulk-inserts fields scoped to siteID. Uses (site_id, field_id) as the natural key.
// If (site_id, field_id) already exists, key/type/label are updated.
//
// A partial unique index on (site_id, field_id) is required for this to use ON CONFLICT.
// Since the migration does not include such an index, we run upsert by emulating it via a small
// SELECT + UPDATE/INSERT inside the same transaction, in chunks of 200.
func (r *FrmFieldRepository) UpsertMany(ctx context.Context, qu Querier, siteID int64, inputs []FrmFieldInput) error {
	if len(inputs) == 0 {
		return nil
	}
	q := r.q(qu)
	const chunk = 200
	for _, batch := range ChunkRows(inputs, chunk) {
		for _, in := range batch {
			res, err := q.ExecContext(ctx,
				`UPDATE frm_fields SET key=$1, type=$2, label=$3, updated_at=NOW()
				 WHERE site_id=$4 AND field_id=$5`,
				in.Key, in.Type, in.Label, siteID, in.FieldID)
			if err != nil {
				return err
			}
			affected, _ := res.RowsAffected()
			if affected > 0 {
				continue
			}
			if _, err := q.ExecContext(ctx,
				`INSERT INTO frm_fields (field_id, site_id, key, type, label) VALUES ($1, $2, $3, $4, $5)`,
				in.FieldID, siteID, in.Key, in.Type, in.Label); err != nil {
				return err
			}
		}
	}
	return nil
}

// List performs filter/sort/pagination over frm_fields scoped to siteID.
// Filter keys (whitelisted): field_id, type (=); key, label (ILIKE).
// Sort whitelist: id, field_id, key, type, label, created_at, updated_at.
func (r *FrmFieldRepository) List(ctx context.Context, qu Querier, siteID int64, p ListParams) (ListResult[entity.FrmField], error) {
	p.Normalize(25, 200)

	where := []string{"site_id = $1"}
	args := []any{siteID}
	pos := 2

	likeKeys := map[string]bool{"key": true, "label": true}
	eqKeys := map[string]bool{"field_id": true, "type": true}

	for k, v := range p.Filters {
		switch {
		case eqKeys[k]:
			where = append(where, fmt.Sprintf("%s = $%d", k, pos))
			args = append(args, v)
			pos++
		case likeKeys[k]:
			where = append(where, fmt.Sprintf("%s ILIKE $%d", k, pos))
			args = append(args, "%"+toString(v)+"%")
			pos++
		}
	}

	sortCol := "id"
	switch p.SortBy {
	case "id", "field_id", "key", "type", "label", "created_at", "updated_at":
		sortCol = p.SortBy
	}

	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.q(qu).GetContext(ctx, &total, "SELECT COUNT(*) FROM frm_fields WHERE "+whereSQL, args...); err != nil {
		return ListResult[entity.FrmField]{}, err
	}

	listArgs := append(args, p.PerPage, p.Offset())
	stmt := fmt.Sprintf(
		"SELECT %s FROM frm_fields WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		frmFieldColumns, whereSQL, sortCol, p.SortDir, pos, pos+1,
	)

	var data []entity.FrmField
	if err := r.q(qu).SelectContext(ctx, &data, stmt, listArgs...); err != nil {
		return ListResult[entity.FrmField]{}, err
	}

	last := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if last < 1 {
		last = 1
	}
	return ListResult[entity.FrmField]{Data: data, Total: total, Page: p.Page, PerPage: p.PerPage, LastPage: last}, nil
}
