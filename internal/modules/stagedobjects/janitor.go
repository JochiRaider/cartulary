package stagedobjects

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

type JanitorOptions struct {
	Repository       Repository
	Bytes            ByteStore
	Health           *Health
	Now              func() time.Time
	MonotonicNowNS   func() int64
	BatchLimit       int
	OperationTimeout time.Duration
	FatalSink        func(error)
	ErrorSink        func(error)
}

type Janitor struct {
	repository       Repository
	bytes            ByteStore
	health           *Health
	now              func() time.Time
	monotonicNowNS   func() int64
	batchLimit       int
	operationTimeout time.Duration
	fatalSink        func(error)
	errorSink        func(error)

	mu      sync.Mutex
	running bool
	pending bool
}

func NewJanitor(options JanitorOptions) (*Janitor, error) {
	if options.Repository == nil ||
		options.Bytes == nil ||
		options.BatchLimit < 1 ||
		options.OperationTimeout <= 0 ||
		options.FatalSink == nil {
		return nil, errors.New("staged-object janitor dependencies are incomplete")
	}
	if options.Health == nil {
		options.Health = NewHealth()
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	if options.MonotonicNowNS == nil {
		started := time.Now()
		options.MonotonicNowNS = func() int64 { return time.Since(started).Nanoseconds() }
	}
	return &Janitor{
		repository:       options.Repository,
		bytes:            options.Bytes,
		health:           options.Health,
		now:              options.Now,
		monotonicNowNS:   options.MonotonicNowNS,
		batchLimit:       options.BatchLimit,
		operationTimeout: options.OperationTimeout,
		fatalSink:        options.FatalSink,
		errorSink:        options.ErrorSink,
	}, nil
}

// Sweep serializes cleanup and coalesces any number of overlapping triggers
// into exactly one follow-up drain. Each drain captures one cutoff and repeats
// bounded batches until no candidate at or before that cutoff remains.
func (j *Janitor) Sweep(ctx context.Context) error {
	j.mu.Lock()
	if j.running {
		j.pending = true
		j.mu.Unlock()
		return nil
	}
	j.running = true
	j.mu.Unlock()

	var firstErr error
	defer func() {
		j.mu.Lock()
		j.running = false
		j.pending = false
		j.mu.Unlock()
	}()
	followedUp := false
	for {
		err := j.drainCutoff(ctx, j.now().UTC().Truncate(time.Microsecond))
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			j.health.Unavailable(err)
		} else {
			j.health.Available()
		}
		j.mu.Lock()
		followUp := j.pending
		j.pending = false
		j.mu.Unlock()
		if !followUp || followedUp || IsFatalIntegrity(err) {
			return firstErr
		}
		followedUp = true
	}
}

func (j *Janitor) drainCutoff(ctx context.Context, cutoff time.Time) error {
	policy := extensiondeadline.New(j.monotonicNowNS(), durationSeconds(j.operationTimeout), nil)
	var dependencyErr error
	for {
		operationCtx, cancel, err := j.operationContext(ctx, policy)
		if err != nil {
			return err
		}
		candidates, err := j.repository.PrepareCleanupBatch(
			operationCtx,
			cutoff,
			j.now().UTC().Truncate(time.Microsecond),
			j.batchLimit,
		)
		cancel()
		if err != nil {
			return j.classifyRepositoryFailure("prepare_cleanup", err)
		}
		if len(candidates) == 0 {
			return dependencyErr
		}
		for _, candidate := range candidates {
			if candidate.StagingID == "" || candidate.StorageIdentity == "" || candidate.DeleteAttemptCount < 0 {
				return j.fatal(fmt.Errorf("%w: malformed cleanup candidate", ErrIntegrity))
			}
			deleteCtx, cancelDelete, boundaryErr := j.operationContext(ctx, policy)
			if boundaryErr != nil {
				return boundaryErr
			}
			outcome, deleteErr := j.bytes.Delete(deleteCtx, candidate.StorageIdentity)
			cancelDelete()
			switch outcome {
			case DeleteSuccess, DeleteAbsent:
				if err := j.repository.RecordDeletionSuccess(context.WithoutCancel(ctx), candidate.StagingID); err != nil {
					return j.classifyRepositoryFailure("record_deletion_success", err)
				}
			case DeleteRetryableUnknown, DeleteDependency:
				safeCode := "object_delete_retryable"
				if outcome == DeleteDependency {
					safeCode = "object_store_unavailable"
					dependencyErr = NewFailure(FailureDependency, safeCode, deleteErr)
				}
				attemptCount := SaturatingAttempt(candidate.DeleteAttemptCount)
				nextAttemptAt := j.now().UTC().Truncate(time.Microsecond).Add(RetryDelay(attemptCount))
				if err := j.repository.RecordDeletionFailure(
					context.WithoutCancel(ctx),
					DeletionFailure{
						StagingID:     candidate.StagingID,
						AttemptCount:  attemptCount,
						SafeErrorCode: safeCode,
						NextAttemptAt: nextAttemptAt,
					},
				); err != nil {
					return j.classifyRepositoryFailure("record_deletion_failure", err)
				}
			case DeleteIntegrity:
				return j.fatal(fmt.Errorf("%w: delete %s: %v", ErrIntegrity, candidate.StagingID, deleteErr))
			default:
				return j.fatal(fmt.Errorf("%w: unknown delete outcome %q", ErrIntegrity, outcome))
			}
		}
	}
}

