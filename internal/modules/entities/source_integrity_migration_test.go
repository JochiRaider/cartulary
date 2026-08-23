package entities_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEntitySourceIntegrityMigration_Integration(t *testing.T) {
	t.Run("valid 35 to 36 upgrade enforces source tuples and supports disposable rollback", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 35)
		fixture := seedSourceIntegrityFixture(t, migrationDB.SQL())
		ctx := context.Background()

		if err := migrationDB.ApplyThrough(ctx, 36); err != nil {
			t.Fatalf("apply Entities source-integrity migration: %v", err)
		}
		assertEntitySourceIntegrityObjects(t, migrationDB.SQL(), true)
		assertEntitySourceIntegrityFunctionPrivileges(t, migrationDB.SQL())

		if _, err := migrationDB.SQL().ExecContext(ctx, `
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind,
    origin_locator, raw_text, normalized_text, resolution_status,
    row_version, ordinal, created_by_user_id
) VALUES ($1, 'host', 'timeline.host_refs', 'legacy', 'fixture',
          'Legacy host', 'legacy host', 'unresolved', 1, 1, $2)
`, fixture.hostID, fixture.actorID); err == nil || !strings.Contains(err.Error(), "entity_mentions_origin_kind_ck") {
			t.Fatalf("invalid origin error = %v, want entity_mentions_origin_kind_ck", err)
		}

		if err := migrationDB.RollbackThrough(ctx, 35); err != nil {
			t.Fatalf("roll back Entities source-integrity migration: %v", err)
		}
		assertEntitySourceIntegrityObjects(t, migrationDB.SQL(), false)
		if err := migrationDB.ApplyThrough(ctx, 36); err != nil {
			t.Fatalf("reapply Entities source-integrity migration: %v", err)
		}
		assertEntitySourceIntegrityObjects(t, migrationDB.SQL(), true)
	})

	invalidFixtures := []struct {
		name       string
		detailPart string
		mutate     func(testing.TB, *sql.DB, sourceIntegrityFixture)
	}{
		{
			name:       "host envelope type",
			detailPart: "hosts=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				recordID := uuid.New()
				seedSourceIntegrityEnvelope(t, db, fixture.incidentID, fixture.actorID, recordID, "identity")
				insertSourceIntegrityHost(t, db, fixture.incidentID, recordID, "canonical", nil, 1)
			},
		},
		{
			name:       "identity envelope mirror",
			detailPart: "identities=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				recordID := uuid.New()
				seedSourceIntegrityEnvelope(t, db, fixture.incidentID, fixture.actorID, recordID, "identity")
				if _, err := db.ExecContext(context.Background(), `
INSERT INTO identities (
    record_id, incident_id, display_name, identity_state, row_version,
    created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT $1, $2, 'Mismatched identity', 'canonical', 2,
       r.created_at, r.updated_at, r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
`, recordID, fixture.incidentID); err != nil {
					t.Fatalf("seed mismatched identity: %v", err)
				}
			},
		},
		{
			name:       "merge target incident",
			detailPart: "hosts=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				otherIncidentID := uuid.New()
				seedSourceIntegrityIncident(t, db, otherIncidentID, fixture.actorID, "IR-ESI-OTHER")
				targetID := uuid.New()
				seedSourceIntegrityEnvelope(t, db, otherIncidentID, fixture.actorID, targetID, "host")
				insertSourceIntegrityHost(t, db, otherIncidentID, targetID, "canonical", nil, 1)
				loserID := uuid.New()
				seedSourceIntegrityEnvelope(t, db, fixture.incidentID, fixture.actorID, loserID, "host")
				insertSourceIntegrityHost(t, db, fixture.incidentID, loserID, "merged", &targetID, 1)
			},
		},
		{
			name:       "alias owner",
			detailPart: "aliases=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				insertSourceIntegrityAlias(t, db, fixture, "identity", "Alias")
			},
		},
		{
			name:       "alias normalization",
			detailPart: "aliases=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				insertSourceIntegrityAlias(t, db, fixture, "host", " padded alias ")
			},
		},
		{
			name:       "preserved identifier class",
			detailPart: "preserved_identifiers=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				insertSourceIntegrityPreservedIdentifier(t, db, fixture, "unknown", "value", "value")
			},
		},
		{
			name:       "preserved identifier normalization",
			detailPart: "preserved_identifiers=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				insertSourceIntegrityPreservedIdentifier(t, db, fixture, "hostname", "Example.Host", "wrong")
			},
		},
		{
			name:       "mention observation vocabulary",
			detailPart: "mentions=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				insertSourceIntegrityMention(t, db, fixture, "host", "legacy", "unresolved", nil)
			},
		},
		{
			name:       "mention target type",
			detailPart: "mentions=1",
			mutate: func(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture) {
				insertSourceIntegrityMention(t, db, fixture, "host", "manual_entry", "resolved", &fixture.identityID)
			},
		},
	}

	for _, test := range invalidFixtures {
		t.Run("preflight rejects "+test.name+" atomically", func(t *testing.T) {
			harness := pgtest.Start(t)
			migrationDB := harness.MigrationDatabaseThroughT(t, 35)
			fixture := seedSourceIntegrityFixture(t, migrationDB.SQL())
			test.mutate(t, migrationDB.SQL(), fixture)

			err := migrationDB.ApplyThrough(context.Background(), 36)
			var postgresError *pgconn.PgError
			if err == nil ||
				!strings.Contains(err.Error(), "entities_source_integrity_preflight_failed") ||
				!errors.As(err, &postgresError) ||
				!strings.Contains(postgresError.Detail, test.detailPart) {
				t.Fatalf("preflight error = %v, want aggregate %q", err, test.detailPart)
			}
			assertEntitySourceIntegrityObjects(t, migrationDB.SQL(), false)
			var appliedHead int64
			if err := migrationDB.SQL().QueryRowContext(context.Background(), `
SELECT max(version_id)
  FROM goose_db_version
 WHERE is_applied
`).Scan(&appliedHead); err != nil {
				t.Fatalf("inspect migration head after rejected preflight: %v", err)
			}
			if appliedHead != 35 {
				t.Fatalf("migration head after rejected preflight = %d, want 35", appliedHead)
			}
		})
	}
}

