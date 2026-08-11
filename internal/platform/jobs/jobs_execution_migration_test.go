package jobs_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestJobsExecutionHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "jobs-execution-head-contract", postgres.PurposeRuntime)
	var columns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'jobs'
   AND column_name IN ('handler_attempt_id', 'handler_failure_count', 'handler_next_attempt_at')
`).Scan(&columns); err != nil {
		t.Fatal(err)
	}
	if columns != 3 {
		t.Fatalf("jobs execution columns = %d want 3", columns)
	}
}
