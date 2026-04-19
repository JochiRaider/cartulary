package timeline

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase3_ProjectionRowsKeepStableBindingAndDerivedFields_U_3_08(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	occurredAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	recordedAt := occurredAt.Add(2 * time.Minute)
	summary := "Summary"
	projected := projectRecord(sourceRecord{
		RecordID:        recordID,
		IncidentID:      incidentID,
		OccurredAt:      &occurredAt,
		Summary:         &summary,
		CaptureState:    captureStateEnriched,
		RowVersion:      4,
		RecordedAt:      recordedAt,
		EditedAt:        recordedAt,
		CreatedByUserID: incidentID,
		UpdatedByUserID: incidentID,
	}, nil)

	row := BuildRow(projected)
	if row["record_id"] != recordID.String() || row["row_version"] != int64(4) {
		t.Fatalf("expected stable record identity binding, got %#v", row)
	}
	groupValues := row["group_values"].(map[string]any)
	if groupValues["timeline.capture_state"] != captureStateEnriched {
		t.Fatalf("expected derived capture_state group value, got %#v", groupValues)
	}
	cells := row["cells"].(map[string]any)
	if cells["timeline.occurred_day"].(map[string]any)["value"] != "2026-04-10" {
		t.Fatalf("expected derived occurred_day, got %#v", cells["timeline.occurred_day"])
	}
}

func TestPhase3_TimelinePayloadBuildersExposeStableShapes_U_3_09(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	changeSetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	replacementID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	recordedAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	reason := "reviewed in workbook"
	projected := projectedRecord{
		RecordID:            recordID,
		IncidentID:          incidentID,
		RowVersion:          2,
		RecordedAt:          recordedAt,
		EditedAt:            recordedAt,
		SortTs:              recordedAt,
		CaptureState:        captureStateReviewed,
		ReplacementRecordID: &replacementID,
		RecordedDay:         recordedAt,
	}

	mutationPayload := BuildMutationPayload(projected, changeSetID)
	if mutationPayload["view_schema_id"] != TimelineViewSchemaID || mutationPayload["change_set_id"] != changeSetID.String() {
		t.Fatalf("unexpected mutation payload: %#v", mutationPayload)
	}
	row := mutationPayload["row"].(map[string]any)
	if row["record_id"] != recordID.String() || row["row_version"] != int64(2) {
		t.Fatalf("expected payload row identity/version, got %#v", row)
	}

	actionPayload := BuildActionPayload(projected, changeSetID, &reason)
	if actionPayload["capture_state"] != captureStateReviewed || actionPayload["reason"] != reason {
		t.Fatalf("unexpected action payload: %#v", actionPayload)
	}
	if actionPayload["replacement_record_id"] != replacementID.String() {
		t.Fatalf("expected replacement record id in action payload, got %#v", actionPayload)
	}
}

func TestPhase3_SupersedeGuardsAndActionHashes_U_3_10(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	otherIncidentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	otherRecordID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	anotherReplacementID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	if err := ValidateSupersedeReplacement(recordID, incidentID, &recordID, &incidentID); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("self replacement must be rejected, got %v", err)
	}
	if err := ValidateSupersedeReplacement(recordID, incidentID, &otherRecordID, &otherIncidentID); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("cross-incident replacement must be rejected, got %v", err)
	}
	if err := ValidateSupersedeReplacement(recordID, incidentID, &otherRecordID, &incidentID); err != nil {
		t.Fatalf("legal replacement should be allowed, got %v", err)
	}

	reason := "superseded"
	left := TimelineActionRequestHash(4, "txn-u-3-10", &reason, &otherRecordID)
	right := TimelineActionRequestHash(4, "txn-u-3-10", &reason, &otherRecordID)
	if !hashesEqual(left, right) {
		t.Fatal("expected identical supersede request hashes to match")
	}

	if hashesEqual(left, TimelineActionRequestHash(4, "txn-u-3-10", &reason, &anotherReplacementID)) {
		t.Fatal("expected replacement id changes to alter the normalized request hash")
	}

	differentReason := "superseded differently"
	if hashesEqual(left, TimelineActionRequestHash(4, "txn-u-3-10", &differentReason, &otherRecordID)) {
		t.Fatal("expected reason changes to alter the normalized request hash")
	}
}
