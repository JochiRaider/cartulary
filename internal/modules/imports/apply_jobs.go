package imports

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func (s *Service) executeApplyJob(ctx context.Context, execution jobs.Execution) error {
	var payload applyJobHandlerPayload
	if err := s.decodeJobPayload(ctx, execution, &payload); err != nil {
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
	job, err := s.jobManager.ObserveExecution(ctx, execution)
	if err != nil {
		return err
	}
	return s.completeApplyJob(ctx, execution, actor, ApplyStartResult{
		Job:             job,
		IncidentID:      incidentID,
		ImportSessionID: sessionID,
		ClientTxnID:     payload.ClientTxnID,
		SelectedUnitIDs: selected,
	})
}

func (s *Service) completeApplyJob(ctx context.Context, execution jobs.Execution, actor authn.UserRecord, start ApplyStartResult) error {
	total := len(start.SelectedUnitIDs)
	job, getErr := s.jobManager.ObserveExecution(ctx, execution)
	resumingCancellation := getErr == nil && job.Status == jobs.StatusCancelRequested
	if !resumingCancellation && !s.prepareClaimedJob(ctx, execution, total) {
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
		if s.jobCancelRequested(ctx, execution) {
			if _, err := s.store.recordTerminalUnitOutcome(
				ctx,
				execution,
				start,
				unit.UnitID,
				actor.ID,
				"canceled",
				importUnitFailureDetail{
					ErrorCode:  "import_apply_canceled",
					ReasonCode: "cancel_requested",
					Retryable:  false,
					Details:    map[string]any{"reason_code": "cancel_requested"},
				},
				s.now(),
			); err != nil {
				return err
			}
			continue
		}
		if _, err := s.applyUnit(ctx, execution, actor, start, unit.UnitID); err == nil {
			continue
		} else if errors.Is(err, errUnitCommitIndeterminate) {
			return err
		} else {
			status := "failed"
			failure := importUnitFailure(err)
			if errors.Is(err, errImportUnitCanceled) {
				status = "canceled"
			}
			if _, persistErr := s.store.recordTerminalUnitOutcome(
				ctx,
				execution,
				start,
				unit.UnitID,
				actor.ID,
				status,
				failure,
				s.now(),
			); persistErr != nil {
				return persistErr
			}
		}
	}
	return s.finalizeApplyJob(ctx, execution, start)
}

func (s *Service) finalizeApplyJob(ctx context.Context, execution jobs.Execution, start ApplyStartResult) error {
	finalization, err := s.store.prepareApplyFinalization(ctx, start)
	if err != nil {
		return err
	}
	jobID := execution.JobID()
	total := len(start.SelectedUnitIDs)
	progress := jobs.Progress{Completed: total, Total: &total}
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
		completion := jobs.SuccessCompletion{Progress: progress, ResultSummary: jobs.ResultSummary{
			Code:         finalization.ResultCode,
			Message:      "Import session applied.",
			ResourceRefs: finalization.ResourceRefs,
		}}
		_, err = s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
			FinalCommitID: "import.apply:" + jobID.String(),
			Execution:     execution,
			Completion:    completion,
			Mutate:        mutate,
		})
	case jobs.StatusCanceled:
		completion := jobs.CancellationCompletion{Progress: progress, ResultSummary: jobs.ResultSummary{
			Code:         finalization.ResultCode,
			Message:      "Import apply canceled.",
			ResourceRefs: finalization.ResourceRefs,
		}}
		_, err = s.jobSuccessFinalizer.FinalizeImportJobCancellation(ctx, JobCancellationFinalization{
			Execution: execution, Completion: completion, Mutate: mutate,
		})
	case jobs.StatusFailed:
		errorDetails := cloneStringAnyMap(finalization.ErrorDetails)
		if errorDetails == nil {
			errorDetails = map[string]any{}
		}
		completion := jobs.FailureCompletion{Progress: progress, ErrorSummary: jobs.ErrorSummary{
			Code:      finalization.ErrorCode,
			Message:   "Import apply failed.",
			Retryable: finalization.ErrorRetryable,
			Details:   errorDetails,
		}}
		_, err = s.jobSuccessFinalizer.FinalizeImportJobFailure(ctx, JobFailureFinalization{
			Execution: execution, Completion: completion, Mutate: mutate,
		})
	default:
		return fmt.Errorf("unsupported import finalization job status %q", finalization.JobStatus)
	}
	return err
}

func importUnitFailure(err error) importUnitFailureDetail {
	var translated *translatedImportUnitError
	var applyBlocked *ApplyBlockedError
	switch {
	case errors.As(err, &translated):
		return translated.failure
	case errors.As(err, &applyBlocked):
		if failure, ok := commonImportApplyFailure(err); ok {
			return failure
		}
		return genericOwnerApplyFailure()
	case errors.Is(err, errImportUnitCanceled):
		return importUnitFailureDetail{
			ErrorCode:  "import_apply_canceled",
			ReasonCode: "cancel_requested",
			Retryable:  false,
			Details:    map[string]any{"reason_code": "cancel_requested"},
		}
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return importUnitFailureDetail{
			ErrorCode:  "incident_closed",
			ReasonCode: "incident_closed",
			Retryable:  false,
			Details:    map[string]any{"reason_code": "incident_closed"},
		}
	case admission.IsDenied(err, admission.DenialNotVisible),
		admission.IsDenied(err, admission.DenialInsufficientRole),
		errors.Is(err, errImportActorUnauthorized):
		return importUnitFailureDetail{
			ErrorCode:  "authorization_denied",
			ReasonCode: "authorization_changed",
			Retryable:  false,
			Details:    map[string]any{"reason_code": "authorization_changed"},
		}
	default:
		return genericOwnerApplyFailure()
	}
}

func (s *Service) jobCancelRequested(ctx context.Context, execution jobs.Execution) bool {
	job, err := s.jobManager.ObserveExecution(ctx, execution)
	return err == nil && job.Status == jobs.StatusCancelRequested
}
