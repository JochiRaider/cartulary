package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"

	. "github.com/JochiRaider/cartulary/internal/modules/timeline"
)

// U-4-01 / REQ-02-028..REQ-02-036 / AC-019, AC-020, AC-022.
func TestPhase4_BindingMode_U_4_01(t *testing.T) {
	harness := phase4test.StartStore(t, "phase4-u-4-01")
	timelineStore := NewStore(harness.Pool)
	entityStore := entities.NewStore(harness.Pool)
	actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u401@example.test", "U401", "U401Phase4Pass1!", false, false, true)
	incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-01-incident", "IR-U401", "Phase 4 U-4-01")

	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4TimelineViewSchemaID, golden.Phase4FieldTimelineHostRefs, "mention_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4TimelineViewSchemaID, golden.Phase4FieldTimelineIdentityRefs, "mention_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4HostsViewSchemaID, "host.display_name", "entity_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4HostsViewSchemaID, "host.hostname", "entity_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4HostsViewSchemaID, "host.aliases", "entity_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4IdentitiesViewSchemaID, "identity.display_name", "entity_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4IdentitiesViewSchemaID, "identity.email", "entity_origin")
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4IdentitiesViewSchemaID, "identity.aliases", "entity_origin")

	normalizedHostToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	timelineResult, err := timelineStore.CreateRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase4-u-4-01-timeline",
		Summary:     stringPtr("Mention origin row"),
		HostRefs: &CollectionActionPayload{
			Actions: []CollectionAction{
				{
					Op:             "add_token",
					RawText:        "WS-023",
					NormalizedText: normalizedHostToken,
				},
			},
		},
	}, []byte("txn-phase4-u-4-01-timeline"), "req-phase4-u-4-01-timeline", time.Now().UTC())
	if err != nil {
		t.Fatalf("create mention-origin row: %v", err)
	}
	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, timelineResult.RecordID); got != 1 {
		t.Fatalf("expected mention_origin write to create one entity mention, got %d", got)
	}
	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 0 {
		t.Fatalf("mention_origin write must not synthesize hosts, got %d", got)
	}
	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 0 {
		t.Fatalf("mention_origin write must not synthesize identities, got %d", got)
	}

	entityResult, err := entityStore.CreateHostRow(context.Background(), actor, incident.ID, entities.CreateRequest{
		ClientTxnID: "txn-phase4-u-4-01-host",
		Values: map[string]string{
			"host.display_name": "WS-023",
			"host.hostname":     "WS-023",
		},
	}, []byte("txn-phase4-u-4-01-host"), "req-phase4-u-4-01-host", time.Now().UTC())
	if err != nil {
		t.Fatalf("create entity-origin host row: %v", err)
	}
	if entityResult.RecordID == timelineResult.RecordID {
		t.Fatalf("unexpected shared record id between timeline row and host row: %#v %#v", timelineResult, entityResult)
	}
	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
		t.Fatalf("expected entity_origin write to create one host, got %d", got)
	}
	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, entityResult.RecordID); got != 0 {
		t.Fatalf("entity_origin write must not synthesize mentions, got %d", got)
	}
}

// U-4-02 / REQ-02-031..REQ-02-032, REQ-02-058 / AC-019, AC-021.
func TestPhase4_DuplicateMentionProvenance_U_4_02(t *testing.T) {
	harness := phase4test.StartStore(t, "phase4-u-4-02")
	store := NewStore(harness.Pool)
	actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u402@example.test", "U402", "U402Phase4Pass1!", false, false, true)
	incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-02-incident", "IR-U402", "Phase 4 U-4-02")

	normalizedToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	first, err := store.CreateRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase4-u-4-02-first",
		Summary:     stringPtr("Duplicate provenance one"),
		HostRefs: &CollectionActionPayload{
			Actions: []CollectionAction{{Op: "add_token", RawText: "WS-023", NormalizedText: normalizedToken}},
		},
	}, []byte("txn-phase4-u-4-02-first"), "req-phase4-u-4-02-first", time.Now().UTC())
	if err != nil {
		t.Fatalf("create first row: %v", err)
	}
	second, err := store.CreateRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase4-u-4-02-second",
		Summary:     stringPtr("Duplicate provenance two"),
		HostRefs: &CollectionActionPayload{
			Actions: []CollectionAction{{Op: "add_token", RawText: "WS-023", NormalizedText: normalizedToken}},
		},
	}, []byte("txn-phase4-u-4-02-second"), "req-phase4-u-4-02-second", time.Now().UTC())
	if err != nil {
		t.Fatalf("create second row: %v", err)
	}

	rows, err := harness.DB.QueryContext(context.Background(), `
SELECT entity_mention_id::text, source_record_id::text, raw_text, origin_locator
  FROM entity_mentions
 WHERE source_record_id IN ($1, $2)
 ORDER BY created_at ASC, entity_mention_id ASC
`, first.RecordID, second.RecordID)
	if err != nil {
		t.Fatalf("query entity mentions: %v", err)
	}
	defer rows.Close()

	type mentionRow struct {
		id            string
		sourceRecord  string
		rawText       string
		originLocator string
	}
	var mentions []mentionRow
	for rows.Next() {
		var row mentionRow
		if err := rows.Scan(&row.id, &row.sourceRecord, &row.rawText, &row.originLocator); err != nil {
			t.Fatalf("scan entity mention: %v", err)
		}
		mentions = append(mentions, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate entity mentions: %v", err)
	}
	if len(mentions) != 2 {
		t.Fatalf("expected two durable mentions for repeated identical tokens, got %#v", mentions)
	}
	if mentions[0].rawText != "WS-023" || mentions[1].rawText != "WS-023" {
		t.Fatalf("expected repeated raw text preservation, got %#v", mentions)
	}
	if mentions[0].id == mentions[1].id || mentions[0].originLocator == mentions[1].originLocator || mentions[0].sourceRecord == mentions[1].sourceRecord {
		t.Fatalf("expected repeated identical tokens to keep distinct provenance, got %#v", mentions)
	}
	if mentions[0].sourceRecord != first.RecordID.String() || mentions[1].sourceRecord != second.RecordID.String() {
		t.Fatalf("expected mention provenance to stay attached to each source row, got %#v", mentions)
	}
}

func stringPtr(value string) *string {
	return &value
}
