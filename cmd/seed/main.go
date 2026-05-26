// cmd/seed populates the dev database with dummy reference data so the REST API
// can be exercised end-to-end:
//
//   - users:        one superadmin (idempotent on email) with a fresh sanctum token
//   - frm_entry_update_types: code/title pairs used by entry-history payloads
//   - sites:        one "Demo Site" (idempotent on URL)
//   - site_tokens:  one token attached to that site, printed at the end
//   - frm_fields:   a handful of placeholder fields
//   - frm_emails_log: a few rows of dummy email-log entries
//   - frm_entry_history: a few history rows referencing the seeded fields
//
// All seed data lives under cmd/seed/seeds/*.json and is embedded into the
// binary, so the command is self-contained and the data can be edited without
// touching Go code.
//
// Re-running the command is safe: it upserts where possible and skips rows that
// already match. The bearer token is rotated only if the site was newly created.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// entryID is the synthetic entry the email and history fixtures reference.
const entryID int64 = 9001

//go:embed seeds/*.json
var seedFS embed.FS

type adminSeed struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type siteSeed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type updateTypeSeed struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

type fieldSeed struct {
	FieldID int64  `json:"field_id"`
	Key     string `json:"key"`
	Type    string `json:"type"`
	Label   string `json:"label"`
}

type emailSeed struct {
	MessageID    string `json:"message_id"`
	Subject      string `json:"subject"`
	From         string `json:"from"`
	To           string `json:"to"`
	Status       int16  `json:"status"`
	ContentPlain string `json:"content_plain"`
	ContentHTML  string `json:"content_html"`
}

type historySeed struct {
	FieldIndex int     `json:"field_index"`
	UpdateType string  `json:"update_type"`
	OldValue   *string `json:"old_value"`
	NewValue   *string `json:"new_value"`
	HoursAgo   int     `json:"hours_ago"`
}

type seedData struct {
	Admin       adminSeed
	Site        siteSeed
	UpdateTypes []updateTypeSeed
	Fields      []fieldSeed
	Emails      []emailSeed
	History     []historySeed
}

func loadSeeds() seedData {
	var s seedData
	mustDecode("seeds/admin.json", &s.Admin)
	mustDecode("seeds/site.json", &s.Site)
	mustDecode("seeds/update_types.json", &s.UpdateTypes)
	mustDecode("seeds/fields.json", &s.Fields)
	mustDecode("seeds/emails.json", &s.Emails)
	mustDecode("seeds/history.json", &s.History)
	return s
}

func mustDecode(path string, out any) {
	data, err := seedFS.ReadFile(path)
	if err != nil {
		fail("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		fail("decode %s: %v", path, err)
	}
}

func main() {
	seeds := loadSeeds()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load()
	must(err)
	db, err := config.OpenDB(ctx, cfg.DB)
	must(err)
	defer db.Close()

	bcryptCost := cfg.Auth.BcryptCost
	if bcryptCost <= 0 {
		bcryptCost = bcrypt.DefaultCost
	}

	adminID, adminCreated := upsertAdminUser(ctx, db, bcryptCost, seeds.Admin)
	adminToken := issueSanctumToken(ctx, db, adminID)
	siteID, token, isNew := upsertSite(ctx, db, seeds.Site)
	seedUpdateTypes(ctx, db, seeds.UpdateTypes)
	seedFields(ctx, db, siteID, seeds.Fields)
	seedEmails(ctx, db, siteID, seeds.Emails)
	seedHistory(ctx, db, siteID, seeds.History)

	fmt.Println("---")
	fmt.Printf("Admin user:     id=%d email=%s\n", adminID, seeds.Admin.Email)
	fmt.Printf("Admin password: %s\n", seeds.Admin.Password)
	fmt.Printf("Admin sanctum:  %s\n", adminToken)
	if !adminCreated {
		fmt.Println("(admin already existed; password reset to the value above; new sanctum token issued)")
	}
	fmt.Println("---")
	fmt.Printf("Demo site:      id=%d url=%s\n", siteID, seeds.Site.URL)
	fmt.Printf("REST bearer:    %s\n", token)
	if !isNew {
		fmt.Println("(site already existed; token shown is the existing one)")
	}
	fmt.Println("Update types:  ", len(seeds.UpdateTypes))
	fmt.Println("Fields:        ", len(seeds.Fields))
	fmt.Println("Emails:        ", len(seeds.Emails), "rows")
	fmt.Println("Entry history: ", len(seeds.History), "rows")
}

// ---------- admin user (idempotent on email) ----------

func upsertAdminUser(ctx context.Context, db *sqlx.DB, bcryptCost int, admin adminSeed) (userID int64, created bool) {
	hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcryptCost)
	must(err)

	err = db.GetContext(ctx, &userID, "SELECT id FROM users WHERE email = $1", admin.Email)
	if err == nil {
		_, err := db.ExecContext(ctx,
			"UPDATE users SET name=$1, password=$2, updated_at=NOW() WHERE id=$3",
			admin.Name, string(hash), userID)
		must(err)
		return userID, false
	}
	if !errors.Is(err, sql.ErrNoRows) {
		fail("lookup admin: %v", err)
	}
	must(db.QueryRowxContext(ctx,
		"INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id",
		admin.Name, admin.Email, string(hash),
	).Scan(&userID))
	return userID, true
}

