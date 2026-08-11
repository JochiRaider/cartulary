package sourcecatalog

import (
	"database/sql"
	"testing"
	"testing/fstest"

	"github.com/pressly/goose/v3"
	gooselock "github.com/pressly/goose/v3/lock"
)

var _ func(*sql.DB, *Catalog, gooselock.SessionLocker) (*goose.Provider, error) = NewProvider

func TestNewProviderContract(t *testing.T) {
	catalog, err := Build(
		fstest.MapFS{
			"migrations/00001_valid.sql": &fstest.MapFile{
				Data: []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n"),
			},
		},
		"migrations",
		"lineage",
		"boundary",
	)
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	locker, err := NewSessionLocker()
	if err != nil {
		t.Fatalf("construct session locker: %v", err)
	}

	tests := []struct {
		name    string
		db      *sql.DB
		catalog *Catalog
		locker  gooselock.SessionLocker
		wantErr string
	}{
		{name: "nil database", catalog: catalog, locker: locker, wantErr: "migration provider database is nil"},
		{name: "nil catalog", db: &sql.DB{}, locker: locker, wantErr: "migration source is invalid"},
		{name: "nil locker", db: &sql.DB{}, catalog: catalog, wantErr: "migration provider session locker is nil"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, providerErr := NewProvider(test.db, test.catalog, test.locker)
			if providerErr == nil || providerErr.Error() != test.wantErr {
				t.Fatalf("provider error = %v, want %q", providerErr, test.wantErr)
			}
		})
	}

	provider, err := NewProvider(&sql.DB{}, catalog, locker)
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	sources := provider.ListSources()
	if len(sources) != 1 || sources[0].Version != 1 {
		t.Fatalf("provider sources = %#v", sources)
	}
}

func TestSessionLockerPolicy(t *testing.T) {
	if migrationAdvisoryLockID != 4097083626 {
		t.Fatalf("migration advisory lock id = %d", migrationAdvisoryLockID)
	}
	locker, err := NewSessionLocker()
	if err != nil {
		t.Fatalf("construct session locker: %v", err)
	}
	if locker == nil {
		t.Fatal("session locker is nil")
	}
}
