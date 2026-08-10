package revisions

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRevisionsHistoryAssociationsMigration61FreshAndGuarded_Integration(t *testing.T) {
	t.Run("fresh", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 60)
		db := migrationDB.SQL()
		if err := migrationDB.ApplyThrough(context.Background(), 61); err != nil {
			t.Fatalf("apply migration 61: %v", err)
		}

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
	})

	t.Run("existing history requires reset", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 60)
		db := migrationDB.SQL()
		ctx := context.Background()
		if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash)
VALUES ('61000000-0000-4000-8000-000000000001', 'revisions-migration@example.test', 'Revisions Migration', 'not-used');
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES (
    '61000000-0000-4000-8000-000000000002', 'IR-REV-061', 'ir-rev-061',
    'Revisions migration', 'active',
    '61000000-0000-4000-8000-000000000001', '61000000-0000-4000-8000-000000000001'
);
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source)
VALUES (
    '61000000-0000-4000-8000-000000000003',
    '61000000-0000-4000-8000-000000000002',
    '61000000-0000-4000-8000-000000000001',
    'revisions.migration.fixture'
);
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind
) VALUES (
    '61000000-0000-4000-8000-000000000003', 1, 'record',
    '61000000-0000-4000-8000-000000000004', 'fixture'
);
`); err != nil {
			t.Fatalf("seed pre-remediation history: %v", err)
		}

		err := migrationDB.ApplyThrough(ctx, 61)
		if err == nil || !strings.Contains(err.Error(), "requires a pre-production database reset") ||
			!strings.Contains(err.Error(), "change_set_mutations is not empty") {
			t.Fatalf("reset-required diagnostic = %v", err)
		}
		var columns int
		if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'change_set_mutations'
   AND column_name IN ('history_record_ids', 'history_entry_record_ids')
`).Scan(&columns); err != nil || columns != 0 {
			t.Fatalf("failed migration changed schema: columns = %d, err = %v", columns, err)
		}
	})
}
