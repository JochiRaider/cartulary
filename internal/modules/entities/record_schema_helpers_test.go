package entities_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func requireColumns(t testing.TB, db postgres.DB, table string, columns ...string) {
	t.Helper()
	for _, column := range columns {
		assertCount(t, db, `
SELECT COUNT(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = $2
`, 1, table, column)
	}
}

func requireNoColumns(t testing.TB, db postgres.DB, column string, tables ...string) {
	t.Helper()
	for _, table := range tables {
		assertCount(t, db, `
SELECT COUNT(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = $2
`, 0, table, column)
	}
}

func requireGenericRecordFKsDoNotTargetTimeline(t testing.TB, db postgres.DB) {
	t.Helper()
	assertCount(t, db, `
SELECT COUNT(*)
  FROM pg_constraint c
  JOIN pg_class source_table ON source_table.oid = c.conrelid
  JOIN pg_class target_table ON target_table.oid = c.confrelid
 WHERE c.contype = 'f'
   AND source_table.relname IN ('record_revisions', 'record_links', 'record_tags', 'entity_mentions', 'indicator_observations')
   AND target_table.relname = 'timeline_events'
`, 0)
}

func assertCount(t testing.TB, db postgres.DB, query string, want int, args ...any) {
	t.Helper()
	got := appsupport.QueryCount(t, db, query, args...)
	if got != want {
		t.Fatalf("unexpected count: got %d want %d\n%s", got, want, query)
	}
}
