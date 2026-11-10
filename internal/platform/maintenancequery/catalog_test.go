package maintenancequery

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func TestMaintenanceSQLCapabilities(t *testing.T) {
	for _, build := range []func([]string) (sqlshape.Statement, error){LockMutable, TruncateRestore, TruncateVerification} {
		for _, names := range [][]string{nil, {"users"}, {"users", "users"}, {"recovery_target_state"}} {
			if _, err := build(names); err == nil {
				t.Fatalf("partial or protected set admitted: %v", names)
			}
		}
	}

	for _, name := range []string{"future_unadmitted_table", "users; DELETE FROM records", "public.users"} {
		if _, err := From(name); err == nil {
			t.Fatalf("read accepted unadmitted relation %q", name)
		}
		if _, err := Insert(name, sqlshape.Identifier("id"), false); err == nil {
			t.Fatalf("write accepted unadmitted relation %q", name)
		}
	}
	for _, name := range []string{"goose_db_version", "schema_migration_lineage", "recovery_target_state"} {
		if _, err := From(name); err != nil {
			t.Fatal(err)
		}
		for _, build := range []func(string) (sqlshape.Statement, error){Delete} {
			if _, err := build(name); err == nil {
				t.Fatalf("mutation accepted protected relation %q", name)
			}
		}
	}
	if _, err := Insert("timeline_grid_projection", sqlshape.Identifier("record_id"), false); err == nil {
		t.Fatal("derived projections cannot be restored as authoritative rows")
	}
	if _, err := LockShare("recovery_target_state"); err != nil {
		t.Fatal(err)
	}
	if _, err := LockShare("users"); err == nil {
		t.Fatal("control read lock widened")
	}
	query, err := Insert("users", sqlshape.Join([]sqlshape.Fragment{sqlshape.Identifier("id"), sqlshape.Identifier("email")}, sqlshape.Literal(",")), true)
	if err != nil {
		t.Fatal(err)
	}
	text := sqlshape.CommandText(query)
	if !strings.Contains(text, `INSERT INTO "public"."users" ( "id" , "email" )`) || !strings.Contains(text, "IS NOT DISTINCT FROM ($1::jsonb-$2::text[])") {
		t.Fatal(text)
	}
}
