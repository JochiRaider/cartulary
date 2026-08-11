package jobs_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestJobsExpiryHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "jobs-expiry-head-contract", postgres.PurposeRuntime)
	var columns, indexes int
	if err := db.QueryRowContext(context.Background(), `
SELECT (SELECT count(*) FROM information_schema.columns
         WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'expired_at'),
       (SELECT count(*) FROM pg_indexes
         WHERE schemaname = 'public' AND tablename = 'jobs' AND indexname = 'jobs_expiry_candidates_idx')
`).Scan(&columns, &indexes); err != nil {
		t.Fatal(err)
	}
	if columns != 1 || indexes != 1 {
		t.Fatalf("jobs expiry shape = columns %d indexes %d; want 1/1", columns, indexes)
	}
}
