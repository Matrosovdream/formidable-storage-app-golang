package entity

import (
	"database/sql"
	"time"
)

type FrmField struct {
	ID        int64          `db:"id"`
	FieldID   sql.NullInt64  `db:"field_id"`
	SiteID    int64          `db:"site_id"`
	Key       sql.NullString `db:"key"`
	Type      sql.NullString `db:"type"`
	Label     sql.NullString `db:"label"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type FrmEntryUpdateType struct {
	ID        int64          `db:"id"`
	Code      sql.NullString `db:"code"`
	Title     sql.NullString `db:"title"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type FrmEntryHistory struct {
	ID           int64          `db:"id"`
	EntryID      sql.NullInt64  `db:"entry_id"`
	SiteID       int64          `db:"site_id"`
	FieldID      int64          `db:"field_id"`
	UserID       sql.NullInt64  `db:"user_id"`
	UpdateTypeID sql.NullInt64  `db:"update_type_id"`
	OldValue     sql.NullString `db:"old_value"`
	NewValue     sql.NullString `db:"new_value"`
	ChangeDate   sql.NullTime   `db:"change_date"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}
