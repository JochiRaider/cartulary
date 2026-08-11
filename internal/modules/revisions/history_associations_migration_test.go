package revisions

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRevisionsHistoryAssociationsHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "revisions-history-associations-head", postgres.PurposeRuntime)

	var columns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'change_set_mutations'
   AND column_name IN ('history_record_ids', 'history_entry_record_ids')
`).Scan(&columns); err != nil || columns != 2 {
		t.Fatalf("history association columns = %d, err = %v", columns, err)
	}
	var indexMethod string
	if err := db.QueryRowContext(context.Background(), `
SELECT access_method.amname
  FROM pg_index index_row
  JOIN pg_class index_relation ON index_relation.oid = index_row.indexrelid
  JOIN pg_am access_method ON access_method.oid = index_relation.relam
 WHERE index_relation.relname = 'change_set_mutations_history_record_ids_idx'
`).Scan(&indexMethod); err != nil || indexMethod != "gin" {
		t.Fatalf("history lookup index method = %q, err = %v", indexMethod, err)
	}

	var canonical, duplicate, unsorted, zero bool
	if err := db.QueryRowContext(context.Background(), `
SELECT
    change_set_mutations_history_ids_are_canonical(
        ARRAY['10000000-0000-4000-8000-000000000001'::uuid, '20000000-0000-4000-8000-000000000002'::uuid]
    ),
    change_set_mutations_history_ids_are_canonical(
        ARRAY['10000000-0000-4000-8000-000000000001'::uuid, '10000000-0000-4000-8000-000000000001'::uuid]
    ),
    change_set_mutations_history_ids_are_canonical(
        ARRAY['20000000-0000-4000-8000-000000000002'::uuid, '10000000-0000-4000-8000-000000000001'::uuid]
    ),
    change_set_mutations_history_ids_are_canonical(
        ARRAY['00000000-0000-0000-0000-000000000000'::uuid]
    )
`).Scan(&canonical, &duplicate, &unsorted, &zero); err != nil {
		t.Fatalf("evaluate history array invariants: %v", err)
	}
	if !canonical || duplicate || unsorted || zero {
		t.Fatalf("history array invariant results = canonical:%v duplicate:%v unsorted:%v zero:%v", canonical, duplicate, unsorted, zero)
	}
	var conflictTable bool
	if err := db.QueryRowContext(context.Background(), `
SELECT to_regclass('public.record_revision_conflict_facts') IS NOT NULL
`).Scan(&conflictTable); err != nil || !conflictTable {
		t.Fatalf("record revision conflict facts table present = %v, err = %v", conflictTable, err)
	}
}
