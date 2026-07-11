package postgres_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSupportPhase4_EntityAliasMigration31EmptyUpgrade(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseT(t, "entity-alias-31-empty", "up-to", "31")

	var udtName string
	if err := db.QueryRowContext(context.Background(), `
SELECT udt_name
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'entity_aliases'
   AND column_name = 'normalized_text'
`).Scan(&udtName); err != nil {
		t.Fatalf("inspect migrated alias type: %v", err)
	}
	if udtName != "citext" {
		t.Fatalf("expected normalized_text citext after migration 31, got %q", udtName)
	}
}

func TestSupportPhase4_EntityAliasMigration31UpgradeTombstonesCaseEquivalentDuplicates(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseT(t, "entity-alias-31-dedupe", "up-to", "30")
	actorID, incidentID, recordID := seedAliasMigrationHost(t, db, "dedupe")
	oldestID := uuid.MustParse("31000000-0000-4000-8000-000000000001")
	newestID := uuid.MustParse("31000000-0000-4000-8000-000000000002")
	insertLegacyAlias(t, db, oldestID, incidentID, recordID, actorID, "Case Alias", time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC))
	insertLegacyAlias(t, db, newestID, incidentID, recordID, actorID, "case alias", time.Date(2026, 7, 11, 12, 1, 0, 0, time.UTC))

	if _, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "31"); err != nil {
		t.Fatalf("migrate alias duplicate fixture to 31: %v", err)
	}
	var oldestDeletedAt, newestDeletedAt sql.NullTime
	if err := db.QueryRowContext(context.Background(), `SELECT deleted_at FROM entity_aliases WHERE entity_alias_id = $1`, oldestID).Scan(&oldestDeletedAt); err != nil {
		t.Fatalf("load oldest migrated alias: %v", err)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT deleted_at FROM entity_aliases WHERE entity_alias_id = $1`, newestID).Scan(&newestDeletedAt); err != nil {
		t.Fatalf("load newest migrated alias: %v", err)
	}
	if oldestDeletedAt.Valid || !newestDeletedAt.Valid {
		t.Fatalf("expected oldest alias active and later case-equivalent alias tombstoned: oldest=%#v newest=%#v", oldestDeletedAt, newestDeletedAt)
	}
}

func TestSupportPhase4_EntityAliasMigration31RejectsInvalidLegacyRowsWithCountAndSample(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseT(t, "entity-alias-31-invalid", "up-to", "30")
	actorID, incidentID, recordID := seedAliasMigrationHost(t, db, "invalid")
	invalid := []struct {
		id   uuid.UUID
		text string
	}{
		{id: uuid.MustParse("31000000-0000-4000-8000-000000000011"), text: ""},
		{id: uuid.MustParse("31000000-0000-4000-8000-000000000012"), text: "bad\u0085alias"},
		{id: uuid.MustParse("31000000-0000-4000-8000-000000000013"), text: strings.Repeat("x", 257)},
	}
	for index, fixture := range invalid {
		insertLegacyAlias(t, db, fixture.id, incidentID, recordID, actorID, fixture.text, time.Date(2026, 7, 11, 13, index, 0, 0, time.UTC))
	}

	_, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "31")
	if err == nil {
		t.Fatal("expected migration 31 to reject invalid legacy aliases")
	}
	message := err.Error()
	if !strings.Contains(message, "invalid_count=3") {
		t.Fatalf("expected explicit invalid alias count, got %v", err)
	}
	for _, fixture := range invalid {
		if !strings.Contains(message, fixture.id.String()) {
			t.Fatalf("expected invalid alias sample to contain %s, got %v", fixture.id, err)
		}
	}
}

func seedAliasMigrationHost(t *testing.T, db *sql.DB, suffix string) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	actor := scenariotest.SeedLocalUserFlags(t, db, "alias-migration-"+suffix+"@example.test", "Alias Migration", "AliasMigrationPass1!", false, false, true)
	incidentID := uuid.New()
	incidentKey := "IR-ALIAS-" + strings.ToUpper(suffix)
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, $3, $4, 'active', $5, $5)
`, incidentID, incidentKey, strings.ToLower(incidentKey), "Alias migration "+suffix, actor.ID); err != nil {
		t.Fatalf("seed alias migration incident: %v", err)
	}
	recordID := uuid.New()
	scenariotest.SeedHostRecord(t, db, incidentID, actor.ID, recordID, "Alias Host", "ALIAS-HOST", "", "")
	return actor.ID, incidentID, recordID
}

func insertLegacyAlias(t *testing.T, db *sql.DB, aliasID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, text string, createdAt time.Time) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_aliases (
    entity_alias_id, incident_id, record_id, entity_type, raw_text,
    normalized_text, classification, created_by_user_id, created_at, deleted_at
) VALUES ($1, $2, $3, 'host', $4, $4, 'suggestion_only', $5, $6, NULL)
`, aliasID, incidentID, recordID, text, actorID, createdAt.UTC()); err != nil {
		t.Fatalf("insert legacy alias %s: %v", aliasID, err)
	}
}
