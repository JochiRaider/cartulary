package database_migrations_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIncidentMembershipBaseline_Integration(t *testing.T) {
	database := pgtest.Start(t).MigrationDatabaseThroughT(t, 3)
	db := database.SQL()
	assertCount(t, db, `SELECT count(*) FROM public.incident_membership_versions`, 0)
	assertCount(t, db, `
SELECT count(*) FROM pg_catalog.pg_constraint
WHERE conrelid = 'public.incident_memberships'::regclass
  AND conname IN ('incident_memberships_version_history_fkey', 'incident_memberships_version_safe_ck')
`, 2)
	assertCount(t, db, `
SELECT count(*) FROM pg_catalog.pg_indexes
WHERE schemaname = 'public' AND tablename = 'incident_memberships'
  AND indexdef LIKE '% USING btree (incident_id, joined_at, user_id)'
`, 1)
}
