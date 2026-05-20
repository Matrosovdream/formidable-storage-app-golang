package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Querier is satisfied by both *sqlx.DB and *sqlx.Tx so repositories can run
// either standalone or inside a transaction.
type Querier interface {
	GetContext(ctx context.Context, dest any, q string, args ...any) error
	SelectContext(ctx context.Context, dest any, q string, args ...any) error
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, q string, arg any) (sql.Result, error)
	QueryxContext(ctx context.Context, q string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, q string, args ...any) *sqlx.Row
	PrepareNamedContext(ctx context.Context, q string) (*sqlx.NamedStmt, error)
	Rebind(query string) string
}

// WithTx runs fn inside a transaction. On success it commits, on error/panic it rolls back.
func WithTx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	err = fn(tx)
	return
}

// BulkUpsert builds an INSERT ... ON CONFLICT (conflictCols) DO UPDATE SET ... statement.
//
//   table         - table name
//   columns       - ordered column names that match each row's value slice
//   rows          - slice of []any; each inner slice MUST have len(columns)
//   conflictCols  - columns forming the conflict target (must have a unique constraint)
//   updateCols    - columns to overwrite on conflict (typically all data columns except identity/key cols)
//
// updated_at is set to NOW() on update if "updated_at" appears in updateCols.
func BulkUpsert(ctx context.Context, q Querier, table string, columns []string, rows [][]any, conflictCols []string, updateCols []string) error {
	if len(rows) == 0 {
		return nil
	}
	if len(conflictCols) == 0 {
		return fmt.Errorf("bulk upsert: conflictCols required")
	}

	placeholders := make([]string, 0, len(rows))
	args := make([]any, 0, len(rows)*len(columns))
	pos := 1
	for _, row := range rows {
		if len(row) != len(columns) {
			return fmt.Errorf("bulk upsert: row has %d values but %d columns", len(row), len(columns))
		}
		marks := make([]string, len(columns))
		for i := range columns {
			marks[i] = fmt.Sprintf("$%d", pos)
			pos++
		}
		placeholders = append(placeholders, "("+strings.Join(marks, ", ")+")")
		args = append(args, row...)
	}

	updateClauses := make([]string, 0, len(updateCols))
	for _, c := range updateCols {
		if c == "updated_at" {
			updateClauses = append(updateClauses, "updated_at = NOW()")
			continue
		}
		updateClauses = append(updateClauses, fmt.Sprintf("%s = EXCLUDED.%s", c, c))
	}

	var stmt string
	if len(updateClauses) == 0 {
		stmt = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO NOTHING",
			table,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			strings.Join(conflictCols, ", "),
		)
	} else {
		stmt = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s",
			table,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			strings.Join(conflictCols, ", "),
			strings.Join(updateClauses, ", "),
		)
	}

	_, err := q.ExecContext(ctx, stmt, args...)
	return err
}

// ChunkRows splits rows into chunks no larger than size.
func ChunkRows[T any](rows []T, size int) [][]T {
	if size <= 0 {
		return [][]T{rows}
	}
	chunks := make([][]T, 0, (len(rows)+size-1)/size)
	for i := 0; i < len(rows); i += size {
		end := i + size
		if end > len(rows) {
			end = len(rows)
		}
		chunks = append(chunks, rows[i:end])
	}
	return chunks
}
