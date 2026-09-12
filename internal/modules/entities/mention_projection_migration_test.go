package entities_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
)

func TestTimelineMentionProjectionIdentityMigration_Integration(t *testing.T) {
	for _, corruption := range []string{"", "selector", "version", "identity", "source"} {
		t.Run("upgrade binding "+corruption, func(t *testing.T) {
			harness := pgtest.Start(t)
			migrationDB := harness.MigrationDatabaseThroughT(t, 41)
			db := migrationDB.SQL()
			fixture := seedSourceIntegrityFixture(t, db)
			sourceID := uuid.New()
			timelinetest.SeedTimelineRecord(t, db, fixture.incidentID, fixture.actorID, sourceID)
			facts := workbookprojection.CollectionFacts{Mentions: []workbookprojection.MentionFact{
				{MentionID: uuid.New(), EntityType: "host", SourceFieldKey: "timeline.host_refs", RawText: "  unknown.host  ", ResolutionStatus: "unresolved", RowVersion: 1},
				{MentionID: uuid.New(), EntityType: "host", SourceFieldKey: "timeline.host_refs", RawText: "second host", ResolutionStatus: "unresolved", RowVersion: 3},
				{MentionID: uuid.New(), EntityType: "identity", SourceFieldKey: "timeline.identity_refs", RawText: "unknown@identity", ResolutionStatus: "unresolved", RowVersion: 2},
			}}
			ctx := context.Background()
			for index, mention := range facts.Mentions {
				if _, err := db.ExecContext(ctx, `
INSERT INTO entity_mentions (entity_mention_id, source_record_id, entity_type, source_field_key,
    origin_kind, origin_locator, raw_text, normalized_text, resolution_status, row_version, ordinal, created_by_user_id)
VALUES ($1, $2, $3, $4, 'manual_entry', 'projection-upgrade', $5, lower(trim($5)), 'unresolved', $6, $7, $8)
`, mention.MentionID, sourceID, mention.EntityType, mention.SourceFieldKey, mention.RawText, mention.RowVersion, index+1, fixture.actorID); err != nil {
					t.Fatal(err)
				}
			}
			projected := workbookprojection.DerivedRecord{RecordID: sourceID}
			workbookprojection.ApplyCollectionFacts(&projected, facts)
			hosts, err := json.Marshal(projected.HostRefs)
			if err != nil {
				t.Fatal(err)
			}
			identities, err := json.Marshal(projected.IdentityRefs)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `
INSERT INTO timeline_grid_projection (record_id, incident_id, row_version, recorded_at, edited_at, capture_state, host_refs, identity_refs)
SELECT $1, $2, 1, now(), now(), 'reviewed',
    (SELECT jsonb_agg(item - 'entity_mention_id' ORDER BY ordinal) FROM jsonb_array_elements($3::jsonb) WITH ORDINALITY AS items(item, ordinal)),
    (SELECT jsonb_agg(item - 'entity_mention_id' ORDER BY ordinal) FROM jsonb_array_elements($4::jsonb) WITH ORDINALITY AS items(item, ordinal))
`, sourceID, fixture.incidentID, string(hosts), string(identities)); err != nil {
				t.Fatal(err)
			}
			if corruption != "" {
				field, value := "{0,item_ref}", `"unrelated opaque selector"`
				switch corruption {
				case "version":
					field, value = "{0,mention_row_version}", "99"
				case "identity":
					field, value = "{0,entity_mention_id}", `"`+uuid.NewString()+`"`
				case "source":
					// A real mention belonging to a different source must not bind.
					var id string
					if err := db.QueryRowContext(ctx, `SELECT entity_mention_id::text FROM entity_mentions WHERE source_record_id = $1`, fixture.hostID).Scan(&id); err != nil {
						t.Fatal(err)
					}
					value = `"entity_mention:` + id + `"`
				}
				if _, err := db.ExecContext(ctx, `UPDATE timeline_grid_projection SET host_refs = jsonb_set(host_refs, $2::text[], $3::jsonb) WHERE record_id = $1`, sourceID, field, value); err != nil {
					t.Fatal(err)
				}
			}
			before := mentionProjectionProtectedState(t, db)
			err = migrationDB.ApplyThrough(ctx, 42)
			if corruption != "" {
				if err == nil || !strings.Contains(err.Error(), "timeline_mention_projection_binding_invalid") {
					t.Fatalf("invalid binding accepted: %v", err)
				}
				var head int
				if err := db.QueryRowContext(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&head); err != nil {
					t.Fatal(err)
				}
				if head != 41 {
					t.Fatalf("rejected upgrade advanced head: %d", head)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				var equivalent bool
				if err := db.QueryRowContext(ctx, `SELECT host_refs = $2::jsonb AND identity_refs = $3::jsonb FROM timeline_grid_projection WHERE record_id = $1`, sourceID, string(hosts), string(identities)).Scan(&equivalent); err != nil {
					t.Fatal(err)
				}
				if !equivalent {
					t.Fatal("backfill differs from projection owner or changed array order/metadata")
				}
				if err := migrationDB.RollbackThrough(ctx, 41); err != nil {
					t.Fatal(err)
				}
				if err := migrationDB.ApplyThrough(ctx, 42); err != nil {
					t.Fatalf("additive backfill is not repeatable: %v", err)
				}
			}
			if after := mentionProjectionProtectedState(t, db); before != after {
				t.Fatal("projection upgrade changed source, links, receipts or history")
			}
		})
	}
}

func mentionProjectionProtectedState(t testing.TB, db *sql.DB) string {
	t.Helper()
	var state string
	if err := db.QueryRowContext(context.Background(), `
SELECT jsonb_build_object(
    'timeline', (SELECT jsonb_agg(to_jsonb(t) ORDER BY record_id) FROM timeline_events t),
    'mentions', (SELECT jsonb_agg(to_jsonb(m) ORDER BY entity_mention_id) FROM entity_mentions m),
    'links', (SELECT jsonb_agg(to_jsonb(l) ORDER BY record_link_id) FROM record_links l),
    'changes', (SELECT jsonb_agg(to_jsonb(c) ORDER BY change_set_id) FROM change_sets c)
)::text`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return entityUnicodeCompatibilityState(t, db) + state
}
