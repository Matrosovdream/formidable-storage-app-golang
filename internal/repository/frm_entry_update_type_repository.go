package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type FrmEntryUpdateTypeRepository struct {
	db *sqlx.DB
}

func NewFrmEntryUpdateTypeRepository(db *sqlx.DB) *FrmEntryUpdateTypeRepository {
	return &FrmEntryUpdateTypeRepository{db: db}
}

func (r *FrmEntryUpdateTypeRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const frmEntryUpdateTypeColumns = "id, code, title, created_at, updated_at"

func (r *FrmEntryUpdateTypeRepository) FindAll(ctx context.Context, qu Querier) ([]entity.FrmEntryUpdateType, error) {
	var out []entity.FrmEntryUpdateType
	err := r.q(qu).SelectContext(ctx, &out, "SELECT "+frmEntryUpdateTypeColumns+" FROM frm_entry_update_types ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *FrmEntryUpdateTypeRepository) FindByCode(ctx context.Context, qu Querier, code string) (*entity.FrmEntryUpdateType, error) {
	var t entity.FrmEntryUpdateType
	err := r.q(qu).GetContext(ctx, &t, "SELECT "+frmEntryUpdateTypeColumns+" FROM frm_entry_update_types WHERE code = $1", code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}
