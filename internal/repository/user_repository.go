package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const userColumns = "id, name, email, email_verified_at, password, remember_token, created_at, updated_at"

func (r *UserRepository) FindByID(ctx context.Context, qu Querier, id int64) (*entity.User, error) {
	var u entity.User
	err := r.q(qu).GetContext(ctx, &u, "SELECT "+userColumns+" FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, qu Querier, email string) (*entity.User, error) {
	var u entity.User
	err := r.q(qu).GetContext(ctx, &u, "SELECT "+userColumns+" FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, qu Querier, u *entity.User) error {
	const stmt = `INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.q(qu).QueryRowxContext(ctx, stmt, u.Name, u.Email, u.Password).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}