// issueSanctumToken creates a personal_access_tokens row for userID and returns the wire-format "<id>|<plain>".
func issueSanctumToken(ctx context.Context, db *sqlx.DB, userID int64) string {
	plain := randomHex(40)
	hash := sha256.Sum256([]byte(plain))
	var tokenID int64
	must(db.QueryRowxContext(ctx,
		`INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token, abilities)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		"App\\Models\\User", userID, "seed", hex.EncodeToString(hash[:]), `["*"]`,
	).Scan(&tokenID))
	return fmt.Sprintf("%d|%s", tokenID, plain)
}

// ---------- site + token (idempotent) ----------

func upsertSite(ctx context.Context, db *sqlx.DB, site siteSeed) (siteID int64, token string, isNew bool) {
	var existing int64
	err := db.GetContext(ctx, &existing, "SELECT id FROM sites WHERE url = $1", site.URL)
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

	must(tx.QueryRowxContext(ctx, "INSERT INTO sites (name, url) VALUES ($1, $2) RETURNING id", site.Name, site.URL).Scan(&siteID))
	token = randomHex(32)
	_, err = tx.ExecContext(ctx, "INSERT INTO site_tokens (site_id, token) VALUES ($1, $2)", siteID, token)
	must(err)
	must(tx.Commit())
	return siteID, token, true
}

// ---------- frm_entry_update_types (idempotent on code) ----------

func seedUpdateTypes(ctx context.Context, db *sqlx.DB, types []updateTypeSeed) {
	for _, t := range types {
		var id int64
		err := db.GetContext(ctx, &id, "SELECT id FROM frm_entry_update_types WHERE code = $1", t.Code)
		if err == nil {
			_, err := db.ExecContext(ctx, "UPDATE frm_entry_update_types SET title = $1, updated_at = NOW() WHERE id = $2", t.Title, id)
			must(err)
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			fail("lookup update_type %s: %v", t.Code, err)
		}
		_, err = db.ExecContext(ctx, "INSERT INTO frm_entry_update_types (code, title) VALUES ($1, $2)", t.Code, t.Title)
		must(err)
	}
}

// ---------- frm_fields (idempotent on (site_id, field_id)) ----------

func seedFields(ctx context.Context, db *sqlx.DB, siteID int64, fields []fieldSeed) {
	for _, f := range fields {
		var id int64
		err := db.GetContext(ctx, &id, "SELECT id FROM frm_fields WHERE site_id = $1 AND field_id = $2", siteID, f.FieldID)
		if err == nil {
			_, err := db.ExecContext(ctx,
				"UPDATE frm_fields SET key=$1, type=$2, label=$3, updated_at=NOW() WHERE id=$4",
				f.Key, f.Type, f.Label, id)
			must(err)
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			fail("lookup field %d: %v", f.FieldID, err)
		}
		_, err = db.ExecContext(ctx,
			"INSERT INTO frm_fields (field_id, site_id, key, type, label) VALUES ($1,$2,$3,$4,$5)",
			f.FieldID, siteID, f.Key, f.Type, f.Label)
		must(err)
	}
}

// ---------- frm_emails_log (upsert on (site_id, message_id)) ----------

func seedEmails(ctx context.Context, db *sqlx.DB, siteID int64, emails []emailSeed) {
	for _, r := range emails {
		_, err := db.ExecContext(ctx, `
			INSERT INTO frm_emails_log
				(entry_id, site_id, subject, message_id, email_from, email_to, content_plain, content_html, status, date_sent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (site_id, message_id) DO UPDATE SET
				subject=EXCLUDED.subject, email_from=EXCLUDED.email_from, email_to=EXCLUDED.email_to,
				status=EXCLUDED.status, updated_at=NOW()
		`, entryID, siteID, r.Subject, r.MessageID, r.From, r.To, r.ContentPlain, r.ContentHTML, r.Status, time.Now().UTC())
		must(err)
	}
}

// ---------- frm_entry_history (append; only insert if entryID has no rows yet) ----------

func seedHistory(ctx context.Context, db *sqlx.DB, siteID int64, history []historySeed) {
	if len(history) == 0 {
		return
	}

	var existing int64
	must(db.GetContext(ctx, &existing, "SELECT COUNT(*) FROM frm_entry_history WHERE site_id = $1 AND entry_id = $2", siteID, entryID))
	if existing > 0 {
		return
	}

	var fieldRows []int64
	must(db.SelectContext(ctx, &fieldRows, "SELECT id FROM frm_fields WHERE site_id = $1 ORDER BY id ASC", siteID))

	updateTypeIDs := map[string]int64{}
	rows, err := db.QueryxContext(ctx, "SELECT code, id FROM frm_entry_update_types")
	must(err)
	defer rows.Close()
	for rows.Next() {
		var code string
		var id int64
		must(rows.Scan(&code, &id))
		updateTypeIDs[code] = id
	}
	must(rows.Err())

	now := time.Now().UTC()
	for _, h := range history {
		if h.FieldIndex < 0 || h.FieldIndex >= len(fieldRows) {
			fail("history field_index %d out of range (have %d fields)", h.FieldIndex, len(fieldRows))
		}
		typeID, ok := updateTypeIDs[h.UpdateType]
		if !ok {
			fail("history update_type %q not seeded", h.UpdateType)
		}
		_, err := db.ExecContext(ctx, `
			INSERT INTO frm_entry_history (entry_id, site_id, field_id, update_type_id, old_value, new_value, change_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, entryID, siteID, fieldRows[h.FieldIndex], typeID, nullString(h.OldValue), nullString(h.NewValue), now.Add(-time.Duration(h.HoursAgo)*time.Hour))
		must(err)
	}
}

// ---------- helpers ----------

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

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
