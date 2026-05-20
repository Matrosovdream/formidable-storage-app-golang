// cmd/seed populates the dev database with dummy reference data so the REST API
// can be exercised end-to-end:
//
//   - frm_entry_update_types: code/title pairs used by entry-history payloads
//   - sites:        one "Demo Site" (idempotent on URL)
//   - site_tokens:  one token attached to that site, printed at the end
//   - frm_fields:   a handful of placeholder fields
//   - frm_emails_log: a few rows of dummy email-log entries
//   - frm_entry_history: a few history rows referencing the seeded fields
//
// Re-running the command is safe: it upserts where possible and skips rows that
// already match. The bearer token is rotated only if the site was newly created.
package main

import (
	"context"
	"database/sql"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	"github.com/jmoiron/sqlx"
)

const (
	siteName = "Demo Site"
	siteURL  = "https://demo.local/"
)

var updateTypes = []struct{ code, title string }{
	{"created", "Created"},
	{"updated", "Updated"},
	{"deleted", "Deleted"},
	{"imported", "Imported"},
}

var fields = []struct {
	fieldID int64
	key     string
	typ     string
	label   string
}{
	{1001, "first_name", "text", "First name"},
	{1002, "last_name", "text", "Last name"},
	{1003, "email", "email", "Email"},
	{1004, "phone", "tel", "Phone"},
	{1005, "message", "textarea", "Message"},
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load()
	must(err)
	db, err := config.OpenDB(ctx, cfg.DB)
	must(err)
	defer db.Close()

	siteID, token, isNew := upsertSite(ctx, db)
	seedUpdateTypes(ctx, db)
	seedFields(ctx, db, siteID)
	seedEmails(ctx, db, siteID)
	seedHistory(ctx, db, siteID)

	fmt.Println("---")
	fmt.Printf("Demo site:      id=%d url=%s\n", siteID, siteURL)
	fmt.Printf("REST bearer:    %s\n", token)
	if !isNew {
		fmt.Println("(site already existed; token shown is the existing one)")
	}
	fmt.Println("Update types:  ", len(updateTypes))
	fmt.Println("Fields:        ", len(fields))
	fmt.Println("Emails:         3 rows")
	fmt.Println("Entry history:  3 rows")
}

// ---------- site + token (idempotent) ----------

func upsertSite(ctx context.Context, db *sqlx.DB) (siteID int64, token string, isNew bool) {
	var existing int64
	err := db.GetContext(ctx, &existing, "SELECT id FROM sites WHERE url = $1", siteURL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fail("look up site: %v", err)
	}
	if existing > 0 {
		var tok string
		if err := db.GetContext(ctx, &tok, "SELECT token FROM site_tokens WHERE site_id = $1 ORDER BY id DESC LIMIT 1", existing); err == nil {
			return existing, tok, false
		}
		// Site exists but has no token — issue one.
		tok = randomHex(32)
		_, err := db.ExecContext(ctx, `INSERT INTO site_tokens (site_id, token) VALUES ($1, $2)`, existing, tok)
		must(err)
		return existing, tok, false
	}

	tx, err := db.BeginTxx(ctx, nil)
	must(err)
	defer tx.Rollback()

	must(tx.QueryRowxContext(ctx, "INSERT INTO sites (name, url) VALUES ($1, $2) RETURNING id", siteName, siteURL).Scan(&siteID))
	token = randomHex(32)
	_, err = tx.ExecContext(ctx, "INSERT INTO site_tokens (site_id, token) VALUES ($1, $2)", siteID, token)
	must(err)
	must(tx.Commit())
	return siteID, token, true
}

// ---------- frm_entry_update_types (idempotent on code) ----------

func seedUpdateTypes(ctx context.Context, db *sqlx.DB) {
	for _, t := range updateTypes {
		var id int64
		err := db.GetContext(ctx, &id, "SELECT id FROM frm_entry_update_types WHERE code = $1", t.code)
		if err == nil {
			_, err := db.ExecContext(ctx, "UPDATE frm_entry_update_types SET title = $1, updated_at = NOW() WHERE id = $2", t.title, id)
			must(err)
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			fail("lookup update_type %s: %v", t.code, err)
		}
		_, err = db.ExecContext(ctx, "INSERT INTO frm_entry_update_types (code, title) VALUES ($1, $2)", t.code, t.title)
		must(err)
	}
}

