package indicators_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const (
	lifecycleMigrationActorID     = "58000000-0000-4000-8000-000000000001"
	lifecycleMigrationIncidentID  = "58000000-0000-4000-8000-000000000002"
	lifecycleMigrationIndicatorID = "58000000-0000-4000-8000-000000000003"
	lifecycleMigrationSupportID   = "58000000-0000-4000-8000-000000000004"
	lifecycleMigrationIntervalID  = "58000000-0000-4000-8000-000000000005"
)

func TestIndicatorLifecycleIntegrityMigration_Integration(t *testing.T) {
	t.Run("clean install enforces unique lifecycle support references", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseT(t)
		requireIndicatorMigrationHead(t, migrationDB.SQL(), 40)
		requireIndicatorSupportRefsValidity(t, migrationDB.SQL(), duplicateLifecycleSupportRefs(), false)
	})

	t.Run("valid 37 to 38 upgrade preserves data and supports down up", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 37)
		seedIndicatorLifecycleMigrationFixture(t, migrationDB.SQL(), false)
		ctx := context.Background()
		before := indicatorLifecycleSupportRefs(t, migrationDB.SQL())

		if err := migrationDB.ApplyThrough(ctx, 38); err != nil {
			t.Fatalf("apply Indicators lifecycle-integrity migration: %v", err)
		}
		requireIndicatorMigrationHead(t, migrationDB.SQL(), 38)
		requireIndicatorSupportRefsValidity(t, migrationDB.SQL(), duplicateLifecycleSupportRefs(), false)
		if got := indicatorLifecycleSupportRefs(t, migrationDB.SQL()); got != before {
			t.Fatalf("support_refs after upgrade = %s, want unchanged %s", got, before)
		}

		expectIndicatorContractRejection(t, migrationDB.SQL(), "indicator_state_intervals_support_refs_ck", `
INSERT INTO indicator_state_intervals (
    incident_id, indicator_record_id, lifecycle_state, valid_from,
    support_refs, created_by_user_id
) VALUES ($1, $2, 'active', now(), jsonb_build_array($3::text, $3::text), $4)
`, lifecycleMigrationIncidentID, lifecycleMigrationIndicatorID, lifecycleMigrationSupportID, lifecycleMigrationActorID)
		expectIndicatorContractRejection(t, migrationDB.SQL(), "indicator_state_intervals_support_refs_ck", `
UPDATE indicator_state_intervals
   SET support_refs = jsonb_build_array($1::text, $1::text)
 WHERE indicator_state_interval_id = $2
`, lifecycleMigrationSupportID, lifecycleMigrationIntervalID)

		if err := migrationDB.RollbackThrough(ctx, 37); err != nil {
			t.Fatalf("roll back Indicators lifecycle-integrity migration: %v", err)
		}
		requireIndicatorMigrationHead(t, migrationDB.SQL(), 37)
		requireIndicatorSupportRefsValidity(t, migrationDB.SQL(), duplicateLifecycleSupportRefs(), true)
		if got := indicatorLifecycleSupportRefs(t, migrationDB.SQL()); got != before {
			t.Fatalf("support_refs after Down = %s, want unchanged %s", got, before)
		}
		if err := migrationDB.ApplyThrough(ctx, 38); err != nil {
			t.Fatalf("reapply Indicators lifecycle-integrity migration: %v", err)
		}
		requireIndicatorMigrationHead(t, migrationDB.SQL(), 38)
		requireIndicatorSupportRefsValidity(t, migrationDB.SQL(), duplicateLifecycleSupportRefs(), false)
		if got := indicatorLifecycleSupportRefs(t, migrationDB.SQL()); got != before {
			t.Fatalf("support_refs after Up round trip = %s, want unchanged %s", got, before)
		}
	})

	t.Run("duplicate legacy references abort atomically without repair", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 37)
		seedIndicatorLifecycleMigrationFixture(t, migrationDB.SQL(), true)
		before := indicatorLifecycleSupportRefs(t, migrationDB.SQL())

		err := migrationDB.ApplyThrough(context.Background(), 38)
		var postgresError *pgconn.PgError
		if err == nil ||
			!strings.Contains(err.Error(), "indicators_lifecycle_integrity_preflight_failed") ||
			!errors.As(err, &postgresError) ||
			!strings.Contains(postgresError.Detail, "invalid_interval_rows=1") {
			t.Fatalf("lifecycle-integrity preflight error = %v, want one invalid row", err)
		}
		requireIndicatorMigrationHead(t, migrationDB.SQL(), 37)
		requireIndicatorSupportRefsValidity(t, migrationDB.SQL(), duplicateLifecycleSupportRefs(), true)
		if got := indicatorLifecycleSupportRefs(t, migrationDB.SQL()); got != before {
			t.Fatalf("support_refs after rejected upgrade = %s, want preserved %s", got, before)
		}
	})
}

