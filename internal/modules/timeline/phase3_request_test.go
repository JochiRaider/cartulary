package timeline

import (
	"bytes"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase3_U_3_01_CreateRequiresOneNonEmptyUserValue(t *testing.T) {
	request, apiErr := DecodeTimelineCreateRequest(bytes.NewBufferString(`{
		"client_txn_id": "txn-u-3-01",
		"timeline.summary": "  First capture  "
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid create request, got %#v", apiErr)
	}
	requireWritableStringNormalization(t, *request.Summary, "First capture")

	_, apiErr = DecodeTimelineCreateRequest(bytes.NewBufferString(`{
		"client_txn_id": "txn-u-3-01-empty"
	}`))
	if apiErr == nil {
		t.Fatal("expected empty create payload rejection")
	}
	requireClosedVocabularyRejected(t, apiErr.Code, apiErr.Details, "payload", "at_least_one_value_required")

	_, apiErr = DecodeTimelineCreateRequest(bytes.NewBufferString(`{
		"client_txn_id": "txn-u-3-01-invalid",
		"record_id": "11111111-1111-1111-1111-111111111111"
	}`))
	if apiErr == nil {
		t.Fatal("expected server-owned record_id rejection")
	}
	requireClosedVocabularyRejected(t, apiErr.Code, apiErr.Details, "record_id", "unknown_field")
}

func TestPhase3_U_3_02_UsesRoughAsInitialCaptureStateAndRejectsObsoleteTokens(t *testing.T) {
	if InitialCaptureState() != captureStateRough {
		t.Fatalf("unexpected initial capture state: %q", InitialCaptureState())
	}
	for _, supported := range []string{
		captureStateRough,
		captureStateEnriched,
		captureStateReviewed,
		captureStateSuperseded,
	} {
		if !IsSupportedCaptureState(supported) {
			t.Fatalf("expected supported capture_state %q", supported)
		}
	}
	for _, obsolete := range []string{"developing", "complete"} {
		if IsSupportedCaptureState(obsolete) {
			t.Fatalf("obsolete capture_state must be rejected: %q", obsolete)
		}
	}
}

func TestPhase3_U_3_03_FirstMaterialMutationEnrichesAndReviewRequiresReviewableState(t *testing.T) {
	nextState, err := CaptureStateAfterMaterialPatch(captureStateRough)
	if err != nil {
		t.Fatalf("rough patch transition: %v", err)
	}
	if nextState != captureStateEnriched {
		t.Fatalf("expected rough -> enriched, got %q", nextState)
	}

	if !CaptureStateAllowsMarkReviewed(captureStateRough) || !CaptureStateAllowsMarkReviewed(captureStateEnriched) {
		t.Fatal("rough and enriched rows must be reviewable")
	}
	if CaptureStateAllowsMarkReviewed(captureStateReviewed) || CaptureStateAllowsMarkReviewed(captureStateSuperseded) {
		t.Fatal("reviewed and superseded rows must not allow mark-reviewed")
	}
}

func TestPhase3_U_3_04_ReviewedRowsDemoteAndSupersededRowsBecomeTerminal(t *testing.T) {
	nextState, err := CaptureStateAfterMaterialPatch(captureStateReviewed)
	if err != nil {
		t.Fatalf("reviewed patch transition: %v", err)
	}
	if nextState != captureStateEnriched {
		t.Fatalf("expected reviewed -> enriched on material edit, got %q", nextState)
	}
	if !CaptureStateAllowsSupersede(captureStateReviewed) {
		t.Fatal("reviewed rows must allow legal supersede")
	}
	if _, err := CaptureStateAfterMaterialPatch(captureStateSuperseded); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("expected superseded rows to reject ordinary patch semantics, got %v", err)
	}
}

func TestPhase3_U_3_06_PatchRequiresBaseVersionClientTxnAndCanonicalChanges(t *testing.T) {
	request, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 2,
		"client_txn_id": "txn-u-3-06",
		"changes": [
			{ "field_key": "timeline.summary", "value": "B" },
			{ "field_key": "timeline.details", "value": "A" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid patch request, got %#v", apiErr)
	}
	fieldKeys := []string{
		request.CanonicalChange[0].FieldKey,
		request.CanonicalChange[1].FieldKey,
	}
	requireFieldKeyConformance(t, fieldKeys, []string{
		"timeline.details",
		"timeline.occurred_at",
		"timeline.source_text",
		"timeline.summary",
	})

	cases := []struct {
		name   string
		body   string
		field  string
		reason string
	}{
		{
			name:   "missing view schema",
			body:   `{"base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"}]}`,
			field:  "view_schema_id",
			reason: "missing_required_field",
		},
		{
			name:   "missing base row version",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"}]}`,
			field:  "base_row_version",
			reason: "missing_required_field",
		},
		{
			name:   "missing client txn",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"changes":[{"field_key":"timeline.summary","value":"x"}]}`,
			field:  "client_txn_id",
			reason: "missing_required_field",
		},
		{
			name:   "empty changes",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[]}`,
			field:  "changes",
			reason: "empty_changes",
		},
		{
			name:   "duplicate field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"},{"field_key":"timeline.summary","value":"y"}]}`,
			field:  "changes",
			reason: "duplicate_field_key",
		},
		{
			name:   "unsupported field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.capture_state","value":"reviewed"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(tc.body))
			if apiErr == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
			requireClosedVocabularyRejected(t, apiErr.Code, apiErr.Details, tc.field, tc.reason)
		})
	}
}

func TestPhase3_U_3_07_ReplayHashesMatchNormalizedPayloadsAndRejectDivergence(t *testing.T) {
	left, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 3,
		"client_txn_id": "txn-u-3-07",
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
		"client_txn_id": "txn-u-3-07",
		"changes": [
			{ "field_key": "timeline.details", "value": "details" },
			{ "field_key": "timeline.summary", "value": "summary" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("decode right patch: %#v", apiErr)
	}
	if !hashesEqual(TimelinePatchRequestHash(left), TimelinePatchRequestHash(right)) {
		t.Fatal("expected canonical patch replay hash to ignore outer changes[] order")
	}

	divergent := right
	reason := "changed"
	if hashesEqual(TimelinePatchRequestHash(left), TimelinePatchRequestHash(PatchRequest{
		ViewSchemaID:   divergent.ViewSchemaID,
		BaseRowVersion: divergent.BaseRowVersion,
		ClientTxnID:    divergent.ClientTxnID,
		CanonicalChange: []PatchChange{
			{FieldKey: "timeline.details", TextValue: &reason},
			{FieldKey: "timeline.summary", TextValue: right.CanonicalChange[1].TextValue},
		},
	})) {
		t.Fatal("expected divergent replay hash to differ")
	}
}

func TestPhase3_U_3_08_ProjectionRowsKeepStableBindingAndDerivedFields(t *testing.T) {
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

func TestPhase3_U_3_09_MutationPayloadsExposeRevisionIdentityAndHistoryEnvelope(t *testing.T) {
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

func TestPhase3_U_3_10_SupersedeReplacementValidationRejectsIllegalTargetsAndHashesIdempotently(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	otherIncidentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	otherRecordID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

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
		t.Fatal("expected identical supersede replays to hash identically")
	}

	differentReason := "superseded differently"
	if hashesEqual(left, TimelineActionRequestHash(4, "txn-u-3-10", &differentReason, &otherRecordID)) {
		t.Fatal("expected divergent supersede replay to hash differently")
	}
}

func requireClosedVocabularyRejected(t testing.TB, code string, details map[string]any, wantField string, wantReasonCode string) {
	t.Helper()
	if code != "invalid_mutation_payload" && code != "invalid_view_query" {
		t.Fatalf("unexpected rejection code: %q", code)
	}
	if details["field"] != wantField {
		t.Fatalf("unexpected rejection field: got %v want %q", details["field"], wantField)
	}
	if details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected rejection reason_code: got %v want %q", details["reason_code"], wantReasonCode)
	}
}

func requireWritableStringNormalization(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected normalized string: got %q want %q", got, want)
	}
}

func requireFieldKeyConformance(t testing.TB, fieldKeys []string, allowed []string) {
	t.Helper()
	if !slices.IsSorted(fieldKeys) {
		t.Fatalf("expected sorted field keys, got %v", fieldKeys)
	}
	for _, fieldKey := range fieldKeys {
		if !slices.Contains(allowed, fieldKey) {
			t.Fatalf("unexpected field key %q not in allowed set %v", fieldKey, allowed)
		}
	}
}