type sourceIntegrityFixture struct {
	actorID    uuid.UUID
	incidentID uuid.UUID
	hostID     uuid.UUID
	identityID uuid.UUID
}

func seedSourceIntegrityFixture(t testing.TB, db *sql.DB) sourceIntegrityFixture {
	t.Helper()
	actor := authflowtest.SeedLocalUserRecord(t, db, "entities-source-integrity@example.test", "Entities Source Integrity", "EntitiesSourceIntegrity1!", false, false, true)
	fixture := sourceIntegrityFixture{
		actorID:    actor.ID,
		incidentID: uuid.New(),
		hostID:     uuid.New(),
		identityID: uuid.New(),
	}
	seedSourceIntegrityIncident(t, db, fixture.incidentID, fixture.actorID, "IR-ESI-SOURCE")
	entitytest.SeedHostRecord(t, db, fixture.incidentID, fixture.actorID, fixture.hostID, "Source host", "source-host", "source.example.test", "AAD-SOURCE")
	entitytest.SeedIdentityRecord(t, db, fixture.incidentID, fixture.actorID, fixture.identityID, "Source identity", "source@example.test", "source@example.test", "SOURCE")
	insertSourceIntegrityAlias(t, db, fixture, "host", "Source Alias")
	insertSourceIntegrityPreservedIdentifier(t, db, fixture, "hostname", "SOURCE-HOST", "source-host")
	insertSourceIntegrityMention(t, db, fixture, "identity", "manual_entry", "resolved", &fixture.identityID)
	return fixture
}

func seedSourceIntegrityIncident(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, incidentKey string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, lower($2), $3, 'active', $4, $4)
`, incidentID, incidentKey, "Entities source integrity "+incidentKey, actorID); err != nil {
		t.Fatalf("seed source-integrity incident: %v", err)
	}
}

func seedSourceIntegrityEnvelope(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, $3, $4, $4)
`, recordID, incidentID, recordType, actorID); err != nil {
		t.Fatalf("seed source-integrity envelope: %v", err)
	}
}

