package converter

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

func NullString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	v := s.String
	return &v
}

func NullInt64(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func NullBool(b sql.NullBool) *bool {
	if !b.Valid {
		return nil
	}
	v := b.Bool
	return &v
}

func NullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func NullDecimal(d decimal.NullDecimal) *string {
	if !d.Valid {
		return nil
	}
	v := d.Decimal.String()
	return &v
}
