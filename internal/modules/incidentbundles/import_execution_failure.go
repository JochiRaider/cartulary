package incidentbundles

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/jackc/pgx/v5"
)

// The common finalizer has either not begun or proved rollback before this is
// called. Keep abort publication outside the expired verification context.
func (w *incidentBundleWorker) finishImportFailure(ctx context.Context, scope *reference_data.IncidentReferenceExecution, execution jobs.Execution, cause error) {
	if ctx.Err() != nil {
		return
	}
	finish, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// One rejected publication may reveal cancellation or a changed frozen
	// dependency. Reclassify proven absence once, without re-verifying/rebasing.
	for pass := 0; pass < 2; pass++ {
		outcome, code, details := scope.ClassifyAbsent(finish, cause)
		if outcome == "recoverable" {
			return
		}
		reason := outcome
		if code == "" {
			var rejected *verificationError
			if errors.As(cause, &rejected) || errors.Is(cause, ErrPortabilityUnavailable) || errors.Is(cause, ErrPortabilityBlocked) || errors.Is(cause, ErrPortabilityLimit) || errors.Is(cause, ErrPortabilityPayload) || errors.Is(cause, ErrPortabilityResult) {
				code = "incident_bundle_import_rejected"
				reason, details = incidentBundleFailureDetails(code, cause)
			} else {
				code = "internal_error"
				details = map[string]any{}
			}
		} else if value, ok := details["reason_code"].(string); ok {
			reason = value
		}
		mutate := func(ctx context.Context, tx pgx.Tx) error {
			if err := scope.AbortTx(ctx, tx); err != nil {
				return err
			}
			return w.store.markJobFailureTx(ctx, tx, execution.JobID(), reason, w.now())
		}
		completion := failedCompletion(code, details)
		publication, stop := scope.FinalizationContext(finish)
		var err error
		switch outcome {
		case "canceled":
			_, err = w.jobFinalizer.FinalizeIncidentBundleJobCancellation(publication, JobCancellationFinalization{Execution: execution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: intPtr(1)}}, Mutate: mutate})
		case "timed_out":
			_, err = w.jobFinalizer.FinalizeIncidentBundleJobTimeout(publication, JobFailureFinalization{Execution: execution, Completion: completion, Mutate: mutate})
		default:
			_, err = w.jobFinalizer.FinalizeIncidentBundleJobFailure(publication, JobFailureFinalization{Execution: execution, Completion: completion, Mutate: mutate})
		}
		stop()
		if err == nil || errors.Is(err, ErrJobFinalizationIndeterminate) {
			return
		}
		var rejected *reference_data.OperationRejection
		if !errors.Is(err, jobs.ErrCancellationRequested) && !errors.Is(err, context.DeadlineExceeded) && !(errors.As(err, &rejected) && rejected.Reason == "stale_admission_state") {
			return
		}
		cause = err
	}
}
