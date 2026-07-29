package recovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

var ErrWorkbookProbeFailed = errors.New("recovery: workbook probe failed")

type RestoreVerificationWorkbookProbe struct {
	Executor workbookprobe.Executor
}

func (probe RestoreVerificationWorkbookProbe) ProbeRestoredBackup(ctx context.Context, result *RestoreResult) error {
	if result == nil {
		return fmt.Errorf("%w: restore result is required", ErrWorkbookProbeFailed)
	}
	if result.SelectedIncidentID == nil {
		result.WorkbookProbe = nil
		return nil
	}
	if probe.Executor == nil {
		return fmt.Errorf("%w: restore verification workbook probe requires executor", ErrWorkbookProbeFailed)
	}
	incidentID, err := uuid.Parse(*result.SelectedIncidentID)
	if err != nil {
		return fmt.Errorf("%w: selected incident_id is invalid", ErrWorkbookProbeFailed)
	}
	executed, err := probe.Executor.ExecuteDefault(ctx, workbookprobe.BaseProfile, incidentID)
	result.WorkbookProbe = &executed
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWorkbookProbeFailed, err)
	}
	return nil
}
