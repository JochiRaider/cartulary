package imports

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
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
	job, getErr := s.jobManager.Get(ctx, jobID)
	resumingCancellation := getErr == nil && job.Status == jobs.StatusCancelRequested
	if !resumingCancellation && !s.markJobRunningOrResume(ctx, jobID, total) {
		return nil
	}
	units, err := s.store.GetApplyUnits(ctx, start.ImportSessionID, start.SelectedUnitIDs)
	if err != nil {
		return err
	}
	if len(units) != len(start.SelectedUnitIDs) {
		return fmt.Errorf("selected import unit set changed")
	}
	for _, unit := range units {
		if s.jobCancelRequested(ctx, jobID) {
			if _, err := s.store.recordTerminalUnitOutcome(
				ctx,
				start,
				unit.UnitID,
				actor.ID,
				"canceled",
				"import_apply_canceled",
				"cancel_requested",
				s.now(),
			); err != nil {
				return err
			}
			continue
		}
		if _, err := s.applyUnit(ctx, actor, start, unit.UnitID); err == nil {
			continue
		} else if errors.Is(err, errUnitCommitIndeterminate) {
			return err
		} else {
			status := "failed"
			errorCode, reasonCode := importUnitFailure(err)
			if errors.Is(err, errImportUnitCanceled) {
				status = "canceled"
			}
			if _, persistErr := s.store.recordTerminalUnitOutcome(
				ctx,
				start,
				unit.UnitID,
				actor.ID,
				status,
				errorCode,
				reasonCode,
				s.now(),
			); persistErr != nil {
				return persistErr
			}
		}
	}
	return s.finalizeApplyJob(ctx, start)
}

func (s *Service) finalizeApplyJob(ctx context.Context, start ApplyStartResult) error {
	finalization, err := s.store.prepareApplyFinalization(ctx, start)
	if err != nil {
		return err
	}
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return err
	}
	total := len(start.SelectedUnitIDs)
	transition := jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: total, Total: &total},
	}
	mutate := func(ctx context.Context, tx pgx.Tx) error {
		return s.store.finalizeApplyFromOutcomesTx(
			ctx,
			tx,
			start,
			finalization,
			s.now(),
		)
	}
	switch finalization.JobStatus {
	case jobs.StatusSucceeded:
		transition.ResultSummary = &jobs.ResultSummary{
			Code:         finalization.ResultCode,
			Message:      "Import session applied.",
			ResourceRefs: finalization.ResourceRefs,
		}
		_, err = s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
			FinalCommitID: "import.apply:" + jobID.String(),
			Transition:    transition,
			Mutate:        mutate,
		})
	case jobs.StatusCanceled:
		transition.ResultSummary = &jobs.ResultSummary{
			Code:         finalization.ResultCode,
			Message:      "Import apply canceled.",
			ResourceRefs: finalization.ResourceRefs,
		}
		_, err = s.jobSuccessFinalizer.FinalizeImportJobCancellation(ctx, JobTerminalFinalization{
			Transition: transition,
			Mutate:     mutate,
		})
	case jobs.StatusFailed:
		transition.ErrorSummary = &jobs.ErrorSummary{
			Code:      finalization.ErrorCode,
			Message:   "Import apply failed.",
			Retryable: false,
			Details:   map[string]any{},
		}
		_, err = s.jobSuccessFinalizer.FinalizeImportJobFailure(ctx, JobTerminalFinalization{
			Transition: transition,
			Mutate:     mutate,
		})
	default:
		return fmt.Errorf("unsupported import finalization job status %q", finalization.JobStatus)
	}
	return err
}

func importUnitFailure(err error) (string, string) {
	var applyBlocked *ApplyBlockedError
	switch {
	case errors.As(err, &applyBlocked):
		return "import_apply_blocked", applyBlocked.ReasonCode
	case errors.Is(err, errImportUnitCanceled):
		return "import_apply_canceled", "cancel_requested"
	case errors.Is(err, incidents.ErrIncidentClosed):
		return "incident_closed", "incident_closed"
	case errors.Is(err, incidents.ErrIncidentNotFound),
		errors.Is(err, incidents.ErrMembershipNotFound),
		errors.Is(err, incidents.ErrIncidentRoleDenied),
		errors.Is(err, errImportActorUnauthorized):
		return "authorization_denied", "authorization_changed"
	default:
		return "import_apply_failed", "owner_apply_failed"
	}
}

func (s *Service) jobCancelRequested(ctx context.Context, jobID uuid.UUID) bool {
	job, err := s.jobManager.Get(ctx, jobID)
	return err == nil && job.Status == jobs.StatusCancelRequested
}
