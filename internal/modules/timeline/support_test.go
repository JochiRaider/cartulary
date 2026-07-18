package timeline

import (
	"bytes"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUnit_CreateRequestCoverage(t *testing.T) {
	t.Run("zero field create is allowed for timeline", func(t *testing.T) {
		request, apiErr := DecodeTimelineCreateRequest(bytes.NewBufferString(`{
			"client_txn_id": "txn-support-phase3-zero"
		}`))
		if apiErr != nil {
			t.Fatalf("expected zero-field timeline create to decode, got %#v", apiErr)
		}
		if CreateRequestHasUserValue(request) {
			t.Fatalf("expected zero-field create request to have no user values, got %#v", request)
		}
	})

	t.Run("one non-empty value remains valid and normalizes", func(t *testing.T) {
		request, apiErr := DecodeTimelineCreateRequest(bytes.NewBufferString(`{
			"client_txn_id": "txn-support-phase3-one-value",
			"timeline.activity_synopsis_text": "  First capture  "
		}`))
		if apiErr != nil {
			t.Fatalf("expected valid create request, got %#v", apiErr)
		}
		if request.ActivitySynopsisText == nil {
			t.Fatalf("expected summary value, got %#v", request)
		}
		requireWritableStringNormalization(t, *request.ActivitySynopsisText, "  First capture  ")
	})

	t.Run("client-owned system fields fail closed", func(t *testing.T) {
		_, apiErr := DecodeTimelineCreateRequest(bytes.NewBufferString(`{
			"client_txn_id": "txn-support-phase3-invalid",
			"timeline.capture_state": "reviewed"
		}`))
		if apiErr == nil {
			t.Fatal("expected direct write to timeline.capture_state to fail")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"timeline.capture_state",
			"unknown_field",
		)
	})
}

func TestUnit_InitialStateVocabulary(t *testing.T) {
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

func TestUnit_CaptureStateHelpers(t *testing.T) {
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

	nextState, err = CaptureStateAfterMaterialPatch(captureStateReviewed)
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

func TestUnit_PatchRequestHashNormalization(t *testing.T) {
	left, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v2",
		"base_row_version": 3,
		"client_txn_id": "txn-support-phase3-hash",
		"changes": [
			{ "field_key": "timeline.activity_synopsis_text", "value": "summary" },
			{ "field_key": "timeline.raw_activity_text", "value": "details" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("decode left patch: %#v", apiErr)
	}
	right, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v2",
		"base_row_version": 3,
		"client_txn_id": "txn-support-phase3-hash",
		"changes": [
			{ "field_key": "timeline.raw_activity_text", "value": "details" },
			{ "field_key": "timeline.activity_synopsis_text", "value": "summary" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("decode right patch: %#v", apiErr)
	}
	if !hashesEqual(TimelinePatchRequestHash(left), TimelinePatchRequestHash(right)) {
		t.Fatal("expected canonical patch request hash to ignore outer changes[] order")
	}

	changed := "changed"
	if hashesEqual(TimelinePatchRequestHash(left), TimelinePatchRequestHash(PatchRequest{
		ViewSchemaID:   right.ViewSchemaID,
		BaseRowVersion: right.BaseRowVersion,
		ClientTxnID:    right.ClientTxnID,
		CanonicalChange: []PatchChange{
			{FieldKey: "timeline.raw_activity_text", TextValue: &changed},
			{FieldKey: "timeline.activity_synopsis_text", TextValue: right.CanonicalChange[1].TextValue},
		},
	})) {
		t.Fatal("expected divergent normalized patch request hash to differ")
	}
}

func TestUnit_PayloadBuildersExposeStableShapes(t *testing.T) {
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
		ActivitySortTS:      &recordedAt,
		CaptureState:        captureStateReviewed,
		ReplacementRecordID: &replacementID,
		DateEnteredSortDay:  &recordedAt,
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

func TestUnit_SupersedeGuardAndHashHelpers(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	otherIncidentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	otherRecordID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	anotherReplacementID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	if err := ValidateSupersedeReplacement(recordID, incidentID, &recordID, &incidentID); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("self replacement must be rejected, got %v", err)
	} else {
		var transitionErr *IllegalTransitionError
		if !errors.As(err, &transitionErr) {
			t.Fatalf("expected typed illegal transition, got %T", err)
		}
		if !slices.Equal(transitionErr.ViolatedGuards, []string{supersedeGuardReplacementDifferent}) {
			t.Fatalf("unexpected guard tokens: got %#v", transitionErr.ViolatedGuards)
		}
	}
	if err := ValidateSupersedeReplacement(recordID, incidentID, &otherRecordID, &otherIncidentID); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("cross-incident replacement must be rejected, got %v", err)
	} else {
		var transitionErr *IllegalTransitionError
		if !errors.As(err, &transitionErr) {
			t.Fatalf("expected typed illegal transition, got %T", err)
		}
		if !slices.Equal(transitionErr.ViolatedGuards, []string{supersedeGuardReplacementVisibleActiveSameIncident}) {
			t.Fatalf("unexpected guard tokens: got %#v", transitionErr.ViolatedGuards)
		}
	}
	if err := ValidateSupersedeReplacement(recordID, incidentID, &otherRecordID, &incidentID); err != nil {
		t.Fatalf("legal replacement should be allowed, got %v", err)
	}

	reason := "superseded"
	left := TimelineActionRequestHash(4, "txn-support-phase3-supersede", &reason, &otherRecordID)
	right := TimelineActionRequestHash(4, "txn-support-phase3-supersede", &reason, &otherRecordID)
	if !hashesEqual(left, right) {
		t.Fatal("expected identical supersede request hashes to match")
	}
	if !hashesEqual(left, TimelineActionRequestHash(4, "txn-support-phase3-supersede-different-key", &reason, &otherRecordID)) {
		t.Fatal("expected client_txn_id changes to be excluded from the normalized request hash")
	}

	if hashesEqual(left, TimelineActionRequestHash(4, "txn-support-phase3-supersede", &reason, &anotherReplacementID)) {
		t.Fatal("expected replacement id changes to alter the normalized request hash")
	}

	differentReason := "superseded differently"
	if hashesEqual(left, TimelineActionRequestHash(4, "txn-support-phase3-supersede", &differentReason, &otherRecordID)) {
		t.Fatal("expected reason changes to alter the normalized request hash")
	}
}
