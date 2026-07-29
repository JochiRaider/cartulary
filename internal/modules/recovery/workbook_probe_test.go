package recovery_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

type workbookProbeExecutor struct {
	calls      int
	profile    string
	incidentID uuid.UUID
	result     workbookprobe.Result
	err        error
}

func (executor *workbookProbeExecutor) ExecuteDefault(
	_ context.Context,
	profile string,
	incidentID uuid.UUID,
) (workbookprobe.Result, error) {
	executor.calls++
	executor.profile = profile
	executor.incidentID = incidentID
	return executor.result, executor.err
}

func TestRestoreVerificationWorkbookProbe(t *testing.T) {
	ctx := context.Background()
	t.Run("nil result fails", func(t *testing.T) {
		err := (recovery.RestoreVerificationWorkbookProbe{}).ProbeRestoredBackup(ctx, nil)
		if err == nil || !strings.Contains(err.Error(), "restore result is required") {
			t.Fatalf("nil result error got %v", err)
		}
	})

	t.Run("zero incidents skips without executor", func(t *testing.T) {
		result := recovery.RestoreResult{}
		if err := (recovery.RestoreVerificationWorkbookProbe{}).ProbeRestoredBackup(ctx, &result); err != nil {
			t.Fatalf("zero incidents should skip probe: %v", err)
		}
		if result.WorkbookProbe != nil {
			t.Fatalf("zero incidents unexpectedly recorded a workbook probe: %#v", result.WorkbookProbe)
		}
	})

	t.Run("selected incident executes once and binds result identity", func(t *testing.T) {
		incidentID := uuid.MustParse("00000000-0000-0000-0000-000000009101")
		selected := incidentID.String()
		executor := &workbookProbeExecutor{result: workbookprobe.Result{
			RegistrationID: "timeline.base_restore_probe.v1",
			ViewSchemaID:   "cartulary.view.timeline.v2",
			RowCount:       0,
		}}
		result := recovery.RestoreResult{SelectedIncidentID: &selected}
		if err := (recovery.RestoreVerificationWorkbookProbe{Executor: executor}).ProbeRestoredBackup(ctx, &result); err != nil {
			t.Fatalf("empty workbook query should not fail probe: %v", err)
		}
		if executor.calls != 1 || executor.profile != workbookprobe.BaseProfile || executor.incidentID != incidentID {
			t.Fatalf("probe execution got calls=%d profile=%q incident=%s", executor.calls, executor.profile, executor.incidentID)
		}
		if result.WorkbookProbe == nil || *result.WorkbookProbe != executor.result {
			t.Fatalf("bound workbook probe result got %#v want %#v", result.WorkbookProbe, executor.result)
		}
	})
}
