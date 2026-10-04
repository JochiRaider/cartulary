package reference_data

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Execute is the only Reference Pack Jobs handler. Jobs owns queueing and the
// execution lease; this coordinator owns frozen inputs, verification, semantic
// publication and the terminal receipt transaction.
func (c *Coordinator) Execute(ctx context.Context, execution jobs.Execution) (resultErr error) {
	monotonicStart := time.Now()
	start := c.now().UTC()
	executionCtx, executionCancel := context.WithDeadline(ctx, monotonicStart.Add(time.Duration(c.limits.ReferencePacks.MaxVerificationSeconds)*time.Second))
	defer executionCancel()
	if _, err := c.operations.ObserveExecution(executionCtx, execution); err != nil {
		return err
	}
	var operationID uuid.UUID
	if err := c.pool.QueryRow(executionCtx, `SELECT operation_id FROM reference_pack_operations WHERE job_id=$1`, execution.JobID()).Scan(&operationID); err != nil {
		return err
	}
	a, err := c.beginAttempt(executionCtx, execution, operationID, start)
	if err != nil {
		return err
	}
	semanticOutcome := "failed"
	if a.Frozen.Kind == "reverify" || a.Frozen.Kind == "refresh" {
		var end func(string)
		executionCtx, end = observeReferenceOperation(executionCtx, c.observer, "reference_pack."+a.Frozen.Kind)
		defer func() {
			if resultErr != nil {
				end(referenceOutcome(resultErr))
			} else {
				end(semanticOutcome)
			}
		}()
	}
	runReferencePackWorkerStartHook(a.Frozen.Kind)
	deadline := extensiondeadline.New(0, a.Frozen.TimeoutSeconds, nil)
	budgetCtx, cancel := context.WithDeadline(executionCtx, monotonicStart.Add(time.Duration(a.Frozen.TimeoutSeconds)*time.Second))
	defer cancel()
	preparationCtx, cancelPreparation := context.WithCancelCause(budgetCtx)
	defer cancelPreparation(nil)
	observer := c.observeDuringPreparation(preparationCtx, execution, monotonicStart, cancelPreparation)
	// Changes to deployment limits between admission and restart cannot change
	// the interpretation of frozen bytes. They require a new explicit operation.
	var configuration string
	if err = c.pool.QueryRow(budgetCtx, `SELECT configuration_sha256 FROM reference_pack_current_set WHERE singleton`).Scan(&configuration); err == nil && configuration != a.Frozen.ConfigurationSHA256 {
		err = &OperationRejection{Reason: "stale_admission_state"}
	}
	var cancellationSample *int64
	observe := func() error {
		job, observeErr := c.operations.ObserveExecution(budgetCtx, execution)
		if observeErr != nil {
			return observeErr
		}
		if job.Status == jobs.StatusCancelRequested {
			sample := time.Since(monotonicStart).Nanoseconds()
			cancellationSample = &sample
			return jobs.ErrCancellationRequested
		}
		return budgetCtx.Err()
	}
	if err == nil {
		err = observe()
	}
	for ordinal := int64(1); err == nil && ordinal <= a.Count; ordinal++ {
		var member frozenMember
		member, err = c.frozenMember(preparationCtx, a, ordinal)
		if err == nil {
			err = c.prepareMember(preparationCtx, a, member)
		}
		if err == nil {
			err = observe()
		}
	}
	observedCancellation, observedError := observer.finish()
	if observedError != nil {
		err = observedError
		cancellationSample = observedCancellation
	}
	if err == nil {
		var failures int64
		err = c.pool.QueryRow(budgetCtx, `SELECT count(*) FROM reference_pack_attempt_members WHERE attempt_id=$1 AND verdict='content_rejected'`, a.ID).Scan(&failures)
		if err == nil {
			mutate := func(ctx context.Context, tx pgx.Tx) error { return c.publishAttemptTx(ctx, tx, a, c.now().UTC()) }
			if failures > 0 {
				var code, check, summaryID string
				var encoded []byte
				err = c.pool.QueryRow(budgetCtx, `SELECT failure_code,check_id,validation_summary_id,canonical_validation_summary FROM reference_pack_attempt_members WHERE attempt_id=$1 AND verdict='content_rejected' ORDER BY ordinal LIMIT 1`, a.ID).Scan(&code, &check, &summaryID, &encoded)
				if err == nil {
					var summary *packformat.ValidationSummary
					summary, err = packformat.DecodeValidationSummary(encoded)
					if err == nil {
						_, err = c.finalizer.FinalizeReferencePackJobFailure(budgetCtx, JobFailureFinalization{Execution: execution, Completion: failedCompletion("reference_pack_verification_failed", map[string]any{"reason_code": code, "check_id": check, "failed_count": failures, "primary_issue_id": summary.PrimaryIssueID, "validation_summary_ref": summaryID, "total_issue_count": summary.Total, "retained_issue_count": summary.Retained, "issues_truncated": summary.Truncated}), Mutate: mutate})
					}
				}
			} else {
				code := map[string]string{"import": ResultReferencePackImported, "reverify": ResultReferencePackReverified, "refresh": ResultReferencePacksRefreshed}[a.Frozen.Kind]
				refs, refErr := c.successfulResourceRefs(budgetCtx, a)
				err = refErr
				if err == nil {
					_, err = c.finalizer.FinalizeReferencePackJobSuccess(budgetCtx, JobSuccessFinalization{Execution: execution, FinalCommitID: ProfileID + ":" + a.OperationID.String(), Completion: jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPtr(1)}, ResultSummary: jobs.ResultSummary{Code: code, Message: code, ResourceRefs: refs}}, Mutate: mutate})
				}
			}
			// The owner finalizer has proved commit, including lost acknowledgements.
			// A later cancellation or deadline cannot change this outcome.
			if err == nil {
				semanticOutcome = "success"
				if failures > 0 {
					semanticOutcome = "rejected"
				}
				return nil
			}
			if errors.Is(err, ErrJobFinalizationIndeterminate) {
				return err
			}
		}
	}
	// A lost execution lease or process shutdown is recovered by Jobs. Leave
	// inputs and unpublished artifacts intact for the next execution attempt.
	if ctx.Err() != nil || errors.Is(err, jobs.ErrExecutionLost) {
		return err
	}
	if errors.Is(err, jobs.ErrCancellationRequested) && cancellationSample == nil {
		sample := time.Since(monotonicStart).Nanoseconds()
		cancellationSample = &sample
	}
	var expirySample *int64
	if deadline.Expired(time.Since(monotonicStart).Nanoseconds()) {
		sample := deadline.MonotonicNS
		expirySample = &sample
	}
	outcome := "execution_failed"
	code := "internal_error"
	details := map[string]any{}
	switch extensiondeadline.Classify(extensiondeadline.CommitProvenAbsent, cancellationSample, expirySample) {
	case extensiondeadline.OutcomeCanceled:
		outcome = "canceled"
	case extensiondeadline.OutcomeTimedOut:
		outcome = "timed_out"
		code = "reference_pack_verification_failed"
		details["reason_code"] = "verification_timeout"
	default:
		var rejected *OperationRejection
		if errors.As(err, &rejected) {
			code = "reference_pack_operation_rejected"
			details["reason_code"] = rejected.Reason
			if rejected.Reason == "stale_admission_state" {
				outcome = "stale_state"
			}
		}
	}
	// Verification is over. This transaction records only the proven abort and
	// clears an initial staged candidate; it cannot publish prepared content.
	completionCtx, finishCancel := context.WithTimeout(ctx, 10*time.Second)
	defer finishCancel()
	mutate := func(ctx context.Context, tx pgx.Tx) error { return abortAttemptTx(ctx, tx, a, c.now().UTC(), outcome) }
	if outcome == "canceled" {
		_, err = c.finalizer.FinalizeReferencePackJobCancellation(completionCtx, JobCancellationFinalization{Execution: execution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: intPtr(1)}}, Mutate: mutate})
	} else if outcome == "timed_out" {
		_, err = c.finalizer.FinalizeReferencePackJobTimeout(completionCtx, JobFailureFinalization{Execution: execution, Completion: failedCompletion(code, details), Mutate: mutate})
	} else {
		_, err = c.finalizer.FinalizeReferencePackJobFailure(completionCtx, JobFailureFinalization{Execution: execution, Completion: failedCompletion(code, details), Mutate: mutate})
	}
	switch outcome {
	case "canceled":
		semanticOutcome = "canceled"
	case "timed_out":
		semanticOutcome = "timeout"
	case "stale_state":
		semanticOutcome = "conflict"
	}
	return err
}

func (c *Coordinator) successfulResourceRefs(ctx context.Context, a executionAttempt) ([]jobs.ResourceRef, error) {
	rows, err := c.pool.Query(ctx, `SELECT pack_key,pack_version FROM reference_pack_attempt_members WHERE attempt_id=$1 AND verdict='succeeded' ORDER BY ordinal LIMIT 1024`, a.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refs := []jobs.ResourceRef{}
	for rows.Next() {
		var key, version string
		if err := rows.Scan(&key, &version); err != nil {
			return nil, err
		}
		route := referencePackRoute(key, version)
		refs = append(refs, jobs.ResourceRef{Kind: "reference_pack_version", ID: route, Route: route})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if a.Frozen.Kind != "refresh" && len(refs) != 1 {
		return nil, errors.New("reference pack: successful operation lacks exact resource")
	}
	return refs, nil
}
