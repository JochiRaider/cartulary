package timeline

import (
	"errors"
	"testing"
)

func TestPhase3_CaptureStateTransitions_U_3_03(t *testing.T) {
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

func TestPhase3_ReviewedRowsDemoteAndSupersededRowsAreTerminal_U_3_04(t *testing.T) {
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
