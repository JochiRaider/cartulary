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

func TestIndicatorEnvelopeContractMigration57DownUp_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 56)
	db := migrationDB.SQL()
	ctx := context.Background()

	const (
		actorID     = "57000000-0000-4000-8000-000000000001"
		incidentID  = "57000000-0000-4000-8000-000000000002"
		indicatorID = "57000000-0000-4000-8000-000000000003"
		sourceID    = "57000000-0000-4000-8000-000000000004"
		otherIncID  = "57000000-0000-4000-8000-000000000007"
		otherSrcID  = "57000000-0000-4000-8000-000000000008"
		dedupeKey   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash)
VALUES ($1, 'indicator-contract@example.test', 'Indicator Contract', 'not-used')
`, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, 'IR-IND-CONTRACT', 'ir-ind-contract', 'Indicator contract', 'active', $2, $2)
`, incidentID, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, 'IR-IND-CONTRACT-OTHER', 'ir-ind-contract-other', 'Other incident', 'active', $2, $2)
`, otherIncID, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id,
    updated_by_user_id, row_version
)
VALUES
    ($1, $2, 'indicator', $3, $3, 3),
    ($4, $2, 'timeline_event', $3, $3, 2)
`, indicatorID, incidentID, actorID, sourceID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id,
    updated_by_user_id, row_version
)
VALUES ($1, $2, 'timeline_event', $3, $3, 1)
`, otherSrcID, otherIncID, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO indicators (
    record_id, incident_id, indicator_type, value_kind, display_value,
    normalized_value, dedupe_key, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id
)
SELECT
    record_id, incident_id, 'domain_name', 'atomic', 'contract.example',
    'contract.example', $2, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id
  FROM records
 WHERE record_id = $1
`, indicatorID, dedupeKey); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO indicator_observations (
    indicator_observation_id, incident_id, source_record_id, source_field_key,
    origin_kind, origin_locator, observed_text, parsed_indicator_type,
    normalized_candidate, resolution_status, created_by_user_id
)
VALUES (
    '57000000-0000-4000-8000-000000000005', $1, $2, 'timeline.raw_activity_text',
    'manual_entry', 'record://source/field/0:16', 'contract.example', 'domain_name',
    'contract.example', 'unresolved', $3
)
`, incidentID, sourceID, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO indicator_state_intervals (
    indicator_state_interval_id, incident_id, indicator_record_id,
    lifecycle_state, valid_from, support_refs, created_by_user_id
)
VALUES (
    '57000000-0000-4000-8000-000000000006', $1, $2,
    'active', now(), jsonb_build_array($3::text), $4
)
`, incidentID, indicatorID, sourceID, actorID); err != nil {
		t.Fatal(err)
	}

	if err := migrationDB.ApplyThrough(ctx, 57); err != nil {
		t.Fatalf("apply Indicator contract migration: %v", err)
	}
	requireIndicatorMirrorColumns(t, db, false)
	requireIndicatorContractConstraintsValidated(t, db)

	if err := migrationDB.RollbackThrough(ctx, 56); err != nil {
		t.Fatalf("reconstruct expand schema: %v", err)
	}
	requireIndicatorMirrorColumns(t, db, true)
	var mirroredVersion int64
	if err := db.QueryRowContext(ctx, `
SELECT indicator.row_version
  FROM indicators AS indicator
  JOIN records AS envelope ON envelope.record_id = indicator.record_id
 WHERE indicator.record_id = $1
   AND indicator.row_version = envelope.row_version
   AND indicator.created_at = envelope.created_at
   AND indicator.updated_at = envelope.updated_at
   AND indicator.created_by_user_id = envelope.created_by_user_id
   AND indicator.updated_by_user_id = envelope.updated_by_user_id
   AND indicator.deleted_at IS NOT DISTINCT FROM envelope.deleted_at
   AND indicator.deleted_by_user_id IS NOT DISTINCT FROM envelope.deleted_by_user_id
`, indicatorID).Scan(&mirroredVersion); err != nil {
		t.Fatalf("load reconstructed mirrors: %v", err)
	}
	if mirroredVersion != 3 {
		t.Fatalf("reconstructed row version = %d, want 3", mirroredVersion)
	}

	if err := migrationDB.ApplyThrough(ctx, 57); err != nil {
		t.Fatalf("reapply Indicator contract migration: %v", err)
	}
	requireIndicatorMirrorColumns(t, db, false)
	requireIndicatorContractConstraintsValidated(t, db)

	expectIndicatorContractRejection(t, db, "indicators_indicator_type_ck", `
INSERT INTO indicators (
    record_id, incident_id, indicator_type, value_kind, display_value, dedupe_key
) VALUES (
    '57000000-0000-4000-8000-000000000011', $1, 'legacy_domain', 'atomic',
    'invalid.example', 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'
)
`, incidentID)
	expectIndicatorContractRejection(t, db, "indicator_state_intervals_lifecycle_state_ck", `
INSERT INTO indicator_state_intervals (
    incident_id, indicator_record_id, lifecycle_state, valid_from,
    support_refs, created_by_user_id
) VALUES ($1, $2, 'unknown', now(), '[]'::jsonb, $3)
`, incidentID, indicatorID, actorID)
	expectIndicatorContractRejection(t, db, "indicator_observations_resolution_tuple_ck", `
INSERT INTO indicator_observations (
    incident_id, source_record_id, source_field_key, origin_kind,
    origin_locator, observed_text, resolution_status,
    resolved_indicator_record_id, created_by_user_id
) VALUES (
    $1, $2, 'timeline.raw_activity_text', 'manual_entry',
    'record://source/field/0:1', 'x', 'resolved', $3, $4
)
`, incidentID, sourceID, indicatorID, actorID)
	expectIndicatorContractRejection(t, db, "indicator lifecycle support reference is outside the incident", `
INSERT INTO indicator_state_intervals (
    incident_id, indicator_record_id, lifecycle_state, valid_from,
    support_refs, created_by_user_id
) VALUES ($1, $2, 'active', now(), jsonb_build_array($3::text), $4)
`, incidentID, indicatorID, otherSrcID, actorID)
	expectIndicatorContractRejection(t, db, "indicator_state_intervals_support_refs_ck", `
INSERT INTO indicator_state_intervals (
    incident_id, indicator_record_id, lifecycle_state, valid_from,
    support_refs, created_by_user_id
) VALUES ($1, $2, 'active', now(), 'null'::jsonb, $3)
`, incidentID, indicatorID, actorID)
}

