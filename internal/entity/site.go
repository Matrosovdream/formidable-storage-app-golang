package entity

import (
	"database/sql"
	"time"
)

type Site struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type SiteToken struct {
	ID         int64        `db:"id"`
	SiteID     int64        `db:"site_id"`
	Token      string       `db:"token"`
	ValidUntil sql.NullTime `db:"valid_until"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
}
