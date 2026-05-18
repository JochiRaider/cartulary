package workbook_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	phase4storetest "github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func TestPhase9_NotesAreArtifactBackedRows_U_9_03(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-u-9-03-notes")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u903@example.test", "U903 Notes", "U903NotesPass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-03-incident", "IR-U903", "Phase 9 U-9-03")
	sourceRecordID := uuid.New()
	phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceRecordID)

	created, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.NotesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-03-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Phase 9 artifact note"),
			"note.body":  textChange("Artifact-backed note body"),
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"note.tags": {
				Actions: []workbook.CollectionAction{{Op: "add_tag", RawText: "phase9-sprint3", NormalizedText: "phase9-sprint3"}},
			},
		},
	}, []byte("txn-phase9-u-9-03-note"), "req-phase9-u-9-03-note", time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create note through workbook store: %v", err)
	}

	requireScalarCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN artifacts a ON a.incident_id = r.incident_id AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = 'note'
`, created.RecordID, 1)
	requireScalarCount(t, harness, `SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'notes'`, 0)
	requireScalarCount(t, harness, `SELECT count(*) FROM record_tags WHERE incident_id = $1 AND record_id = $2 AND normalized_tag_name = 'phase9-sprint3' AND deleted_at IS NULL`, incident.ID, created.RecordID, 1)
	requireScalarCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND row_version = 1`, created.RecordID, 1)

	linked, err := store.CreateLinkedNote(context.Background(), actor, sourceRecordID, workbook.LinkedNoteCreateRequest{
		ClientTxnID: "txn-phase9-u-9-03-linked-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Phase 9 linked note"),
			"note.body":  textChange("Linked through references_artifact"),
		},
	}, workbook.LinkedNoteCreateRequestHash(sourceRecordID, workbook.LinkedNoteCreateRequest{
		ClientTxnID: "txn-phase9-u-9-03-linked-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Phase 9 linked note"),
			"note.body":  textChange("Linked through references_artifact"),
		},
	}), "req-phase9-u-9-03-linked-note", time.Date(2026, 5, 17, 15, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create linked note: %v", err)
	}
	requireScalarCount(t, harness, `
SELECT count(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'references_artifact'
   AND field_key IS NULL
   AND deleted_at IS NULL
`, incident.ID, sourceRecordID, linked.RecordID, 1)

	rows, err := store.QueryRows(context.Background(), incident.ID, workbook.NotesViewSchemaID, mustQueryMeta(t, workbook.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes rows: %v", err)
	}
	linkedRow := requireQueriedRow(t, rows, linked.RecordID)
	if got := linkedRow["cells"].(map[string]any)["note.linked_record_count"].(map[string]any)["value"]; got != float64(1) && got != int32(1) && got != int64(1) && got != 1 {
		t.Fatalf("expected linked note count to include incoming contextual link, got %#v", got)
	}
}

func TestPhase9_NotesAndIndicatorsQueryThroughWorkbookProjections_I_9_02(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-i-9-02-notes-indicators")
	workbookStore := workbook.NewStore(harness.DB)
	entityStore := entities.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "i902@example.test", "I902 Projection", "I902ProjectionPass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-i-9-02-incident", "IR-I902", "Phase 9 I-9-02")

	note, err := workbookStore.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.NotesViewSchemaID,
		ClientTxnID:  "txn-phase9-i-9-02-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Projection-backed note"),
			"note.body":  textChange("query token phase9-note-projection"),
		},
	}, []byte("txn-phase9-i-9-02-note"), "req-phase9-i-9-02-note", time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection note: %v", err)
	}
	noteRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.NotesViewSchemaID, mustQueryMeta(t, workbook.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes projection: %v", err)
	}
	requireQueriedRow(t, noteRows, note.RecordID)

	indicator, err := entityStore.CreateIndicatorRow(context.Background(), actor, incident.ID, entities.CreateRequest{
		ClientTxnID: "txn-phase9-i-9-02-indicator",
		Values: map[string]string{
			"indicator.indicator_type": "ipv4_addr",
			"indicator.value_kind":     "atomic",
			"indicator.display_value":  "203.0.113.45",
		},
	}, []byte("txn-phase9-i-9-02-indicator"), "req-phase9-i-9-02-indicator", time.Date(2026, 5, 17, 16, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection indicator: %v", err)
	}
	indicatorRows, err := workbookStore.QueryRows(context.Background(), incident.ID, entities.IndicatorsViewSchemaID, mustQueryMeta(t, entities.IndicatorsViewSchemaID))
	if err != nil {
		t.Fatalf("query indicators through workbook store: %v", err)
	}
	requireQueriedRow(t, indicatorRows, indicator.RecordID)
}

func textChange(value string) workbook.ValueChange {
	return workbook.ValueChange{Kind: "text", Text: &value}
}

func mustQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("missing view schema %s", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func requireQueriedRow(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("missing queried row %s in %#v", recordID, rows)
	return nil
}

func requireScalarCount(t testing.TB, harness *phase4storetest.StoreHarness, query string, args ...any) {
	t.Helper()
	want := args[len(args)-1].(int)
	args = args[:len(args)-1]
	var got int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query scalar count: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}
