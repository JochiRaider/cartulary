package entities_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEntitySourceIntegrityBaseline_Integration(t *testing.T) {

	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	fixture := seedSourceIntegrityFixture(t, db)
	assertEntitySourceIntegrityObjects(t, db, true)
	assertEntitySourceIntegrityFunctionPrivileges(t, db)
	for _, query := range []string{
		`UPDATE hosts SET row_version = row_version + 1 WHERE record_id = $1`,
		`UPDATE entity_aliases SET entity_type = 'identity' WHERE record_id = $1`,
		`UPDATE entity_aliases SET normalized_text = ' padded ' WHERE record_id = $1`,
		`UPDATE entity_preserved_identifiers SET normalized_value = 'wrong' WHERE record_id = $1`,
		`UPDATE entity_mentions SET origin_kind = 'legacy' WHERE source_record_id = $1`,
		`UPDATE entity_mentions SET entity_type = 'host' WHERE source_record_id = $1`,
	} {
		if _, err := db.ExecContext(t.Context(), query, fixture.hostID); err == nil {
			t.Fatalf("invalid source mutation accepted: %s", query)
		}
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
	// These immutable-schema fixtures need an attribution identity, not a
	// current credential record or an authentication-capable account.
	actorID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO users (id, email, display_name, password_hash, is_active)
VALUES ($1, 'entities-source-integrity@example.test', 'Entities Source Integrity',
        'historical-migration-fixture-no-login', false)
`, actorID); err != nil {
		t.Fatalf("seed historical source-integrity actor: %v", err)
	}
	fixture := sourceIntegrityFixture{
		actorID:    actorID,
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
