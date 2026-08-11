package jobs_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestJobsDefinitionHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "jobs-definition-head-contract", postgres.PurposeRuntime)
	assertJobsDefinitionHeadShape(t, db)
}

func assertJobsDefinitionHeadShape(t testing.TB, db *sql.DB) {
	t.Helper()
	var columns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'jobs'
   AND column_name IN ('job_kind', 'progress_unit_id')
`).Scan(&columns); err != nil {
		t.Fatal(err)
	}
	if columns != 2 {
		t.Fatalf("jobs definition columns = %d, want 2", columns)
	}

	var constraints int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM pg_catalog.pg_constraint
 WHERE connamespace = 'public'::regnamespace
   AND conrelid = 'public.jobs'::regclass
   AND convalidated
   AND conname IN ('jobs_job_kind_nonempty_ck', 'jobs_progress_unit_id_shape_ck')
`).Scan(&constraints); err != nil {
		t.Fatal(err)
	}
	if constraints != 2 {
		t.Fatalf("validated jobs definition constraints = %d, want 2", constraints)
	}
}
