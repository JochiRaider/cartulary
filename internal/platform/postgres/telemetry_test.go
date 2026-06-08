package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestTelemetryDBPreservesDBBehaviorNoSDK(t *testing.T) {
	fake := &telemetryFakeDB{row: telemetryFakeRow{value: "ok"}}
	db := InstrumentDB(fake, "0.0.0+unknown")

	if _, err := db.Exec(t.Context(), "SELECT private_table WHERE id=$1", "secret"); err != nil {
		t.Fatalf("exec through telemetry wrapper: %v", err)
	}
	var value string
	if err := db.QueryRow(t.Context(), "SELECT private_column", "secret").Scan(&value); err != nil {
		t.Fatalf("query row through telemetry wrapper: %v", err)
	}
	if value != "ok" {
		t.Fatalf("unexpected scan value: %q", value)
	}
	if fake.execSQL == "" || fake.queryRowSQL == "" {
		t.Fatalf("wrapper did not call fake DB: %#v", fake)
	}
}

func TestPostgresErrorClass(t *testing.T) {
	if got := postgresErrorClass(context.DeadlineExceeded); got != "timeout" {
		t.Fatalf("unexpected deadline class: %q", got)
	}
	if got := postgresErrorClass(&pgconn.PgError{Code: "40001"}); got != "serialization_conflict" {
		t.Fatalf("unexpected serialization class: %q", got)
	}
	if got := postgresErrorClass(&pgconn.PgError{Code: "23505"}); got != "constraint_violation" {
		t.Fatalf("unexpected constraint class: %q", got)
	}
	if got := postgresErrorClass(errors.New("network unavailable")); got != "dependency_unavailable" {
		t.Fatalf("unexpected dependency class: %q", got)
	}
}

type telemetryFakeDB struct {
	row         telemetryFakeRow
	execSQL     string
	querySQL    string
	queryRowSQL string
}

func (db *telemetryFakeDB) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	db.execSQL = sql
	return pgconn.NewCommandTag("SELECT 1"), nil
}

func (db *telemetryFakeDB) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	db.querySQL = sql
	return nil, nil
}

func (db *telemetryFakeDB) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	db.queryRowSQL = sql
	return db.row
}

func (db *telemetryFakeDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, nil
}

type telemetryFakeRow struct {
	value string
	err   error
}

func (row telemetryFakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(dest) == 1 {
		if target, ok := dest[0].(*string); ok {
			*target = row.value
		}
	}
	return nil
}
