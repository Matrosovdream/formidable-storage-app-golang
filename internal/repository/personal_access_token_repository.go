package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type PersonalAccessTokenRepository struct {
	db *sqlx.DB
}

func NewPersonalAccessTokenRepository(db *sqlx.DB) *PersonalAccessTokenRepository {
	return &PersonalAccessTokenRepository{db: db}
}

func (r *PersonalAccessTokenRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const patColumns = "id, tokenable_type, tokenable_id, name, token, abilities, last_used_at, expires_at, created_at, updated_at"

func (r *PersonalAccessTokenRepository) FindByID(ctx context.Context, qu Querier, id int64) (*entity.PersonalAccessToken, error) {
	var t entity.PersonalAccessToken
	err := r.q(qu).GetContext(ctx, &t, "SELECT "+patColumns+" FROM personal_access_tokens WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *PersonalAccessTokenRepository) FindByHash(ctx context.Context, qu Querier, hash string) (*entity.PersonalAccessToken, error) {
	var t entity.PersonalAccessToken
	err := r.q(qu).GetContext(ctx, &t, "SELECT "+patColumns+" FROM personal_access_tokens WHERE token = $1", hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *PersonalAccessTokenRepository) Create(ctx context.Context, qu Querier, t *entity.PersonalAccessToken) error {
	const stmt = `INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token, abilities, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	return r.q(qu).QueryRowxContext(ctx, stmt,
		t.TokenableType, t.TokenableID, t.Name, t.Token, t.Abilities, t.ExpiresAt,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *PersonalAccessTokenRepository) Touch(ctx context.Context, qu Querier, id int64) error {
	_, err := r.q(qu).ExecContext(ctx, `UPDATE personal_access_tokens SET last_used_at = $1, updated_at = NOW() WHERE id = $2`, time.Now().UTC(), id)
	return err
}

func (r *PersonalAccessTokenRepository) Delete(ctx context.Context, qu Querier, id int64) error {
	_, err := r.q(qu).ExecContext(ctx, "DELETE FROM personal_access_tokens WHERE id = $1", id)
	return err
}
