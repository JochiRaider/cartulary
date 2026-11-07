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
func (w *incidentBundleWorker) finishImportFailure(ctx context.Context, scope *reference_data.IncidentReferenceExecution, execution jobs.Execution, cause error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Ordinary cancellation permits bounded abort publication, but its local
	// timeout must retain the enclosing execution's tracker and fatal selection.
	return scope.CompleteFinalization(func(completion context.Context) error {
		return w.finalizeImportFailure(completion, scope, execution, cause)
	})
}

func (w *incidentBundleWorker) finalizeImportFailure(ctx context.Context, scope *reference_data.IncidentReferenceExecution, execution jobs.Execution, cause error) error {
	finish, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// One rejected publication may reveal cancellation or a changed frozen
	// dependency. Reclassify proven absence once, without re-verifying/rebasing.
	for pass := 0; pass < 2; pass++ {
		outcome, code, details, classificationErr := scope.ClassifyAbsent(finish, cause)
		if classificationErr != nil || outcome == "recoverable" {
			return classificationErr
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
			return err
		}
		var rejected *reference_data.OperationRejection
		if !errors.Is(err, jobs.ErrCancellationRequested) && !errors.Is(err, context.DeadlineExceeded) && !(errors.As(err, &rejected) && rejected.Reason == "stale_admission_state") {
			return err
		}
		cause = err
	}
	return cause
}

func fatalImportFailure(err error) bool {
	var fatal interface{ FatalReasonCode() string }
	return errors.As(err, &fatal) && fatal.FatalReasonCode() != ""
}

func importCleanupError(primary, cleanup error) error {
	if fatalImportFailure(primary) {
		return primary
	}
	if cleanup != nil {
		return cleanup
	}
	return primary
}
