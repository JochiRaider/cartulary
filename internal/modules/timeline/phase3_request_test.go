package timeline

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase3_TimelineCreateValidation_U_3_01(t *testing.T) {
	request, apiErr := DecodeTimelineCreateRequest(bytes.NewBufferString(`{
		"client_txn_id": "txn-u-3-01",
		"timeline.summary": "  First capture  ",
		"timeline.details": " detail line ",
		"timeline.occurred_at": "2026-04-10T10:00:00Z"
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid create request, got %#v", apiErr)
	}
	if request.ClientTxnID != "txn-u-3-01" {
		t.Fatalf("unexpected client_txn_id: %#v", request)
	}
	if request.Summary == nil || *request.Summary != "First capture" {
		t.Fatalf("expected normalized summary, got %#v", request.Summary)
	}
	if request.OccurredAt == nil || request.OccurredAt.UTC().Format(time.RFC3339) != "2026-04-10T10:00:00Z" {
		t.Fatalf("expected normalized occurred_at, got %#v", request.OccurredAt)
	}

	_, apiErr = DecodeTimelineCreateRequest(bytes.NewBufferString(`{
		"client_txn_id": "txn-u-3-01-invalid",
		"timeline.capture_state": "rough"
	}`))
	if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
		t.Fatalf("expected invalid create payload for forbidden field, got %#v", apiErr)
	}
}

func TestPhase3_TimelinePatchCanonicalization_U_3_02(t *testing.T) {
	request, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 2,
		"client_txn_id": "txn-u-3-02",
		"changes": [
			{ "field_key": "timeline.summary", "value": "B" },
			{ "field_key": "timeline.details", "value": "A" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid patch request, got %#v", apiErr)
	}
	if len(request.CanonicalChange) != 2 {
		t.Fatalf("unexpected canonical changes length: %#v", request.CanonicalChange)
	}
	if request.CanonicalChange[0].FieldKey != "timeline.details" || request.CanonicalChange[1].FieldKey != "timeline.summary" {
		t.Fatalf("expected canonical field ordering, got %#v", request.CanonicalChange)
	}

	_, apiErr = DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 2,
		"client_txn_id": "txn-u-3-02-duplicate",
		"changes": [
			{ "field_key": "timeline.summary", "value": "B" },
			{ "field_key": "timeline.summary", "value": "C" }
		]
	}`))
	if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
		t.Fatalf("expected duplicate field rejection, got %#v", apiErr)
	}
}

func TestPhase3_PatchReplayHashCanonicalization_U_3_06(t *testing.T) {
	left, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 3,
		"client_txn_id": "txn-u-3-06",
		"changes": [
			{ "field_key": "timeline.summary", "value": "summary" },
			{ "field_key": "timeline.details", "value": "details" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("decode left patch: %#v", apiErr)
	}
	right, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 3,
		"client_txn_id": "txn-u-3-06",
		"changes": [
			{ "field_key": "timeline.details", "value": "details" },
			{ "field_key": "timeline.summary", "value": "summary" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("decode right patch: %#v", apiErr)
	}
	if !hashesEqual(TimelinePatchRequestHash(left), TimelinePatchRequestHash(right)) {
		t.Fatal("expected canonical patch hash to ignore outer changes[] order")
	}
}

func TestPhase3_ProjectionRowShape_U_3_07(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	recordedAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	editedAt := recordedAt.Add(2 * time.Minute)
	summary := "Summary"
	row := BuildRow(projectedRecord{
		RecordID:              recordID,
		IncidentID:            incidentID,
		RowVersion:            4,
		Summary:               &summary,
		RecordedAt:            recordedAt,
		EditedAt:              editedAt,
		SortTs:                recordedAt,
		CaptureState:          "rough",
		RecordedDay:           recordedAt,
		EvidenceCount:         0,
		HasEvidence:           false,
		HasUnresolvedMentions: false,
	})
	if row["record_id"] != recordID.String() || row["row_version"] != int64(4) {
		t.Fatalf("expected stable record identity, got %#v", row)
	}
	cells := row["cells"].(map[string]any)
	if cells["timeline.tags"].(map[string]any)["value"].(map[string]any)["kind"] != "collection_value_v1" {
		t.Fatalf("expected collection_value_v1 for tags, got %#v", cells["timeline.tags"])
	}
}

func TestPhase3_ChangedFieldKeysAreCanonical_U_3_08(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	recordedAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	summary := "Before"
	before := projectedRecord{
		RecordID:     recordID,
		IncidentID:   incidentID,
		RowVersion:   1,
		Summary:      &summary,
		RecordedAt:   recordedAt,
		EditedAt:     recordedAt,
		SortTs:       recordedAt,
		CaptureState: "rough",
		RecordedDay:  recordedAt,
	}
	afterSummary := "After"
	after := before
	after.RowVersion = 2
	after.Summary = &afterSummary
	after.CaptureState = "enriched"
	after.EditedAt = recordedAt.Add(time.Minute)

	changed := ComputeChangedFieldKeys(&before, after)
	if len(changed) == 0 {
		t.Fatal("expected changed field keys")
	}
	if changed[0] != "timeline.capture_state" {
		t.Fatalf("expected lexicographic ordering, got %#v", changed)
	}
}

func TestPhase3_ViewQueryValidation_U_3_09(t *testing.T) {
	_, apiErr := DecodeTimelineQueryRequest(bytes.NewBufferString(`{"sort":[]}`))
	if apiErr == nil || apiErr.Code != "invalid_view_query" {
		t.Fatalf("expected invalid view query rejection, got %#v", apiErr)
	}
}

func TestPhase3_MarkReviewedRequestValidation_U_3_03(t *testing.T) {
	request, apiErr := DecodeTimelineActionRequest(bytes.NewBufferString(`{
		"base_row_version": 2,
		"client_txn_id": "txn-u-3-03",
		"reason": "review note"
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid reviewed action request, got %#v", apiErr)
	}
	if request.BaseRowVersion != 2 || request.Reason == nil || *request.Reason != "review note" {
		t.Fatalf("unexpected reviewed action request: %#v", request)
	}
}

func TestPhase3_SupersedeRequestValidation_U_3_04(t *testing.T) {
	request, apiErr := DecodeTimelineSupersedeRequest(bytes.NewBufferString(`{
		"base_row_version": 4,
		"client_txn_id": "txn-u-3-04",
		"reason": "superseded",
		"replacement_record_id": "33333333-3333-3333-3333-333333333333"
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid supersede request, got %#v", apiErr)
	}
	if request.ReplacementRecordID == nil || request.ReplacementRecordID.String() != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("unexpected replacement_record_id: %#v", request)
	}

	_, apiErr = DecodeTimelineSupersedeRequest(bytes.NewBufferString(`{
		"base_row_version": 4,
		"client_txn_id": "txn-u-3-04-invalid",
		"reason": ""
	}`))
	if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
		t.Fatalf("expected required supersede reason validation, got %#v", apiErr)
	}
}

func TestPhase3_MaterialChangeClassification_U_3_10(t *testing.T) {
	current := sourceRecord{}
	next := current
	if hasMaterialChange(current, next) {
		t.Fatal("expected identical records to be non-material")
	}
	summary := "changed"
	next.Summary = &summary
	if !hasMaterialChange(current, next) {
		t.Fatal("expected summary edit to be capture-state-material")
	}
}
