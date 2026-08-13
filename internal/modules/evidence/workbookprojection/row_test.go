package workbookprojection

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestViewRowMapsRichProjectionInput(t *testing.T) {
	recordID := uuid.New()
	collectorPartyID := uuid.New()
	sourcePartyID := uuid.New()
	title := "Forensic image"
	storageRef := "case/evidence/image"
	blobHash := "sha256:abcdef"
	collectorPartyText := "IR lead"
	sourcePartyText := "endpoint owner"
	requestedAt := time.Date(2026, 4, 30, 12, 1, 2, 3, time.FixedZone("offset", -4*60*60))
	receivedAt := time.Date(2026, 4, 30, 13, 1, 2, 3, time.UTC)
	editedAt := time.Date(2026, 4, 30, 14, 1, 2, 3, time.UTC)

	row := ViewRow(ProjectionInput{
		RecordID:           recordID,
		RowVersion:         7,
		Title:              &title,
		LifecycleState:     "available",
		RequestedAt:        &requestedAt,
		ReceivedAt:         &receivedAt,
		StorageRef:         &storageRef,
		BlobHash:           &blobHash,
		CollectorPartyText: &collectorPartyText,
		CollectorPartyID:   &collectorPartyID,
		SourcePartyText:    &sourcePartyText,
		SourcePartyID:      &sourcePartyID,
		UploadState:        "available",
		LinkedRecordCount:  3,
		EditedAt:           editedAt,
	})

	if row["record_id"] != recordID.String() {
		t.Fatalf("record_id got %#v want %q", row["record_id"], recordID.String())
	}
	if row["row_version"] != int64(7) {
		t.Fatalf("row_version got %#v want 7", row["row_version"])
	}
	cells := requireCells(t, row)
	if len(cells) != 13 {
		t.Fatalf("cells length got %d want 13: %#v", len(cells), cells)
	}
	assertCellValue(t, cells, "evidence.title", title)
	assertCellValue(t, cells, "evidence.lifecycle_state", "available")
	assertCellValue(t, cells, "evidence.requested_at", requestedAt.UTC().Format(time.RFC3339Nano))
	assertCellValue(t, cells, "evidence.received_at", receivedAt.UTC().Format(time.RFC3339Nano))
	assertCellValue(t, cells, "evidence.storage_ref", storageRef)
	assertCellValue(t, cells, "evidence.blob_hash", blobHash)
	assertCellValue(t, cells, "evidence.collector_party_text", collectorPartyText)
	assertCellValue(t, cells, "evidence.collector_party_id", collectorPartyID.String())
	assertCellValue(t, cells, "evidence.source_party_text", sourcePartyText)
	assertCellValue(t, cells, "evidence.source_party_id", sourcePartyID.String())
	assertCellValue(t, cells, "evidence.upload_state", "available")
	assertCellValue(t, cells, "evidence.linked_record_count", int32(3))
	assertCellValue(t, cells, "evidence.edited_at", editedAt.UTC().Format(time.RFC3339Nano))
}

func TestViewRowPreservesNullProjectionCells(t *testing.T) {
	row := ViewRow(ProjectionInput{
		RecordID:       uuid.New(),
		RowVersion:     1,
		LifecycleState: "requested",
		UploadState:    "none",
		EditedAt:       time.Date(2026, 5, 1, 1, 2, 3, 4, time.UTC),
	})
	cells := requireCells(t, row)
	for _, key := range []string{
		"evidence.title",
		"evidence.requested_at",
		"evidence.received_at",
		"evidence.storage_ref",
		"evidence.blob_hash",
		"evidence.collector_party_text",
		"evidence.collector_party_id",
		"evidence.source_party_text",
		"evidence.source_party_id",
	} {
		assertCellValue(t, cells, key, nil)
	}
}

func TestViewRowPreservesLifecycleAndBlobStates(t *testing.T) {
	states := []struct {
		lifecycle string
		upload    string
	}{
		{lifecycle: "requested", upload: "pending"},
		{lifecycle: "received", upload: "pending"},
		{lifecycle: "available", upload: "available"},
		{lifecycle: "released", upload: "available"},
		{lifecycle: "available", upload: "failed"},
		{lifecycle: "quarantined", upload: "quarantined"},
	}
	for _, state := range states {
		t.Run(state.lifecycle+"/"+state.upload, func(t *testing.T) {
			row := ViewRow(ProjectionInput{
				RecordID:       uuid.New(),
				RowVersion:     9,
				LifecycleState: state.lifecycle,
				UploadState:    state.upload,
				EditedAt:       time.Date(2026, 5, 2, 3, 4, 5, 6, time.UTC),
			})
			cells := requireCells(t, row)
			assertCellValue(t, cells, "evidence.lifecycle_state", state.lifecycle)
			assertCellValue(t, cells, "evidence.upload_state", state.upload)
		})
	}
}

func requireCells(t testing.TB, row map[string]any) map[string]any {
	t.Helper()
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		t.Fatalf("cells got %T", row["cells"])
	}
	return cells
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
