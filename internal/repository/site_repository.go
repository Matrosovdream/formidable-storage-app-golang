package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type SiteRepository struct {
	db *sqlx.DB
}

func NewSiteRepository(db *sqlx.DB) *SiteRepository {
	return &SiteRepository{db: db}
}

func (r *SiteRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const siteColumns = "id, name, url, created_at, updated_at"

func (r *SiteRepository) FindByID(ctx context.Context, qu Querier, id int64) (*entity.Site, error) {
	var s entity.Site
	err := r.q(qu).GetContext(ctx, &s, "SELECT "+siteColumns+" FROM sites WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SiteRepository) FindAll(ctx context.Context, qu Querier) ([]entity.Site, error) {
	var out []entity.Site
	if err := r.q(qu).SelectContext(ctx, &out, "SELECT "+siteColumns+" FROM sites ORDER BY id DESC"); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SiteRepository) Create(ctx context.Context, qu Querier, s *entity.Site) error {
	const stmt = `INSERT INTO sites (name, url) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.q(qu).QueryRowxContext(ctx, stmt, s.Name, s.URL).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SiteRepository) Update(ctx context.Context, qu Querier, s *entity.Site) error {
	const stmt = `UPDATE sites SET name = $1, url = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.q(qu).ExecContext(ctx, stmt, s.Name, s.URL, s.ID)
	return err
}

func (r *SiteRepository) Delete(ctx context.Context, qu Querier, id int64) error {
	_, err := r.q(qu).ExecContext(ctx, "DELETE FROM sites WHERE id = $1", id)
	return err
}

type SiteCounts struct {
	Fields       int64
	Emails       int64
	EntryHistory int64
}

func (r *SiteRepository) Counts(ctx context.Context, qu Querier, siteID int64) (SiteCounts, error) {
	var c SiteCounts
	q := r.q(qu)
	if err := q.GetContext(ctx, &c.Fields, "SELECT COUNT(*) FROM frm_fields WHERE site_id = $1", siteID); err != nil {
		return c, err
	}
	if err := q.GetContext(ctx, &c.Emails, "SELECT COUNT(*) FROM frm_emails_log WHERE site_id = $1", siteID); err != nil {
		return c, err
	}
	if err := q.GetContext(ctx, &c.EntryHistory, "SELECT COUNT(*) FROM frm_entry_history WHERE site_id = $1", siteID); err != nil {
		return c, err
	}
	return c, nil
}