func seedIndicatorLifecycleMigrationFixture(t testing.TB, db *sql.DB, duplicate bool) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash)
VALUES ($1, 'indicator-lifecycle-migration@example.test', 'Indicator Lifecycle Migration', 'not-used')
`, lifecycleMigrationActorID); err != nil {
		t.Fatalf("seed lifecycle migration actor: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, 'IR-IND-LIFECYCLE-MIGRATION', 'ir-ind-lifecycle-migration',
          'Indicator lifecycle migration', 'active', $2, $2)
`, lifecycleMigrationIncidentID, lifecycleMigrationActorID); err != nil {
		t.Fatalf("seed lifecycle migration incident: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id,
    updated_by_user_id, row_version
) VALUES
    ($1, $2, 'indicator', $3, $3, 1),
    ($4, $2, 'timeline_event', $3, $3, 1)
`, lifecycleMigrationIndicatorID, lifecycleMigrationIncidentID, lifecycleMigrationActorID, lifecycleMigrationSupportID); err != nil {
		t.Fatalf("seed lifecycle migration records: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO indicators (
    record_id, incident_id, indicator_type, value_kind,
    display_value, normalized_value, dedupe_key
) VALUES (
    $1, $2, 'domain_name', 'atomic', 'migration.example',
    'migration.example', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'
)
`, lifecycleMigrationIndicatorID, lifecycleMigrationIncidentID); err != nil {
		t.Fatalf("seed lifecycle migration Indicator: %v", err)
	}
	refCount := 1
	if duplicate {
		refCount = 2
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO indicator_state_intervals (
    indicator_state_interval_id, incident_id, indicator_record_id,
    lifecycle_state, valid_from, support_refs, created_by_user_id
) VALUES (
    $1, $2, $3, 'active', now(),
    CASE WHEN $5 = 1
         THEN jsonb_build_array($4::text)
         ELSE jsonb_build_array($4::text, $4::text)
    END,
    $6
)
`, lifecycleMigrationIntervalID, lifecycleMigrationIncidentID, lifecycleMigrationIndicatorID, lifecycleMigrationSupportID, refCount, lifecycleMigrationActorID); err != nil {
		t.Fatalf("seed lifecycle migration interval: %v", err)
	}
}

func requireIndicatorMigrationHead(t testing.TB, db *sql.DB, want int64) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(context.Background(), `
SELECT max(version_id)
  FROM goose_db_version
 WHERE is_applied
`).Scan(&got); err != nil {
		t.Fatalf("inspect migration head: %v", err)
	}
	if got != want {
		t.Fatalf("migration head = %d, want %d", got, want)
	}
}

func requireIndicatorSupportRefsValidity(t testing.TB, db *sql.DB, candidate string, want bool) {
	t.Helper()
	var got bool
	if err := db.QueryRowContext(context.Background(), `
SELECT public.indicator_support_refs_are_valid($1::jsonb)
`, candidate).Scan(&got); err != nil {
		t.Fatalf("evaluate Indicator support-ref helper: %v", err)
	}
	if got != want {
		t.Fatalf("support-ref helper(%s) = %t, want %t", candidate, got, want)
	}
}

func indicatorLifecycleSupportRefs(t testing.TB, db *sql.DB) string {
	t.Helper()
	var got string
	if err := db.QueryRowContext(context.Background(), `
SELECT support_refs::text
  FROM indicator_state_intervals
 WHERE indicator_state_interval_id = $1
`, lifecycleMigrationIntervalID).Scan(&got); err != nil {
		t.Fatalf("read lifecycle support refs: %v", err)
	}
	return got
}

func duplicateLifecycleSupportRefs() string {
	return `["` + lifecycleMigrationSupportID + `","` + lifecycleMigrationSupportID + `"]`
}
