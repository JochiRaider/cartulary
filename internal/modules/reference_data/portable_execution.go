package reference_data

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IncidentReferenceExecution carries one process-local budget through the parent
// import transaction. Only fresh container work activates its verification budget.
// Its callback boundary owns observation and resource completion.
type incidentExecutionKey struct{}

type IncidentReferenceExecution struct {
	active            atomic.Bool
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

// RunImportExecution owns the complete parent operation and joins all cleanup.
// The callback must classify proven absence before publishing ordinary failure.
func (r *incidentReferences) RunImportExecution(ctx context.Context, execution jobs.Execution, started time.Time, callback func(context.Context, *IncidentReferenceExecution) error) error {
	if execution.JobID() == uuid.Nil || started.IsZero() || callback == nil {
		return errors.New("reference pack: parent execution required")
	}
	run, cancel := context.WithCancelCause(r.verifier.lifetime.track(ctx))
	e := &IncidentReferenceExecution{owner: r, execution: execution, started: started, cancel: cancel}
	e.ctx = context.WithValue(run, incidentExecutionKey{}, e)
	return e.run(callback)
}

func (e *IncidentReferenceExecution) run(callback func(context.Context, *IncidentReferenceExecution) error) error {
	e.active.Store(true)
	defer e.active.Store(false)
	// Cancel only after Extensions has classified completion. Normal teardown must
	// not manufacture an interruption (especially when configured grace is zero).
	defer e.cancel(nil)
	return e.owner.verifier.lifetime.Run(e.ctx, func(ctx context.Context) error {
		defer func() {
			if e.observation != nil {
				e.observation.finish()
			}
			if e.timer != nil {
				e.timer.Stop()
			}
		}()
		return callback(ctx, e)
	})
}

// CompleteCleanup detaches semantic cancellation only inside joined completion.
// The scope's original context still enforces the same first grace deadline.
func (e *IncidentReferenceExecution) CompleteCleanup(cleanup func(context.Context) error) error {
	if e == nil || !e.active.Load() || cleanup == nil {
		return errors.New("reference pack: inactive portable completion")
	}
	err := e.owner.verifier.lifetime.completeCleanup(e.ctx, cleanup)
	if err != nil {
		// Physical teardown failure invalidates any unpublished content verdict.
		// A caller must classify the operational failure before an ordinary abort.
		e.requiredRejection = nil
		e.failure = ""
	}
	return err
}

// CompleteFinalization keeps ordinary terminal writes inside the original
// grace. Exhaustion revokes their context before the fatal process signal;
// physical teardown uses CompleteCleanup instead.
func (e *IncidentReferenceExecution) CompleteFinalization(finalize func(context.Context) error) error {
	if e == nil || !e.active.Load() || finalize == nil {
		return errors.New("reference pack: inactive portable finalization")
	}
	completion, err := extensiondeadline.CompleteWithinGrace(e.ctx, e.owner.verifier.lifetime.policy, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, finalize(ctx)
	})
	if err != nil {
		return err
	}
	return completion.Err
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
func (e *IncidentReferenceExecution) ClassifyAbsent(ctx context.Context, err error) (outcome string, code string, details map[string]any, resultErr error) {
	e.failure = ""
	if fatal := e.fatalError(err); fatal != nil {
		return "", "", nil, fatal
	}
	if ctx.Err() != nil || errors.Is(err, jobs.ErrExecutionLost) {
		return "recoverable", "", nil, nil
	}
	var cancellation *int64
	if e.observation != nil {
		sample, observed := e.observation.finish()
		if fatal := e.fatalError(observed); fatal != nil {
			return "", "", nil, fatal
		}
		cancellation = sample
		if errors.Is(observed, jobs.ErrExecutionLost) {
			return "recoverable", "", nil, nil
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
		if fatal := e.fatalError(observeErr); fatal != nil {
			return "", "", nil, fatal
		}
		if errors.Is(observeErr, jobs.ErrExecutionLost) {
			return "recoverable", "", nil, nil
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
	if fatal := e.fatalError(err); fatal != nil {
		return "", "", nil, fatal
	}
	e.failure = outcome
	return outcome, code, details, nil
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
	if fatal := e.fatalError(nil); fatal != nil {
		return fatal
	}
	if !e.active.Load() || e.failure == "" {
		return errors.New("reference pack: unclassified or inactive portable abort")
	}
	if e.operation == uuid.Nil {
		return nil
	}
	var terminal bool
	err := tx.QueryRow(ctx, `SELECT terminal_at IS NOT NULL FROM reference_pack_operations WHERE operation_id=$1 AND job_id=$2 FOR UPDATE`, e.operation, e.execution.JobID()).Scan(&terminal)
	if fatal := e.fatalError(err); fatal != nil {
		return fatal
	}
	if err != nil {
		return err
	}
	if terminal {
		return nil
	}
	if e.failure == "content_rejected" {
		err = e.owner.publishRequiredPortableRejectionTx(ctx, tx, e.requiredRejection, e.owner.verifier.now().UTC())
	} else {
		err = abortReferenceOperationTx(ctx, tx, e.operation, "portable_retention", e.failure, e.owner.verifier.now().UTC())
	}
	if fatal := e.fatalError(err); fatal != nil {
		return fatal
	}
	return err
}

func (e *IncidentReferenceExecution) fatalError(cause error) error {
	if fatal := verificationFatalError(cause); fatal != nil {
		return fatal
	}
	if e.ctx != nil && extensiondeadline.GraceExhausted(e.ctx) {
		return extensiondeadline.ErrGraceExhausted
	}
	return nil
}
