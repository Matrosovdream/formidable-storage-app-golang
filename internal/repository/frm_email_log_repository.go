package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/jmoiron/sqlx"
)

type FrmEmailLogRepository struct {
	db *sqlx.DB
}

func NewFrmEmailLogRepository(db *sqlx.DB) *FrmEmailLogRepository {
	return &FrmEmailLogRepository{db: db}
}

func (r *FrmEmailLogRepository) q(qu Querier) Querier {
	if qu == nil {
		return r.db
	}
	return qu
}

const frmEmailLogColumns = `id, entry_id, site_id, form_id, subject, message_id, email_from, email_to,
	people, headers, error_text, content_plain, content_html, status, date_sent, mailer,
	attachments, initiator_name, initiator_file, original_log_id, created_at, updated_at`

type FrmEmailLogInput struct {
	EntryID       sql.NullInt64
	FormID        sql.NullInt64
	Subject       sql.NullString
	MessageID     sql.NullString
	EmailFrom     sql.NullString
	EmailTo       sql.NullString
	People        sql.NullString
	Headers       sql.NullString
	ErrorText     sql.NullString
	ContentPlain  sql.NullString
	ContentHTML   sql.NullString
	Status        int16
	DateSent      sql.NullTime
	Mailer        sql.NullString
	Attachments   int16
	InitiatorName sql.NullString
	InitiatorFile sql.NullString
	OriginalLogID sql.NullInt64
}

// UpsertMany inserts (or updates on conflict (site_id, message_id)) chunks of 200 rows.
func (r *FrmEmailLogRepository) UpsertMany(ctx context.Context, qu Querier, siteID int64, rows []FrmEmailLogInput) error {
	if len(rows) == 0 {
		return nil
	}
	q := r.q(qu)
	const chunk = 200
	const cols = "entry_id, site_id, form_id, subject, message_id, email_from, email_to, people, headers, error_text, content_plain, content_html, status, date_sent, mailer, attachments, initiator_name, initiator_file, original_log_id, created_at, updated_at"
	const ncols = 21

	for _, batch := range ChunkRows(rows, chunk) {
		args := make([]any, 0, len(batch)*ncols)
		placeholders := make([]string, 0, len(batch))
		pos := 1
		now := time.Now().UTC()
		for _, in := range batch {
			marks := make([]string, ncols)
			for i := 0; i < ncols; i++ {
				marks[i] = "$" + strconv.Itoa(pos)
				pos++
			}
			placeholders = append(placeholders, "("+strings.Join(marks, ", ")+")")
			args = append(args,
				in.EntryID, siteID, in.FormID, in.Subject, in.MessageID, in.EmailFrom, in.EmailTo,
				in.People, in.Headers, in.ErrorText, in.ContentPlain, in.ContentHTML, in.Status,
				in.DateSent, in.Mailer, in.Attachments, in.InitiatorName, in.InitiatorFile, in.OriginalLogID,
				now, now,
			)
		}
		stmt := "INSERT INTO frm_emails_log (" + cols + ") VALUES " + strings.Join(placeholders, ", ") +
			` ON CONFLICT (site_id, message_id) DO UPDATE SET
				entry_id        = EXCLUDED.entry_id,
				form_id         = EXCLUDED.form_id,
				subject         = EXCLUDED.subject,
				email_from      = EXCLUDED.email_from,
				email_to        = EXCLUDED.email_to,
				people          = EXCLUDED.people,
				headers         = EXCLUDED.headers,
				error_text      = EXCLUDED.error_text,
				content_plain   = EXCLUDED.content_plain,
				content_html    = EXCLUDED.content_html,
				status          = EXCLUDED.status,
				date_sent       = EXCLUDED.date_sent,
				mailer          = EXCLUDED.mailer,
				attachments     = EXCLUDED.attachments,
				initiator_name  = EXCLUDED.initiator_name,
				initiator_file  = EXCLUDED.initiator_file,
				original_log_id = EXCLUDED.original_log_id,
				updated_at      = NOW()`
		if _, err := q.ExecContext(ctx, stmt, args...); err != nil {
			return err
		}
	}
	return nil
}

func (r *FrmEmailLogRepository) FindByEntry(ctx context.Context, qu Querier, siteID, entryID int64) ([]entity.FrmEmailLog, error) {
	var out []entity.FrmEmailLog
	err := r.q(qu).SelectContext(ctx, &out,
		"SELECT "+frmEmailLogColumns+" FROM frm_emails_log WHERE site_id = $1 AND entry_id = $2 ORDER BY id DESC",
		siteID, entryID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// List performs filter/sort/pagination over frm_emails_log scoped to siteID.
// Filter keys (whitelisted): entry_id, form_id, status, mailer, message_id,
// subject (ILIKE), email_from (ILIKE), email_to (ILIKE), date_sent_from, date_sent_to.
// Sort whitelist: id, date_sent, status, subject, email_from, email_to, created_at.
func (r *FrmEmailLogRepository) List(ctx context.Context, qu Querier, siteID int64, p ListParams) (ListResult[entity.FrmEmailLog], error) {
	p.Normalize(25, 200)

	where := []string{"site_id = $1"}
	args := []any{siteID}
	pos := 2

	likeKeys := map[string]bool{"subject": true, "email_from": true, "email_to": true}
	eqKeys := map[string]bool{"entry_id": true, "form_id": true, "status": true, "mailer": true, "message_id": true}

	for k, v := range p.Filters {
		switch {
		case eqKeys[k]:
			where = append(where, fmt.Sprintf("%s = $%d", k, pos))
			args = append(args, v)
			pos++
		case likeKeys[k]:
			where = append(where, fmt.Sprintf("%s ILIKE $%d", k, pos))
			args = append(args, "%"+toString(v)+"%")
			pos++
		case k == "date_sent_from":
			where = append(where, fmt.Sprintf("date_sent >= $%d", pos))
			args = append(args, v)
			pos++
		case k == "date_sent_to":
			where = append(where, fmt.Sprintf("date_sent <= $%d", pos))
			args = append(args, v)
			pos++
		}
	}

	sortCol := "id"
	switch p.SortBy {
	case "id", "date_sent", "status", "subject", "email_from", "email_to", "created_at":
		sortCol = p.SortBy
	}

	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.q(qu).GetContext(ctx, &total, "SELECT COUNT(*) FROM frm_emails_log WHERE "+whereSQL, args...); err != nil {
		return ListResult[entity.FrmEmailLog]{}, err
	}

	listArgs := append(args, p.PerPage, p.Offset())
	stmt := fmt.Sprintf(
		"SELECT %s FROM frm_emails_log WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		frmEmailLogColumns, whereSQL, sortCol, p.SortDir, pos, pos+1,
	)

	var data []entity.FrmEmailLog
	if err := r.q(qu).SelectContext(ctx, &data, stmt, listArgs...); err != nil {
		return ListResult[entity.FrmEmailLog]{}, err
	}

	last := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if last < 1 {
		last = 1
	}
	return ListResult[entity.FrmEmailLog]{Data: data, Total: total, Page: p.Page, PerPage: p.PerPage, LastPage: last}, nil
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(v)
	}
}
