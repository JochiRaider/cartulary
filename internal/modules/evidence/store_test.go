package evidence

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvidenceRowFromDataMapsNamedFields(t *testing.T) {
	recordID := uuid.New()
	collectorPartyID := uuid.New()
	sourcePartyID := uuid.New()
	requestedAt := time.Date(2026, 4, 30, 12, 1, 2, 3, time.FixedZone("offset", -4*60*60))
	receivedAt := time.Date(2026, 4, 30, 13, 1, 2, 3, time.UTC)
	editedAt := time.Date(2026, 4, 30, 14, 1, 2, 3, time.UTC)

	row, err := evidenceRowFromData(evidenceRowData{
		RecordID:           recordID,
		RowVersion:         int64(7),
		Title:              "Forensic image",
		LifecycleState:     "available",
		RequestedAt:        requestedAt,
		ReceivedAt:         receivedAt,
		StorageRef:         "case/evidence/image",
		BlobHash:           "sha256:abcdef",
		CollectorPartyText: "IR lead",
		CollectorPartyID:   collectorPartyID,
		SourcePartyText:    "endpoint owner",
		SourcePartyID:      sourcePartyID,
		UploadState:        "available",
		LinkedRecordCount:  int64(3),
		EditedAt:           editedAt,
	})
	if err != nil {
		t.Fatalf("evidenceRowFromData returned error: %v", err)
	}
	if row["record_id"] != recordID.String() {
		t.Fatalf("record_id got %#v want %q", row["record_id"], recordID.String())
	}
	if row["row_version"] != int64(7) {
		t.Fatalf("row_version got %#v want 7", row["row_version"])
	}

	cells, ok := row["cells"].(map[string]any)
	if !ok {
		t.Fatalf("cells got %T", row["cells"])
	}
	if len(cells) != 13 {
		t.Fatalf("cells length got %d want 13: %#v", len(cells), cells)
	}

	assertCellValue(t, cells, "evidence.title", "Forensic image")
	assertCellValue(t, cells, "evidence.lifecycle_state", "available")
	assertCellValue(t, cells, "evidence.requested_at", requestedAt.UTC().Format(time.RFC3339Nano))
	assertCellValue(t, cells, "evidence.received_at", receivedAt.UTC().Format(time.RFC3339Nano))
	assertCellValue(t, cells, "evidence.storage_ref", "case/evidence/image")
	assertCellValue(t, cells, "evidence.blob_hash", "sha256:abcdef")
	assertCellValue(t, cells, "evidence.collector_party_text", "IR lead")
	assertCellValue(t, cells, "evidence.collector_party_id", collectorPartyID.String())
	assertCellValue(t, cells, "evidence.source_party_text", "endpoint owner")
	assertCellValue(t, cells, "evidence.source_party_id", sourcePartyID.String())
	assertCellValue(t, cells, "evidence.upload_state", "available")
	assertCellValue(t, cells, "evidence.linked_record_count", int64(3))
	assertCellValue(t, cells, "evidence.edited_at", editedAt.UTC().Format(time.RFC3339Nano))
}

func assertCellValue(t testing.TB, cells map[string]any, key string, want any) {
	t.Helper()
	cell, ok := cells[key].(map[string]any)
	if !ok {
		t.Fatalf("%s cell got %T", key, cells[key])
	}
	if got := cell["value"]; got != want {
		t.Fatalf("%s value got %#v want %#v", key, got, want)
	}
}