func requireIndicatorMirrorColumns(t testing.TB, db *sql.DB, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'indicators'
   AND column_name = ANY($1::text[])
`, []string{
		"row_version", "created_at", "updated_at", "created_by_user_id",
		"updated_by_user_id", "deleted_at", "deleted_by_user_id",
	}).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count == 7; got != want {
		t.Fatalf("all Indicator mirror columns present = %t (count %d), want %t", got, count, want)
	}
}

func requireIndicatorContractConstraintsValidated(t testing.TB, db *sql.DB) {
	t.Helper()
	var invalidCount int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM pg_constraint
 WHERE connamespace = 'public'::regnamespace
   AND conname = ANY($1::text[])
   AND NOT convalidated
`, []string{
		"indicator_active_identities_indicator_fkey",
		"indicator_active_identities_record_envelope_fkey",
		"indicators_indicator_type_ck",
		"indicator_observations_resolution_tuple_ck",
		"indicator_state_intervals_lifecycle_state_ck",
		"indicator_state_intervals_support_refs_ck",
	}).Scan(&invalidCount); err != nil {
		t.Fatal(err)
	}
	if invalidCount != 0 {
		t.Fatalf("unvalidated Indicator contract constraints = %d", invalidCount)
	}
}

func expectIndicatorContractRejection(t testing.TB, db *sql.DB, want string, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err == nil {
		t.Fatal("invalid Indicator contract row unexpectedly succeeded")
	} else {
		var postgresError *pgconn.PgError
		constraint := ""
		if errors.As(err, &postgresError) {
			constraint = postgresError.ConstraintName
		}
		if !strings.Contains(err.Error(), want) && constraint != want {
			t.Fatalf("contract rejection = %v, constraint = %q, want %q", err, constraint, want)
		}
	}
}