func insertSourceIntegrityHost(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, state string, mergedInto *uuid.UUID, rowVersion int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, host_state, merged_into_record_id,
    row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT $1, $2, 'Source integrity host', $3, $4, $5,
       r.created_at, r.updated_at, r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
`, recordID, incidentID, state, mergedInto, rowVersion); err != nil {
		t.Fatalf("seed source-integrity host: %v", err)
	}
}

func insertSourceIntegrityAlias(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture, entityType string, value string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_aliases (
    incident_id, record_id, entity_type, raw_text, normalized_text,
    classification, created_by_user_id
) VALUES ($1, $2, $3, $4::text, $4::public.citext, 'suggestion_only', $5)
`, fixture.incidentID, fixture.hostID, entityType, value, fixture.actorID); err != nil {
		t.Fatalf("seed source-integrity alias: %v", err)
	}
}

func insertSourceIntegrityPreservedIdentifier(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture, identifierType string, rawValue string, normalizedValue string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_preserved_identifiers (
    incident_id, record_id, entity_type, identifier_type, raw_value,
    normalized_value, classification, created_by_user_id
) VALUES ($1, $2, 'host', $3, $4, $5, 'exact_match_reuse', $6)
`, fixture.incidentID, fixture.hostID, identifierType, rawValue, normalizedValue, fixture.actorID); err != nil {
		t.Fatalf("seed source-integrity preserved identifier: %v", err)
	}
}

func insertSourceIntegrityMention(t testing.TB, db *sql.DB, fixture sourceIntegrityFixture, entityType string, originKind string, status string, targetID *uuid.UUID) {
	t.Helper()
	var resolvedBy any
	var resolvedAt any
	var method any
	if status == "resolved" {
		resolvedBy = fixture.actorID
		resolvedAt = time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
		method = "explicit_resolve_route"
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind,
    origin_locator, raw_text, normalized_text, resolution_status,
    row_version, ordinal, created_by_user_id, resolved_record_id,
    resolved_by_user_id, resolved_at, resolution_method
) VALUES ($1, $2, 'timeline.identity_refs', $3, 'source-integrity-fixture',
          'Observed identity', 'observed identity', $4, 1, 1, $5, $6, $7, $8, $9)
`, fixture.hostID, entityType, originKind, status, fixture.actorID, targetID, resolvedBy, resolvedAt, method); err != nil {
		t.Fatalf("seed source-integrity mention: %v", err)
	}
}

func assertEntitySourceIntegrityObjects(t testing.TB, db *sql.DB, present bool) {
	t.Helper()
	ctx := context.Background()
	for _, object := range []struct {
		kind string
		name string
	}{
		{kind: "trigger", name: "entities_assert_host_source"},
		{kind: "trigger", name: "entities_assert_entity_mention"},
		{kind: "constraint", name: "entity_mentions_resolution_tuple_ck"},
		{kind: "routine", name: "entities_assert_entity_child_v1"},
	} {
		var exists bool
		var err error
		switch object.kind {
		case "trigger":
			err = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = $1 AND NOT tgisinternal)`, object.name).Scan(&exists)
		case "constraint":
			err = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = $1)`, object.name).Scan(&exists)
		case "routine":
			err = db.QueryRowContext(ctx, `SELECT to_regprocedure('public.' || $1 || '()') IS NOT NULL`, object.name).Scan(&exists)
		default:
			err = fmt.Errorf("unsupported object kind %q", object.kind)
		}
		if err != nil {
			t.Fatalf("inspect %s %s: %v", object.kind, object.name, err)
		}
		if exists != present {
			t.Fatalf("%s %s present = %t, want %t", object.kind, object.name, exists, present)
		}
	}
}

func assertEntitySourceIntegrityFunctionPrivileges(t testing.TB, db *sql.DB) {
	t.Helper()
	for _, signature := range []string{
		"public.entities_source_codepoints_admitted_v1(text,boolean)",
		"public.entities_trim_unicode_space_v1(text)",
		"public.entities_assert_entity_source_v1()",
		"public.entities_assert_entity_envelope_update_v1()",
		"public.entities_reject_entity_source_delete_v1()",
		"public.entities_assert_entity_child_v1()",
	} {
		var publicCanExecute bool
		if err := db.QueryRowContext(context.Background(), `SELECT has_function_privilege('public', $1, 'EXECUTE')`, signature).Scan(&publicCanExecute); err != nil {
			t.Fatalf("inspect function privilege %s: %v", signature, err)
		}
		if publicCanExecute {
			t.Fatalf("PUBLIC can execute owner-private function %s", signature)
		}
	}
}
