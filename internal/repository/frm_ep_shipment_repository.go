package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type FrmEpShipmentRepository struct {
	db *sqlx.DB
}

func NewFrmEpShipmentRepository(db *sqlx.DB) *FrmEpShipmentRepository {
	return &FrmEpShipmentRepository{db: db}
}

func (r *FrmEpShipmentRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const frmEpShipmentColumns = "id, easypost_shipment_id, entry_id, site_id, is_return, status, tracking_code, tracking_url, refund_status, mode, created_at, updated_at"

type FrmEpShipmentInput struct {
	EasypostShipmentID sql.NullString
	EntryID            sql.NullInt64
	IsReturn           bool
	Status             sql.NullString
	TrackingCode       sql.NullString
	TrackingURL        sql.NullString
	RefundStatus       sql.NullString
	Mode               sql.NullString
}

// UpsertMany on (site_id, easypost_shipment_id). Updates status/tracking/refund/mode/is_return on conflict.
func (r *FrmEpShipmentRepository) UpsertMany(ctx context.Context, qu Querier, siteID int64, rows []FrmEpShipmentInput) error {
	if len(rows) == 0 {
		return nil
	}
	q := r.q(qu)
	const chunk = 200
	for _, batch := range ChunkRows(rows, chunk) {
		for _, in := range batch {
			_, err := q.ExecContext(ctx,
				`INSERT INTO frm_easypost_shipments (
					easypost_shipment_id, entry_id, site_id, is_return, status,
					tracking_code, tracking_url, refund_status, mode
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				ON CONFLICT (site_id, easypost_shipment_id) DO UPDATE SET
					entry_id      = EXCLUDED.entry_id,
					is_return     = EXCLUDED.is_return,
					status        = EXCLUDED.status,
					tracking_code = EXCLUDED.tracking_code,
					tracking_url  = EXCLUDED.tracking_url,
					refund_status = EXCLUDED.refund_status,
					mode          = EXCLUDED.mode,
					updated_at    = NOW()`,
				in.EasypostShipmentID, in.EntryID, siteID, in.IsReturn, in.Status,
				in.TrackingCode, in.TrackingURL, in.RefundStatus, in.Mode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *FrmEpShipmentRepository) FindByEasypostID(ctx context.Context, qu Querier, siteID int64, easypostShipmentID string) (*entity.FrmEpShipment, error) {
	var s entity.FrmEpShipment
	err := r.q(qu).GetContext(ctx, &s,
		"SELECT "+frmEpShipmentColumns+" FROM frm_easypost_shipments WHERE site_id = $1 AND easypost_shipment_id = $2",
		siteID, easypostShipmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}
