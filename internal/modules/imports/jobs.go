package imports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func (s *Service) registerJobHandlers() error {
	if s == nil || s.jobRunner == nil {
		return nil
	}
	if err := s.jobRunner.RegisterHandler(importDiscoveryJobHandlerName, s.executeDiscoveryJob); err != nil && !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		return err
	}
	if err := s.jobRunner.RegisterHandler(importApplyJobHandlerName, s.executeApplyJob); err != nil && !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		return err
	}
	return nil
}

func (s *Service) recoverImportJobs(ctx context.Context) error {
	if s == nil || s.jobRunner == nil {
		return nil
	}
	if err := s.jobRunner.RecoverHandler(ctx, importDiscoveryJobHandlerName); err != nil {
		return err
	}
	return s.jobRunner.RecoverHandler(ctx, importApplyJobHandlerName)
}

func (s *Service) executeDiscoveryJob(ctx context.Context, jobID uuid.UUID) error {
	var payload discoveryJobHandlerPayload
	if err := s.decodeJobPayload(ctx, jobID, &payload); err != nil {
		return err
	}
	sessionID, err := uuid.Parse(payload.ImportSessionID)
	if err != nil {
		return err
	}
	return s.completeDiscoveryJob(ctx, jobID, sessionID)
}

func (s *Service) completeDiscoveryJob(ctx context.Context, jobID uuid.UUID, importSessionID uuid.UUID) error {
	total := 1
	if !s.markJobRunningOrResume(ctx, jobID, total) {
		job, err := s.jobManager.Get(ctx, jobID)
		if err == nil && job.Status == jobs.StatusCanceled {
			return s.store.CancelDiscovery(ctx, importSessionID, s.now())
		}
		return nil
	}
	_, err := s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
		FinalCommitID: "import.discovery:" + jobID.String(),
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
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
			return s.store.MarkDiscoveredTx(ctx, tx, importSessionID, s.now())
		},
	})
	return err
}

func (s *Service) decodeJobPayload(ctx context.Context, jobID uuid.UUID, target any) error {
	payload, err := s.jobManager.HandlerPayload(ctx, jobID)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return fmt.Errorf("missing import job payload")
	}
	return json.Unmarshal(payload, target)
}

func (s *Service) markJobRunningOrResume(ctx context.Context, jobID uuid.UUID, total int) bool {
	if total <= 0 {
		total = 1
	}
	if _, err := s.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil); err == nil {
		return true
	}
	job, err := s.jobManager.Get(ctx, jobID)
	if err != nil {
		return false
	}
	switch job.Status {
	case jobs.StatusRunning:
		return true
	case jobs.StatusCancelRequested:
		_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 0, Total: &total},
		})
		return false
	default:
		return false
	}
}
