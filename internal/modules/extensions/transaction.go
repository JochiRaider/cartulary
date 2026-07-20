package extensions

import (
	"context"
	"errors"
	"fmt"
	"math"
)

const (
	TransactionParticipantLimit = 16384
	TransactionByteLimit        = 64 * 1024 * 1024
)

var (
	ErrTransactionParticipants  = errors.New("extension_transaction_participant_limit")
	ErrTransactionInput         = errors.New("extension_transaction_input_limit")
	ErrTransactionPrepare       = errors.New("extension_transaction_prepare_invalid")
	ErrTransactionCancelled     = errors.New("extension_transaction_cancelled")
	ErrTransactionTimeout       = errors.New("extension_transaction_timeout")
	ErrTransactionIndeterminate = errors.New("extension_transaction_commit_indeterminate")
)

type TransactionParticipant interface {
	ID() string
	InputSize() int64
	Prepare(context.Context) (PreparedTransactionResult, error)
	Validate(context.Context, PreparedTransactionResult) error
}

type PreparedTransactionResult struct {
	CanonicalBytes []byte
	Value          any
}

type TransactionBackend interface {
	Begin(context.Context) (SharedTransaction, error)
}

type SharedTransaction interface {
	Write(context.Context, string, PreparedTransactionResult) error
	Commit(context.Context) (TransactionCommitOutcome, error)
	Rollback(context.Context) error
}

// extensionstoreCommitOutcome intentionally mirrors the platform boundary
// without importing a PostgreSQL adapter into generic coordination.
type TransactionCommitOutcome string

const (
	TransactionCommitProven  TransactionCommitOutcome = "committed"
	TransactionCommitAbsent  TransactionCommitOutcome = "absent"
	TransactionCommitUnknown TransactionCommitOutcome = "indeterminate"
)

type TransactionCoordinator struct {
	backend TransactionBackend
}

func NewTransactionCoordinator(backend TransactionBackend) (*TransactionCoordinator, error) {
	if backend == nil {
		return nil, errors.New("extension transaction backend unavailable")
	}
	return &TransactionCoordinator{backend: backend}, nil
}

func ValidateTransactionInputs(participants []TransactionParticipant) error {
	if len(participants) < 1 || len(participants) > TransactionParticipantLimit {
		return ErrTransactionParticipants
	}
	var aggregate int64
	previousID := ""
	for _, participant := range participants {
		if participant == nil || participant.ID() == "" || (previousID != "" && previousID >= participant.ID()) {
			return ErrTransactionParticipants
		}
		previousID = participant.ID()
		size := participant.InputSize()
		if size < 0 || size > TransactionByteLimit || aggregate > math.MaxInt64-size {
			return ErrTransactionInput
		}
		aggregate += size
		if aggregate > TransactionByteLimit {
			return ErrTransactionInput
		}
	}
	return nil
}

func (c *TransactionCoordinator) Execute(ctx context.Context, participants []TransactionParticipant) error {
	if c == nil || c.backend == nil {
		return ErrTransactionPrepare
	}
	if err := ValidateTransactionInputs(participants); err != nil {
		return err
	}
	if err := transactionContextError(ctx); err != nil {
		return err
	}
	prepared := make([]PreparedTransactionResult, len(participants))
	var aggregateResult int64
	for index, participant := range participants {
		if err := transactionContextError(ctx); err != nil {
			return err
		}
		result, err := participant.Prepare(ctx)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrTransactionPrepare, participant.ID())
		}
		if err := transactionContextError(ctx); err != nil {
			return err
		}
		resultSize := int64(len(result.CanonicalBytes))
		if resultSize > TransactionByteLimit || aggregateResult > math.MaxInt64-resultSize {
			return ErrTransactionPrepare
		}
		aggregateResult += resultSize
		if aggregateResult > TransactionByteLimit {
			return ErrTransactionPrepare
		}
		prepared[index] = result
	}
	for index, participant := range participants {
		if err := transactionContextError(ctx); err != nil {
			return err
		}
		if err := participant.Validate(ctx, prepared[index]); err != nil {
			return fmt.Errorf("%w: %s", ErrTransactionPrepare, participant.ID())
		}
		if err := transactionContextError(ctx); err != nil {
			return err
		}
	}
	tx, err := c.backend.Begin(ctx)
	if err != nil {
		return err
	}
	finalized := false
	defer func() {
		if !finalized {
			_ = tx.Rollback(context.WithoutCancel(ctx))
		}
	}()
	for index, participant := range participants {
		if err := transactionContextError(ctx); err != nil {
			return err
		}
		if err := tx.Write(ctx, participant.ID(), prepared[index]); err != nil {
			return err
		}
		if err := transactionContextError(ctx); err != nil {
			return err
		}
	}
	if err := transactionContextError(ctx); err != nil {
		return err
	}
	outcome, commitErr := tx.Commit(ctx)
	finalized = true
	if outcome == TransactionCommitProven {
		return nil
	}
	if outcome == TransactionCommitUnknown {
		return fmt.Errorf("%w: %v", ErrTransactionIndeterminate, commitErr)
	}
	if err := transactionContextError(ctx); err != nil {
		return err
	}
	if commitErr != nil {
		return commitErr
	}
	return ErrTransactionIndeterminate
}

func transactionContextError(ctx context.Context) error {
	if ctx == nil {
		return ErrTransactionCancelled
	}
	err := ctx.Err()
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTransactionTimeout
	}
	if err != nil {
		return ErrTransactionCancelled
	}
	return nil
}
