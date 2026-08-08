package database_migrations

import (
	"context"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTargetedOperationValidation(t *testing.T) {
	missing := NewMigrationSource("/path/that/must/not/be-inspected")
	if _, err := ApplyThrough(context.Background(), nil, missing, 0); err == nil || err.Error() != "migration apply-through version must be positive" {
		t.Fatalf("unexpected apply-through validation error: %v", err)
	}
	if _, err := RollbackThrough(context.Background(), nil, missing, -1); err == nil || err.Error() != "migration rollback-through version must be non-negative" {
		t.Fatalf("unexpected rollback-through validation error: %v", err)
	}

	empty := NewEmbeddedMigrationSource(fstest.MapFS{}, ".", "empty")
	status, err := RollbackThrough(context.Background(), nil, empty, 0)
	if err != nil {
		t.Fatalf("rollback-through zero over an empty source: %v", err)
	}
	if status.SourceName != "empty" || !status.Empty {
		t.Fatalf("unexpected empty rollback status: %#v", status)
	}
}

func TestPublicSurfaceAndLedgerReaderCapability(t *testing.T) {
	readerType := reflect.TypeOf((*LedgerReader)(nil)).Elem()
	if readerType.NumMethod() != 2 {
		t.Fatalf("LedgerReader method count = %d, want 2", readerType.NumMethod())
	}
	for index, name := range []string{"Query", "QueryRow"} {
		if got := readerType.Method(index).Name; got != name {
			t.Fatalf("LedgerReader method %d = %q, want %q", index, got, name)
		}
	}

	statusType := reflect.TypeOf(MigrationStatus{})
	if statusType.NumField() != 2 || statusType.Field(0).Name != "SourceName" || statusType.Field(1).Name != "Empty" {
		t.Fatalf("unexpected MigrationStatus surface: %#v", statusType)
	}
}

var _ LedgerReader = (*pgxpool.Pool)(nil)
