package imports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func (s *service) registerJobHandlers() error {
	if s == nil || s.jobRunner == nil {
		return fmt.Errorf("imports worker registration unavailable")
	}
	if err := s.jobRunner.RegisterHandler(importDiscoveryJobHandlerName, s.executeDiscoveryJob); err != nil {
		return err
	}
	if err := s.jobRunner.RegisterHandler(importApplyJobHandlerName, s.executeApplyJob); err != nil {
		return err
	}
	return nil
}

func (s *service) executeDiscoveryJob(ctx context.Context, execution jobs.Execution) error {
	var payload discoveryJobHandlerPayload
	if err := s.decodeJobPayload(ctx, execution, &payload); err != nil {
		return err
	}
	sessionID, err := uuid.Parse(payload.ImportSessionID)
	if err != nil {
		return err
	}
	return s.completeDiscoveryJob(ctx, execution, sessionID)
}

func (s *service) completeDiscoveryJob(ctx context.Context, execution jobs.Execution, importSessionID uuid.UUID) error {
	jobID := execution.JobID()
	total := 1
	if !s.prepareClaimedJob(ctx, execution, total) {
		job, err := s.jobManager.ObserveExecution(ctx, execution)
		if err == nil && job.Status == jobs.StatusCanceled {
			return s.store.cancelDiscovery(ctx, importSessionID, s.now())
		}
		return nil
	}
	_, err := s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
		FinalCommitID: "import.discovery:" + jobID.String(),
		Execution:     execution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
				Code:    "import_session_discovered",
				Message: "Import session discovered.",
				ResourceRefs: []jobs.ResourceRef{{
					Kind:  "import_session",
					ID:    importSessionID.String(),
					Route: "/api/v1/import-sessions/" + importSessionID.String(),
				}},
			},
		},
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return s.store.markDiscoveredTx(ctx, tx, importSessionID, s.now())
		},
	})
	return err
}

func (s *service) decodeJobPayload(ctx context.Context, execution jobs.Execution, target any) error {
	payload, err := s.jobManager.HandlerPayload(ctx, execution)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return fmt.Errorf("missing import job payload")
	}
	return json.Unmarshal(payload, target)
}

func (s *service) prepareClaimedJob(ctx context.Context, execution jobs.Execution, total int) bool {
	if total <= 0 {
		total = 1
	}
	if _, err := s.jobManager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 0, Total: &total}, nil); err == nil {
		return true
	}
	job, err := s.jobManager.ObserveExecution(ctx, execution)
	if err != nil {
		return false
	}
	switch job.Status {
	case jobs.StatusCancelRequested:
		_, _ = s.jobManager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{
			Progress: jobs.Progress{Completed: 0, Total: &total},
		})
		return false
	default:
		return false
	}
}
