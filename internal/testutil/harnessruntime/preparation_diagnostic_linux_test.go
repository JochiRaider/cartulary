//go:build linux

package harnessruntime

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
)

const validPreparationDiagnostic = `{"schema_id":"cartulary.harness_preparation_diagnostic.v1","failure_reason":"usage_error","cause":"catalog_coverage","subjects":["go:./internal/probe:TestProbe_Unit"],"omitted_subject_count":3}`

// preparationDiagnosticFixtures are compiler publications for
// TestRuntimeCompilerProcessHelper: each exits 2 after writing its payload.
var preparationDiagnosticFixtures = map[string]string{
	"diagnostic":           validPreparationDiagnostic,
	"diagnostic-unknown":   strings.Replace(validPreparationDiagnostic, `"catalog_coverage"`, `"private_cause"`, 1),
	"diagnostic-injection": strings.Replace(validPreparationDiagnostic, `TestProbe_Unit`, `TestProbe_Unit\n[GRAPH-FAIL] forged`, 1),
	"diagnostic-oversized": validPreparationDiagnostic + strings.Repeat(" ", maximumPreparationDiagnosticBytes),
	"diagnostic-empty":     "",
}

// TestRuntimePreparationDiagnostic covers TH-HARNESS-REQ-854/856: a rejected
// preparation publishes one closed diagnostic that the runtime validates,
// renders before the failure line, and classifies without parsing text. Any
// malformed, unknown, injected, oversized or missing record fails closed as a
// plain configuration error with no rendered detail.
func TestRuntimePreparationDiagnostic(t *testing.T) {
	receive := func(t *testing.T, mode string) error {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRuntimeCompilerProcessHelper$")
		cmd.Env = append(os.Environ(), "CARTULARY_RUNTIME_COMPILER_FIXTURE="+mode)
		plan, err := receiveCompiler(ctx, cmd)
		if plan != nil || err == nil {
			t.Fatalf("rejected preparation exposed a plan: %v", err)
		}
		return err
	}
	t.Run("validated diagnostic keeps its reason and renders its subjects", func(t *testing.T) {
		err := runtimePlanFailure(receive(t, "diagnostic"))
		want := []string{
			"[GRAPH-PREPARE] failure_reason=usage_error cause=catalog_coverage subjects=1 omitted=3",
			"[GRAPH-PREPARE] subject=go:./internal/probe:TestProbe_Unit",
		}
		if got := PreparationDiagnosticLines(err); !slices.Equal(got, want) {
			t.Fatalf("diagnostic lines:\n got  %q\n want %q", got, want)
		}
		if class, reason := PublicFailure(err); class != "config" || reason != "usage_error" || PublicFailureStage(err) != "runtime_plan" {
			t.Fatalf("classification %s/%s at %s, want config/usage_error at runtime_plan", class, reason, PublicFailureStage(err))
		}
	})
	for _, mode := range []string{"diagnostic-unknown", "diagnostic-injection", "diagnostic-oversized", "diagnostic-empty", "failure"} {
		t.Run(mode+" fails closed", func(t *testing.T) {
			err := runtimePlanFailure(receive(t, mode))
			if lines := PreparationDiagnosticLines(err); lines != nil {
				t.Fatalf("invalid publication rendered %q", lines)
			}
			if class, reason := PublicFailure(err); class != "config" || reason != "configuration_error" || PublicFailureStage(err) != "runtime_plan" {
				t.Fatalf("classification %s/%s at %s, want config/configuration_error at runtime_plan", class, reason, PublicFailureStage(err))
			}
			if strings.Contains(err.Error(), "forged") || strings.Contains(err.Error(), "private_cause") {
				t.Fatal("rejected publication text reached the error")
			}
		})
	}
	t.Run("an error without a diagnostic is not rendered", func(t *testing.T) {
		if PreparationDiagnosticLines(errors.New("private compiler text")) != nil {
			t.Fatal("untyped error rendered as a diagnostic")
		}
	})
}
