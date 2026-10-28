package indicators_test

import (
	"context"
	"database/sql"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"testing"
)

const (
	lifecycleMigrationActorID     = "58000000-0000-4000-8000-000000000001"
	lifecycleMigrationIncidentID  = "58000000-0000-4000-8000-000000000002"
	lifecycleMigrationIndicatorID = "58000000-0000-4000-8000-000000000003"
	lifecycleMigrationSupportID   = "58000000-0000-4000-8000-000000000004"
	lifecycleMigrationIntervalID  = "58000000-0000-4000-8000-000000000005"
)

func TestIndicatorLifecycleIntegrityBaseline_Integration(t *testing.T) {

	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	seedIndicatorLifecycleMigrationFixture(t, db, false)
	requireIndicatorSupportRefsValidity(t, db, duplicateLifecycleSupportRefs(), false)
	expectIndicatorContractRejection(t, db, "indicator_state_intervals_support_refs_ck", `
 UPDATE indicator_state_intervals SET support_refs=jsonb_build_array($1::text,$1::text)
 WHERE indicator_state_interval_id=$2`, lifecycleMigrationSupportID, lifecycleMigrationIntervalID)
	// Overlapping observations are intentional; temporal non-overlap would change semantics.
	if _, err := db.ExecContext(t.Context(), `INSERT INTO indicator_state_intervals
 (incident_id,indicator_record_id,lifecycle_state,valid_from,support_refs,created_by_user_id)
 VALUES ($1,$2,'active',now(),'[]'::jsonb,$3)`, lifecycleMigrationIncidentID, lifecycleMigrationIndicatorID, lifecycleMigrationActorID); err != nil {
		t.Fatal(err)
	}

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
    'migration.example', 'domain_name:migration.example'
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

func duplicateLifecycleSupportRefs() string {
	return `["` + lifecycleMigrationSupportID + `","` + lifecycleMigrationSupportID + `"]`
}
