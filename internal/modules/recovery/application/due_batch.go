package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

type dueVerificationAttempt func(
	context.Context,
	recovery.BackupSet,
	uuid.UUID,
) (Result, error, bool)

func executeDueVerificationBatch(
	ctx context.Context,
	due []recovery.BackupSet,
	attemptTimeout time.Duration,
	newAttemptID func() uuid.UUID,
	run dueVerificationAttempt,
) (Result, error) {
	if len(due) == 0 {
		return Result{ArtifactRefs: []ArtifactRef{}, Status: ResultNoOp}, nil
	}
	if attemptTimeout <= 0 || newAttemptID == nil || run == nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, errors.New("due verification batch configuration is invalid"))
	}

	outcome := ResultForStoredBackupSet(due[0])
	var firstErr error
	for _, backupSet := range due {
		if err := ctx.Err(); err != nil {
			if firstErr != nil {
				return outcome, firstErr
			}
			return outcome, EnsureFailure(FailureVerificationInvariantCheck, err)
		}
		attemptID := newAttemptID()
		if attemptID == uuid.Nil {
			return outcome, NewFailure(FailureVerificationInvariantCheck, errors.New("due verification attempt ID is empty"))
		}

		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		attemptOutcome, attemptErr, stop := run(attemptCtx, backupSet, attemptID)
		contextErr := attemptCtx.Err()
		cancel()

		outcome.ArtifactRefs = append(outcome.ArtifactRefs, attemptOutcome.ArtifactRefs...)
		if contextErr != nil {
			attemptErr = dueAttemptContextFailure(attemptCtx, attemptErr)
			stop = true
		}
		if attemptErr != nil && firstErr == nil {
			firstErr = attemptErr
		}
		if stop {
			if firstErr == nil {
				firstErr = NewFailure(FailureVerificationInvariantCheck, errors.New("due verification stopped without a recorded failure"))
			}
			return outcome, firstErr
		}
	}
	return outcome, firstErr
}

func dueAttemptContextFailure(ctx context.Context, err error) error {
	if ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return NewFailure(FailureTimeoutElapsed, context.DeadlineExceeded)
	}
	if err == nil && ctx != nil {
		err = ctx.Err()
	}
	if err == nil {
		return nil
	}
	return EnsureFailure(FailureVerificationInvariantCheck, err)
}

func dueAttemptMustStop(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if kind, ok := FailureKindOf(err); ok {
		switch kind {
		case FailureTimeoutElapsed,
			FailureTargetServingTraffic,
			FailureTargetDatabaseNotFresh,
			FailureTargetObjectNamespaceNotFresh,
			FailureTargetMarkerMissing,
			FailureTargetMarkerInvalid,
			FailureOperationLockUnavailable,
			FailureVerificationJournalWrite:
			return true
		}
	}
	var stageErr *recovery.RestoreStageError
	if errors.As(err, &stageErr) {
		switch stageErr.Stage {
		case recovery.RestoreStepPostgresRestore, recovery.RestoreStepObjectStoreRestore:
			return true
		}
	}
	return false
}
