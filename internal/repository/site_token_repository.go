package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type SiteTokenRepository struct {
	db *sqlx.DB
}

func NewSiteTokenRepository(db *sqlx.DB) *SiteTokenRepository {
	return &SiteTokenRepository{db: db}
}

func (r *SiteTokenRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const siteTokenColumns = "id, site_id, token, valid_until, created_at, updated_at"

func (r *SiteTokenRepository) FindByToken(ctx context.Context, qu Querier, token string) (*entity.SiteToken, error) {
	var t entity.SiteToken
	err := r.q(qu).GetContext(ctx, &t, "SELECT "+siteTokenColumns+" FROM site_tokens WHERE token = $1", token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *SiteTokenRepository) FindBySiteID(ctx context.Context, qu Querier, siteID int64) (*entity.SiteToken, error) {
	var t entity.SiteToken
	err := r.q(qu).GetContext(ctx, &t, "SELECT "+siteTokenColumns+" FROM site_tokens WHERE site_id = $1 ORDER BY id DESC LIMIT 1", siteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *SiteTokenRepository) Create(ctx context.Context, qu Querier, t *entity.SiteToken) error {
	const stmt = `INSERT INTO site_tokens (site_id, token, valid_until) VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.q(qu).QueryRowxContext(ctx, stmt, t.SiteID, t.Token, t.ValidUntil).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *SiteTokenRepository) DeleteBySiteID(ctx context.Context, qu Querier, siteID int64) error {
	_, err := r.q(qu).ExecContext(ctx, "DELETE FROM site_tokens WHERE site_id = $1", siteID)
	return err
}