func (j *Janitor) Run(ctx context.Context, interval time.Duration) error {
	if j == nil || interval <= 0 {
		return errors.New("staged-object sweep interval is invalid")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := j.runScheduledSweep(ctx); err != nil {
				return err
			}
		}
	}
}

func (j *Janitor) runScheduledSweep(ctx context.Context) error {
	err := j.Sweep(ctx)
	if err == nil {
		return nil
	}
	if j.errorSink != nil {
		j.errorSink(err)
	}
	if IsFatalIntegrity(err) {
		return err
	}
	return nil
}

func (j *Janitor) classifyRepositoryFailure(action string, err error) error {
	if FailureKind(err) == FailureIntegrity {
		return j.fatal(fmt.Errorf("%w: %s: %v", ErrIntegrity, action, err))
	}
	if FailureKind(err) == FailureDependency {
		return NewFailure(FailureDependency, "postgres_unavailable", err)
	}
	return NewFailure(FailureRetryable, action+"_failed", err)
}

func (j *Janitor) fatal(err error) error {
	j.fatalSink(err)
	return &FatalIntegrityError{Cause: err}
}

type FatalIntegrityError struct {
	Cause error
}

func (e *FatalIntegrityError) Error() string {
	return fmt.Sprintf("fatal staged-object integrity failure: %v", e.Cause)
}

func (e *FatalIntegrityError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *FatalIntegrityError) FatalReasonCode() string {
	return "staged_object_publication_mismatch"
}

func IsFatalIntegrity(err error) bool {
	var fatal *FatalIntegrityError
	return errors.As(err, &fatal)
}

func (j *Janitor) operationContext(ctx context.Context, policy extensiondeadline.Deadline) (context.Context, context.CancelFunc, error) {
	now := j.monotonicNowNS()
	if policy.Expired(now) {
		return nil, func() {}, ErrCleanupTimeout
	}
	remaining := policy.MonotonicNS - now
	operationCtx, cancel := context.WithTimeout(ctx, time.Duration(remaining))
	return operationCtx, cancel, nil
}

func durationSeconds(value time.Duration) int64 {
	seconds := int64(value / time.Second)
	if value%time.Second != 0 {
		seconds++
	}
	return seconds
}

func SaturatingAttempt(current int32) int32 {
	if current < 0 {
		return 0
	}
	if current == 1<<31-1 {
		return current
	}
	return current + 1
}

func RetryDelay(attempt int32) time.Duration {
	if attempt <= 1 {
		return time.Minute
	}
	if attempt >= 12 {
		return 24 * time.Hour
	}
	seconds := int64(60) << (attempt - 1)
	if seconds > 86400 {
		seconds = 86400
	}
	return time.Duration(seconds) * time.Second
}
