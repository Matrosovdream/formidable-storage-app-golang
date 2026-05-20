package helpers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/app"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// SeedUserWithToken creates a user and issues a Sanctum-style personal access token.
// Returns the user id and the plaintext "<id>|<plain>" token.
func SeedUserWithToken(t *testing.T, deps *app.Deps) (userID int64, bearerToken string) {
	t.Helper()
	ctx := context.Background()

	email := fmt.Sprintf("http-test-%d-%s@example.com", time.Now().UnixNano(), randomHex(4))
	require.NoError(t, deps.DB.QueryRowxContext(ctx,
		"INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id",
		"HTTP Test", email, "$2a$04$abcdefghijklmnopqrstuv",
	).Scan(&userID))
	t.Cleanup(func() {
		_, _ = deps.DB.ExecContext(ctx, "DELETE FROM personal_access_tokens WHERE tokenable_id = $1", userID)
		_, _ = deps.DB.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	})

	plain := randomHex(40)
	hash := sha256.Sum256([]byte(plain))
	var tokenID int64
	require.NoError(t, deps.DB.QueryRowxContext(ctx,
		`INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		"App\\Models\\User", userID, "auth", hex.EncodeToString(hash[:]),
	).Scan(&tokenID))

	return userID, fmt.Sprintf("%d|%s", tokenID, plain)
}

// SeedSiteWithToken creates a site and an associated site_tokens row.
// Returns the site id and the plaintext bearer token used by /rest/v1/* middleware.
func SeedSiteWithToken(t *testing.T, db *sqlx.DB) (siteID int64, token string) {
	t.Helper()
	ctx := context.Background()
	url := fmt.Sprintf("https://site-%d-%s.test/", time.Now().UnixNano(), randomHex(4))
	require.NoError(t, db.QueryRowxContext(ctx, "INSERT INTO sites (name, url) VALUES ($1, $2) RETURNING id", "Test Site", url).Scan(&siteID))
	token = randomHex(32)
	_, err := db.ExecContext(ctx, "INSERT INTO site_tokens (site_id, token) VALUES ($1, $2)", siteID, token)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM sites WHERE id = $1", siteID)
	})
	return siteID, token
}

// SeedField inserts a field row scoped to the site and returns its id.
func SeedField(t *testing.T, db *sqlx.DB, siteID, fieldID int64, key, label string) int64 {
	t.Helper()
	var id int64
	require.NoError(t, db.QueryRowxContext(context.Background(),
		`INSERT INTO frm_fields (field_id, site_id, key, type, label) VALUES ($1,$2,$3,'text',$4) RETURNING id`,
		fieldID, siteID, key, label,
	).Scan(&id))
	return id
}

// SeedUpdateType inserts a frm_entry_update_types row (idempotent on code).
func SeedUpdateType(t *testing.T, db *sqlx.DB, code, title string) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	err := db.GetContext(ctx, &id, "SELECT id FROM frm_entry_update_types WHERE code = $1", code)
	if err == nil {
		return id
	}
	require.NoError(t, db.QueryRowxContext(ctx,
		"INSERT INTO frm_entry_update_types (code, title) VALUES ($1, $2) RETURNING id",
		code, title,
	).Scan(&id))
	return id
}

func randomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// NullString returns a sql.NullString matching the input.
func NullString(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
