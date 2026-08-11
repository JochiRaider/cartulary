package entities_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEntityAliasHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "entity-alias-head-contract", postgres.PurposeRecovery)
	ctx := context.Background()

	var udtName string
	if err := db.QueryRowContext(ctx, `
SELECT udt_name
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'entity_aliases'
   AND column_name = 'normalized_text'
`).Scan(&udtName); err != nil {
		t.Fatalf("inspect alias type: %v", err)
	}
	if udtName != "citext" {
		t.Fatalf("entity_aliases.normalized_text type = %q, want citext", udtName)
	}

	actorID, incidentID, recordID := seedAliasHeadHost(t, db)
	insertAliasHeadRow(t, db, uuid.New(), incidentID, recordID, actorID, "Case Alias")
	if _, err := db.ExecContext(ctx, `
INSERT INTO public.entity_aliases (
    entity_alias_id, incident_id, record_id, entity_type, raw_text,
    normalized_text, classification, created_by_user_id, created_at, deleted_at
) VALUES ($1, $2, $3, 'host', 'case alias', 'case alias', 'suggestion_only', $4, $5, NULL)
`, uuid.New(), incidentID, recordID, actorID, time.Now().UTC()); err == nil || !strings.Contains(err.Error(), "entity_aliases_record_unique_idx") {
		t.Fatalf("case-equivalent active alias error = %v, want entity_aliases_record_unique_idx", err)
	}

	for name, fixture := range map[string]struct {
		text       string
		constraint string
	}{
		"empty":   {text: "", constraint: "entity_aliases_alias_text_nonempty_ck"},
		"control": {text: "bad\u0085alias", constraint: "entity_aliases_alias_text_controls_ck"},
		"long":    {text: strings.Repeat("x", 257), constraint: "entity_aliases_alias_text_length_ck"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, `
INSERT INTO public.entity_aliases (
    entity_alias_id, incident_id, record_id, entity_type, raw_text,
    normalized_text, classification, created_by_user_id, created_at, deleted_at
) VALUES ($1, $2, $3, 'host', $4::text, $4::public.citext, 'suggestion_only', $5, $6, NULL)
`, uuid.New(), incidentID, recordID, fixture.text, actorID, time.Now().UTC())
			if err == nil || !strings.Contains(err.Error(), fixture.constraint) {
				t.Fatalf("invalid alias error = %v, want %s", err, fixture.constraint)
			}
		})
	}
}

func seedAliasHeadHost(t *testing.T, db *sql.DB) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	actor := authflowtest.SeedLocalUserRecord(t, db, "alias-head@example.test", "Alias Head", "AliasHeadPass1!", false, false, true)
	incidentID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO public.incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, 'IR-ALIAS-HEAD', 'ir-alias-head', 'Alias head contract', 'active', $2, $2)
`, incidentID, actor.ID); err != nil {
		t.Fatalf("seed alias incident: %v", err)
	}
	recordID := uuid.New()
	entitytest.SeedHostRecord(t, db, incidentID, actor.ID, recordID, "Alias Host", "ALIAS-HOST", "", "")
	return actor.ID, incidentID, recordID
}

func insertAliasHeadRow(t *testing.T, db *sql.DB, aliasID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, text string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO public.entity_aliases (
    entity_alias_id, incident_id, record_id, entity_type, raw_text,
    normalized_text, classification, created_by_user_id, created_at, deleted_at
) VALUES ($1, $2, $3, 'host', $4::text, $4::public.citext, 'suggestion_only', $5, $6, NULL)
`, aliasID, incidentID, recordID, text, actorID, time.Now().UTC()); err != nil {
		t.Fatalf("insert alias: %v", err)
	}
}
