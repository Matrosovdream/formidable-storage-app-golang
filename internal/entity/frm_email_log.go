package entity

import (
	"database/sql"
	"time"
)

type FrmEmailLog struct {
	ID             int64          `db:"id"`
	EntryID        sql.NullInt64  `db:"entry_id"`
	SiteID         int64          `db:"site_id"`
	FormID         sql.NullInt64  `db:"form_id"`
	Subject        sql.NullString `db:"subject"`
	MessageID      sql.NullString `db:"message_id"`
	EmailFrom      sql.NullString `db:"email_from"`
	EmailTo        sql.NullString `db:"email_to"`
	People         sql.NullString `db:"people"`
	Headers        sql.NullString `db:"headers"`
	ErrorText      sql.NullString `db:"error_text"`
	ContentPlain   sql.NullString `db:"content_plain"`
	ContentHTML    sql.NullString `db:"content_html"`
	Status         int16          `db:"status"`
	DateSent       sql.NullTime   `db:"date_sent"`
	Mailer         sql.NullString `db:"mailer"`
	Attachments    int16          `db:"attachments"`
	InitiatorName  sql.NullString `db:"initiator_name"`
	InitiatorFile  sql.NullString `db:"initiator_file"`
	OriginalLogID  sql.NullInt64  `db:"original_log_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
