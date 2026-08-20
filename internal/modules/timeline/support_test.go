package timeline

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

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

func TestUnit_PayloadBuildersExposeStableShapes(t *testing.T) {
	recordVersionID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	if got, want := versionid.Format(recordVersionID, 2), "timeline_record:"+recordVersionID.String()+":2"; got != want {
		t.Fatalf("Timeline version ID changed: got %q want %q", got, want)
	}

	t.Run("current facade method inventory is explicit", func(t *testing.T) {
		facadeType := reflect.TypeOf((*Facade)(nil))
		got := make([]string, 0, facadeType.NumMethod())
		for index := range facadeType.NumMethod() {
			got = append(got, facadeType.Method(index).Name)
		}
		want := []string{
			"ApplyClipboardPaste",
			"ApplyFillDown",
			"ApplyMultiRowTagAssignment",
			"CreateImportRowTx",
			"CreateImportedRow",
			"CreateRow",
			"GetTimeConversionProfile",
			"MarkReviewedRow",
			"ParseConflictToken",
			"PatchRow",
			"PutTimeConversionProfile",
			"RecordIncident",
			"ResolveConflict",
			"SupersedeRow",
		}
		if !slices.Equal(got, want) {
			t.Fatalf("Timeline facade method set changed: got %#v want %#v", got, want)
		}
	})

	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	changeSetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	replacementID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	recordedAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	reason := "reviewed in workbook"
	projected := workbookprojection.DerivedRecord{
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

	if got, want := fmt.Sprintf("%x", valuecodec.CanonicalJSONSHA256(mutationPayload)), "c9fb6ab2d93f42c7375dc3c10bc41c9a4f90503b8d4cd5d2721c796576d976db"; got != want {
		t.Fatalf("canonical mutation payload hash changed: got %s want %s", got, want)
	}
	if got, want := fmt.Sprintf("%x", valuecodec.CanonicalJSONSHA256(actionPayload)), "8ac588868d66bef540c019c9c55af59e2cd032029184af3567d2b0136458b653"; got != want {
		t.Fatalf("canonical action payload hash changed: got %s want %s", got, want)
	}
}

func TestUnit_SupersedeGuards(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	otherIncidentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	otherRecordID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

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

}
