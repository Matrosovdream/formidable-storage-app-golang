package entity

import (
	"database/sql"
	"time"
)

type User struct {
	ID              int64          `db:"id"`
	Name            string         `db:"name"`
	Email           string         `db:"email"`
	EmailVerifiedAt sql.NullTime   `db:"email_verified_at"`
	Password        string         `db:"password"`
	RememberToken   sql.NullString `db:"remember_token"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

type PasswordResetToken struct {
	Email     string         `db:"email"`
	Token     string         `db:"token"`
	CreatedAt sql.NullTime   `db:"created_at"`
}

type Session struct {
	ID           string         `db:"id"`
	UserID       sql.NullInt64  `db:"user_id"`
	IPAddress    sql.NullString `db:"ip_address"`
	UserAgent    sql.NullString `db:"user_agent"`
	Payload      string         `db:"payload"`
	LastActivity int64          `db:"last_activity"`
}

type PersonalAccessToken struct {
	ID            int64          `db:"id"`
	TokenableType string         `db:"tokenable_type"`
	TokenableID   int64          `db:"tokenable_id"`
	Name          string         `db:"name"`
	Token         string         `db:"token"`
	Abilities     sql.NullString `db:"abilities"`
	LastUsedAt    sql.NullTime   `db:"last_used_at"`
	ExpiresAt     sql.NullTime   `db:"expires_at"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}
