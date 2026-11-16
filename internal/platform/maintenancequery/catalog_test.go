package maintenancequery

import (
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func TestMaintenanceSQLCapabilities(t *testing.T) {
	refused := func(t *testing.T, err error, want string) {
		t.Helper()
		if err == nil || err.Error() != want {
			t.Fatalf("error = %v, want %q", err, want)
		}
	}
	for _, set := range []struct {
		name  string
		build func([]string) (sqlshape.Statement, error)
	}{{"LockMutable", LockMutable}, {"TruncateRestore", TruncateRestore}, {"TruncateVerification", TruncateVerification}} {
		for _, names := range [][]string{nil, {"users"}, {"users", "users"}, {"recovery_target_state"}} {
			_, err := set.build(names)
			refused(t, err, "unadmitted "+set.name+" relation set")
		}
	}

	for _, name := range []string{"future_unadmitted_table", "users; DELETE FROM records", "public.users"} {
		_, err := From(name)
		refused(t, err, fmt.Sprintf("unadmitted maintenance relation %q", name))
		_, err = Insert(name, sqlshape.Identifier("id"), false)
		refused(t, err, fmt.Sprintf("unadmitted restore relation %q", name))
	}
	for _, name := range []string{"goose_db_version", "schema_migration_lineage", "recovery_target_state"} {
		if _, err := From(name); err != nil {
			t.Fatal(err)
		}
		_, err := Delete(name)
		refused(t, err, fmt.Sprintf("unadmitted Delete relation %q", name))
	}
	_, err := Insert("timeline_grid_projection", sqlshape.Identifier("record_id"), false)
	refused(t, err, `unadmitted restore relation "timeline_grid_projection"`)
	if _, err := LockShare("recovery_target_state"); err != nil {
		t.Fatal(err)
	}
	_, err = LockShare("users")
	refused(t, err, `unadmitted LockShare relation "users"`)
	query, err := Insert("users", sqlshape.Join([]sqlshape.Fragment{sqlshape.Identifier("id"), sqlshape.Identifier("email")}, sqlshape.Literal(",")), true)
	if err != nil {
		t.Fatal(err)
	}
	text := sqlshape.CommandText(query)
	if !strings.Contains(text, `INSERT INTO "public"."users" ( "id" , "email" )`) || !strings.Contains(text, "IS NOT DISTINCT FROM ($1::jsonb-$2::text[])") {
		t.Fatal(text)
	}
}
