package pgtest

import (
	"context"
	"database/sql"
	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"reflect"
	"testing"
)

func TestMigrationDatabaseTargetedOperationValidation(t *testing.T) {
	dbOps := defaultDatabaseOperations()
	calls, sourceLoads := 0, 0
	dbOps.source = func() (*database_migrations.Source, error) { sourceLoads++; return dbmigrations.Source() }
	canonical, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	want, err := database_migrations.InspectSource(canonical)
	if err != nil {
		t.Fatal(err)
	}
	operation := func(_ context.Context, db *sql.DB, source *database_migrations.Source, version int64) error {
		calls++
		got, err := database_migrations.InspectSource(source)
		if err != nil || !reflect.DeepEqual(got, want) || db == nil {
			t.Fatal("targeted operation lost canonical source or scratch handle", err)
		}
		if (calls == 1 && version != 1) || (calls == 2 && version != 0) {
			t.Fatal("wrong target", version)
		}
		return nil
	}
	dbOps.applyThrough = operation
	dbOps.rollbackThrough = operation
	for _, capability := range []*MigrationDatabase{nil, {}, {db: &sql.DB{}}, {db: &sql.DB{}, identity: &migrationDatabaseIdentity{}, operations: dbOps}} {
		if borrowed, err := capability.RuntimeDatabase(); err == nil || borrowed != nil {
			t.Fatal("unissued capability exposed runtime database")
		}
		if capability.SQL() != nil {
			t.Fatal("unissued capability exposed SQL")
		}
		if err := capability.ApplyThrough(t.Context(), 1); err == nil {
			t.Fatal("unissued apply accepted")
		}
		if err := capability.RollbackThrough(t.Context(), 0); err == nil {
			t.Fatal("unissued rollback accepted")
		}
	}
	issued := &MigrationDatabase{db: &sql.DB{}, identity: issuedMigrationDatabaseIdentity, operations: dbOps}
	if _, err := issued.RuntimeDatabase(); err == nil {
		t.Fatal("missing scratch identity exposed runtime database")
	}
	issued.database = TestDatabase{Name: "scratch", DSN: "test-only"}
	borrowed, err := issued.RuntimeDatabase()
	if err != nil || *borrowed != issued.database {
		t.Fatal("borrowed database lost scratch identity", err)
	}
	borrowed.Name = "changed"
	if issued.database.Name != "scratch" {
		t.Fatal("borrower changed issuing capability identity")
	}
	for _, version := range []int64{-1, 0} {
		if err := issued.ApplyThrough(t.Context(), version); err == nil {
			t.Fatal("invalid apply target accepted")
		}
	}
	if err := issued.RollbackThrough(t.Context(), -1); err == nil {
		t.Fatal("negative rollback accepted")
	}
	if calls != 0 || sourceLoads != 0 {
		t.Fatal("rejected operation reached source/database execution")
	}
	if issued.SQL() != issued.db {
		t.Fatal("issued capability lost handle")
	}
	if err := issued.ApplyThrough(t.Context(), 1); err != nil {
		t.Fatal(err)
	}
	if err := issued.RollbackThrough(t.Context(), 0); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || sourceLoads != 2 {
		t.Fatal("legal targeted operations not executed", calls)
	}
}
