package reference_data

import (
	"context"
	"errors"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IncidentReferenceExecution carries one process-local budget through the parent
// import transaction. Only fresh container work activates it. Close releases
// observation resources; it never changes durable state or classifies commit.
type IncidentReferenceExecution struct {
	owner             *incidentReferences
	execution         jobs.Execution
	started           time.Time
	ctx               context.Context
	cancel            context.CancelCauseFunc
	timer             *time.Timer
	observation       *executionObservation
	seconds           int64
	operation         uuid.UUID
	failure           string
	requiredRejection *PreparedReferenceImport
}

func (r *incidentReferences) BeginImportExecution(ctx context.Context, execution jobs.Execution, started time.Time) (*IncidentReferenceExecution, error) {
	if execution.JobID() == uuid.Nil || started.IsZero() {
		return nil, errors.New("reference pack: parent execution required")
	}
	run, cancel := context.WithCancelCause(ctx)
	return &IncidentReferenceExecution{owner: r, execution: execution, started: started, ctx: run, cancel: cancel}, nil
}
func (e *IncidentReferenceExecution) Context() context.Context { return e.ctx }
func (e *IncidentReferenceExecution) Close() {
	if e.timer != nil {
		e.timer.Stop()
	}
	e.cancel(nil)
	if e.observation != nil {
		e.observation.finish()
	}
}
func (e *IncidentReferenceExecution) activate(seconds int64) error {
	if seconds < 60 || seconds > 86400 {
		return errors.New("reference pack: invalid portable execution budget")
	}
	if e.seconds != 0 {
		if e.seconds != seconds {
			return &OperationRejection{Reason: "stale_admission_state"}
		}
		return e.ctx.Err()
	}
	e.seconds = seconds
	remaining := time.Until(e.started.Add(time.Duration(seconds) * time.Second))
	if remaining <= 0 {
		e.cancel(context.DeadlineExceeded)
		return context.DeadlineExceeded
	}
	e.timer = time.AfterFunc(remaining, func() { e.cancel(context.DeadlineExceeded) })
	e.observation = e.owner.verifier.observeDuringPreparation(e.ctx, e.execution, e.started, e.cancel)
	return e.ctx.Err()
}

// ClassifyAbsent is called only after the parent finalizer proves no commit.
// Proven success and indeterminate commit bypass it. A stopped process or lost
// lease remains recoverable and does not authorize terminal mutation.
func (e *IncidentReferenceExecution) ClassifyAbsent(ctx context.Context, err error) (outcome string, code string, details map[string]any) {
	if ctx.Err() != nil || errors.Is(err, jobs.ErrExecutionLost) {
		return "recoverable", "", nil
	}
	var cancellation *int64
	if e.observation != nil {
		sample, observed := e.observation.finish()
		cancellation = sample
		if errors.Is(observed, jobs.ErrExecutionLost) {
			return "recoverable", "", nil
		}
		if observed != nil {
			err = observed
		}
	}
	if errors.Is(err, jobs.ErrCancellationRequested) && cancellation == nil {
		sample := time.Since(e.started).Nanoseconds()
		cancellation = &sample
	}
	if cancellation == nil {
		job, observeErr := e.owner.verifier.operations.ObserveExecution(ctx, e.execution)
		if errors.Is(observeErr, jobs.ErrExecutionLost) {
			return "recoverable", "", nil
		}
		if observeErr == nil && job.Status == jobs.StatusCancelRequested {
			sample := time.Since(e.started).Nanoseconds()
			cancellation = &sample
		}
	}
	var expiry *int64
	if e.seconds != 0 {
		deadline := extensiondeadline.New(0, e.seconds, nil)
		if deadline.Expired(time.Since(e.started).Nanoseconds()) {
			expiry = &deadline.MonotonicNS
		}
	}
	outcome = "execution_failed"
	if e.requiredRejection != nil {
		outcome = "content_rejected"
	}
	switch extensiondeadline.Classify(extensiondeadline.CommitProvenAbsent, cancellation, expiry) {
	case extensiondeadline.OutcomeCanceled:
		outcome = "canceled"
	case extensiondeadline.OutcomeTimedOut:
		outcome = "timed_out"
		code = "reference_pack_verification_failed"
		details = map[string]any{"reason_code": "verification_timeout"}
	default:
		var rejected *OperationRejection
		if errors.As(err, &rejected) {
			code = "reference_pack_operation_rejected"
			details = map[string]any{"reason_code": rejected.Reason}
			if rejected.Reason == "stale_admission_state" {
				outcome = "stale_state"
			}
		}
	}
	e.failure = outcome
	return outcome, code, details
}

// FinalizationContext keeps content-verdict publication within the original
// monotonic budget. Operational outcomes use the parent's bounded completion
// context after the verification budget has expired.
func (e *IncidentReferenceExecution) FinalizationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if e.failure == "content_rejected" {
		return context.WithDeadline(ctx, e.started.Add(time.Duration(e.seconds)*time.Second))
	}
	return context.WithCancel(ctx)
}

// AbortTx shares the parent's terminal transaction. Generic Jobs termination
// remains the crash/exhaustion fallback when no execution classification exists.
func (e *IncidentReferenceExecution) AbortTx(ctx context.Context, tx pgx.Tx) error {
	if e.operation == uuid.Nil {
		return nil
	}
	if e.failure == "" {
		return errors.New("reference pack: unclassified portable abort")
	}
	var terminal bool
	if err := tx.QueryRow(ctx, `SELECT terminal_at IS NOT NULL FROM reference_pack_operations WHERE operation_id=$1 AND job_id=$2 FOR UPDATE`, e.operation, e.execution.JobID()).Scan(&terminal); err != nil {
		return err
	}
	if terminal {
		return nil
	}
	if e.failure == "content_rejected" {
		return e.owner.publishRequiredPortableRejectionTx(ctx, tx, e.requiredRejection, e.owner.verifier.now().UTC())
	}
	return abortReferenceOperationTx(ctx, tx, e.operation, "portable_retention", e.failure, e.owner.verifier.now().UTC())
}