// ---------- frm_fields (idempotent on (site_id, field_id)) ----------

func seedFields(ctx context.Context, db *sqlx.DB, siteID int64) {
	for _, f := range fields {
		var id int64
		err := db.GetContext(ctx, &id, "SELECT id FROM frm_fields WHERE site_id = $1 AND field_id = $2", siteID, f.fieldID)
		if err == nil {
			_, err := db.ExecContext(ctx,
				"UPDATE frm_fields SET key=$1, type=$2, label=$3, updated_at=NOW() WHERE id=$4",
				f.key, f.typ, f.label, id)
			must(err)
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			fail("lookup field %d: %v", f.fieldID, err)
		}
		_, err = db.ExecContext(ctx,
			"INSERT INTO frm_fields (field_id, site_id, key, type, label) VALUES ($1,$2,$3,$4,$5)",
			f.fieldID, siteID, f.key, f.typ, f.label)
		must(err)
	}
}

// ---------- frm_emails_log (upsert on (site_id, message_id)) ----------

func seedEmails(ctx context.Context, db *sqlx.DB, siteID int64) {
	type row struct {
		messageID, subject, from, to string
		status                       int16
	}
	rows := []row{
		{"demo-msg-1", "Thanks for your order", "noreply@demo.local", "alice@example.com", 1},
		{"demo-msg-2", "Password reset request", "noreply@demo.local", "bob@example.com", 1},
		{"demo-msg-3", "Weekly digest", "digest@demo.local", "carol@example.com", 0},
	}
	for _, r := range rows {
		_, err := db.ExecContext(ctx, `
			INSERT INTO frm_emails_log
				(entry_id, site_id, subject, message_id, email_from, email_to, content_plain, content_html, status, date_sent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (site_id, message_id) DO UPDATE SET
				subject=EXCLUDED.subject, email_from=EXCLUDED.email_from, email_to=EXCLUDED.email_to,
				status=EXCLUDED.status, updated_at=NOW()
		`, 9001, siteID, r.subject, r.messageID, r.from, r.to, "plain body", "<p>html body</p>", r.status, time.Now().UTC())
		must(err)
	}
}

// ---------- frm_entry_history (append; only insert if entry_id 9001 has no rows yet) ----------

func seedHistory(ctx context.Context, db *sqlx.DB, siteID int64) {
	var existing int64
	must(db.GetContext(ctx, &existing, "SELECT COUNT(*) FROM frm_entry_history WHERE site_id = $1 AND entry_id = 9001", siteID))
	if existing > 0 {
		return
	}

	var fieldRows []int64
	must(db.SelectContext(ctx, &fieldRows, "SELECT id FROM frm_fields WHERE site_id = $1 ORDER BY id ASC LIMIT 3", siteID))
	if len(fieldRows) < 3 {
		return
	}

	var createdType, updatedType int64
	must(db.GetContext(ctx, &createdType, "SELECT id FROM frm_entry_update_types WHERE code='created'"))
	must(db.GetContext(ctx, &updatedType, "SELECT id FROM frm_entry_update_types WHERE code='updated'"))

	now := time.Now().UTC()
	rows := []struct {
		fieldID  int64
		typeID   int64
		oldV     sql.NullString
		newV     sql.NullString
		changeAt time.Time
	}{
		{fieldRows[0], createdType, sql.NullString{}, sql.NullString{String: "Alice", Valid: true}, now.Add(-2 * time.Hour)},
		{fieldRows[1], updatedType, sql.NullString{String: "Smyth", Valid: true}, sql.NullString{String: "Smith", Valid: true}, now.Add(-1 * time.Hour)},
		{fieldRows[2], updatedType, sql.NullString{}, sql.NullString{String: "alice@example.com", Valid: true}, now},
	}
	for _, r := range rows {
		_, err := db.ExecContext(ctx, `
			INSERT INTO frm_entry_history (entry_id, site_id, field_id, update_type_id, old_value, new_value, change_date)
			VALUES (9001, $1, $2, $3, $4, $5, $6)
		`, siteID, r.fieldID, r.typeID, r.oldV, r.newV, r.changeAt)
		must(err)
	}
}

// ---------- helpers ----------

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		fail("rand: %v", err)
	}
	return hex.EncodeToString(buf)
}

func must(err error) {
	if err != nil {
		fail("%v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "seed: "+format+"\n", args...)
	os.Exit(1)
}
