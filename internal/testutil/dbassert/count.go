package dbassert

import (
	"context"
	"database/sql"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func CountSQL(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func CountPostgres(t testing.TB, db postgres.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
