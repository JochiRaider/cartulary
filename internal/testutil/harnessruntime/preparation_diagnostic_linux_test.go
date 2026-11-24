//go:build linux

package harnessruntime

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
)

// TestRuntimePreparationDiagnostic covers TH-HARNESS-REQ-854/856: a rejected
// preparation carries one closed diagnostic that the runtime renders before
// the failure line and classifies without parsing text. An error without a
// diagnostic is a plain configuration error with no rendered detail.
func TestRuntimePreparationDiagnostic(t *testing.T) {
	t.Run("a diagnostic keeps its reason and renders its subjects", func(t *testing.T) {
		err := runtimePlanFailure(plannerFailure(fmt.Errorf("%w: unknown owner probe", harnessplan.ErrSelection)))
		want := []string{
			"[GRAPH-PREPARE] failure_reason=usage_error cause=command_selection subjects=1 omitted=0",
			"[GRAPH-PREPARE] subject=" + harnessplan.ErrSelection.Error() + ": unknown owner probe",
		}
		if got := PreparationDiagnosticLines(err); !slices.Equal(got, want) {
			t.Fatalf("diagnostic lines:\n got  %q\n want %q", got, want)
		}
		if class, reason := PublicFailure(err); class != "config" || reason != "usage_error" || PublicFailureStage(err) != "runtime_plan" {
			t.Fatalf("classification %s/%s at %s, want config/usage_error at runtime_plan", class, reason, PublicFailureStage(err))
		}
	})
	t.Run("an error without a diagnostic is a configuration error", func(t *testing.T) {
		err := runtimePlanFailure(errors.New("private planner text"))
		if lines := PreparationDiagnosticLines(err); lines != nil {
			t.Fatalf("untyped error rendered %q", lines)
		}
		if class, reason := PublicFailure(err); class != "config" || reason != "configuration_error" || PublicFailureStage(err) != "runtime_plan" {
			t.Fatalf("classification %s/%s at %s, want config/configuration_error at runtime_plan", class, reason, PublicFailureStage(err))
		}
	})
}
