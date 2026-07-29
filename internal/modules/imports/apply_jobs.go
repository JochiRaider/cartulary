package imports

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func (s *Service) executeApplyJob(ctx context.Context, jobID uuid.UUID) error {
	var payload applyJobHandlerPayload
	if err := s.decodeJobPayload(ctx, jobID, &payload); err != nil {
		return err
	}
	incidentID, err := uuid.Parse(payload.IncidentID)
	if err != nil {
		return err
	}
	sessionID, err := uuid.Parse(payload.ImportSessionID)
	if err != nil {
		return err
	}
	actorID, err := uuid.Parse(payload.ActorUserID)
	if err != nil {
		return err
	}
	selected, err := parseUUIDStrings(payload.SelectedUnitIDs)
	if err != nil {
		return err
	}
	actor, err := s.authStore.GetUserByID(ctx, actorID)
	if err != nil {
		return err
	}
	job, err := s.jobManager.Get(ctx, jobID)
	if err != nil {
		return err
	}
	return s.completeApplyJob(ctx, actor, ApplyStartResult{
		Job:             job,
		IncidentID:      incidentID,
		ImportSessionID: sessionID,
		ClientTxnID:     payload.ClientTxnID,
		SelectedUnitIDs: selected,
	})
}

func (s *Service) completeApplyJob(ctx context.Context, actor authn.UserRecord, start ApplyStartResult) error {
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return err
	}
	total := len(start.SelectedUnitIDs)
	if !s.markJobRunningOrResume(ctx, jobID, total) {
		s.cancelApplySessionForTerminalJob(ctx, jobID, start)
		return nil
	}
	units, err := s.store.GetApplyUnits(ctx, start.ImportSessionID, start.SelectedUnitIDs)
	if err != nil {
		s.failApplyJob(ctx, jobID, start, "import_apply_failed", err)
		return nil
	}
	extensionRefs := make([]jobs.ResourceRef, 0)
	completed := 0
	failed := 0
	seenUnits := make(map[uuid.UUID]struct{}, len(units))
	for _, unit := range units {
		seenUnits[unit.UnitID] = struct{}{}
		if s.jobCancelRequested(ctx, jobID) {
			if err := s.store.CancelApply(ctx, start.ImportSessionID, start.SelectedUnitIDs, s.now()); err != nil {
				return err
			}
			_, err := s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
				JobID:    jobID,
				Progress: jobs.Progress{Completed: completed, Total: &total},
			})
			return err
		}
		refs, err := s.applyUnit(ctx, actor, start, unit)
		if err != nil {
			if statusErr := s.store.MarkApplyUnitStatus(ctx, start.ImportSessionID, unit.UnitID, "failed", s.now()); statusErr != nil {
				s.failApplyJob(ctx, jobID, start, "import_apply_failed", statusErr)
				return nil
			}
			failed++
			completed++
			continue
		}
		if err := s.store.MarkApplyUnitStatus(ctx, start.ImportSessionID, unit.UnitID, "applied", s.now()); err != nil {
			s.failApplyJob(ctx, jobID, start, "import_apply_failed", err)
			return nil
		}
		extensionRefs = append(extensionRefs, refs...)
		completed++
	}
	for _, unitID := range start.SelectedUnitIDs {
		if _, ok := seenUnits[unitID]; ok {
			continue
		}
		if err := s.store.MarkApplyUnitStatus(ctx, start.ImportSessionID, unitID, "failed", s.now()); err != nil {
			s.failApplyJob(ctx, jobID, start, "import_apply_failed", err)
			return nil
		}
		failed++
		completed++
	}
	if completed-failed == 0 {
		s.failApplyJob(ctx, jobID, start, "import_apply_failed", fmt.Errorf("all selected import units failed"))
		return nil
	}
	status := "applied"
	code := "import_session_applied"
	if failed > 0 {
		status = "partially_applied"
		code = "import_session_partially_applied"
	}
	_, err = s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
		FinalCommitID: "import.apply:" + jobID.String(),
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: total, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:         code,
				Message:      "Import session applied.",
				ResourceRefs: importApplyResourceRefs(start.ImportSessionID, extensionRefs),
			},
		},
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return s.store.CompleteApplyTx(ctx, tx, start.ImportSessionID, start.SelectedUnitIDs, status, s.now())
		},
	})
	return err
}

func (s *Service) failApplyJob(ctx context.Context, jobID uuid.UUID, start ApplyStartResult, code string, err error) {
	_ = s.store.FailApply(ctx, start.ImportSessionID, start.SelectedUnitIDs, s.now())
	_, _ = s.jobManager.CompleteFailed(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 0, Total: intPtr(len(start.SelectedUnitIDs))},
		ErrorSummary: &jobs.ErrorSummary{
			Code:      code,
			Message:   "Import apply failed.",
			Retryable: false,
			Details:   map[string]any{},
		},
	})
	_ = err
}

func (s *Service) cancelApplySessionForTerminalJob(ctx context.Context, jobID uuid.UUID, start ApplyStartResult) {
	job, err := s.jobManager.Get(ctx, jobID)
	if err == nil && job.Status == jobs.StatusCanceled {
		_ = s.store.CancelApply(ctx, start.ImportSessionID, start.SelectedUnitIDs, s.now())
	}
}

func (s *Service) jobCancelRequested(ctx context.Context, jobID uuid.UUID) bool {
	job, err := s.jobManager.Get(ctx, jobID)
	return err == nil && job.Status == jobs.StatusCancelRequested
}
