package records

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
)

// REQ-01-705 classifies a record mutation target in a fixed order. Records owns
// the first two steps for every owner: a missing envelope is not found, and a
// soft-deleted envelope must be restored before it can be mutated. The route's
// own type, incident-lifecycle and version steps follow in the owner.
var (
	// ErrTargetNotFound reports that no envelope exists. It also matches
	// enveloperead.ErrEnvelopeNotFound.
	ErrTargetNotFound = fmt.Errorf("records: target record not found: %w", enveloperead.ErrEnvelopeNotFound)
	// ErrTargetDeleted reports a soft-deleted envelope. Owners map it to 409
	// record_deleted_use_restore (REQ-01-211, REQ-01-705).
	ErrTargetDeleted = errors.New("records: target record is deleted; restore it first")
)

// TargetLock selects the envelope row lock taken by the classifying read, so the
// classification holds until the caller's transaction ends.
type TargetLock uint8

const (
	// TargetLockNone reads without a row lock. Callers use it when they already
	// hold the envelope lock or take it later in canonical record order.
	TargetLockNone TargetLock = iota
	// TargetLockShare takes FOR SHARE.
	TargetLockShare
	// TargetLockNoKeyUpdate takes FOR NO KEY UPDATE.
	TargetLockNoKeyUpdate
	// TargetLockUpdate takes FOR UPDATE.
	TargetLockUpdate
)

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// ClassifyTargetTx reads the target envelope in the caller's transaction under
// the requested lock and applies the existence and deletion steps of
// REQ-01-705. It returns ErrTargetNotFound for a missing envelope and
// ErrTargetDeleted for a soft-deleted one. With ErrTargetDeleted the envelope is
// still returned, so a route can decide incident visibility first (REQ-01-704).
// Any other failure is returned wrapped and is never a classification.
func ClassifyTargetTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock TargetLock) (enveloperead.Envelope, error) {
	return classifyTarget(ctx, tx, recordID, lock)
}

// ClassifyTargetTx is the Store form of the package classifier, for owners that
// receive Records through an injected capability.
func (s *Store) ClassifyTargetTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock TargetLock) (enveloperead.Envelope, error) {
	return ClassifyTargetTx(ctx, tx, recordID, lock)
}

// ClassifyTargetPresence applies the same two steps to an envelope the caller
// already read, for example under a canonical multi-record lock.
func ClassifyTargetPresence(found bool, deletedAt *time.Time) error {
	switch {
	case !found:
		return ErrTargetNotFound
	case deletedAt != nil:
		return ErrTargetDeleted
	default:
		return nil
	}
}

func classifyTarget(ctx context.Context, querier rowQuerier, recordID uuid.UUID, lock TargetLock) (enveloperead.Envelope, error) {
	if lock > TargetLockUpdate {
		return enveloperead.Envelope{}, fmt.Errorf("records: unknown target lock %d", lock)
	}
	envelope, err := enveloperead.ScanRow(querier.QueryRow(ctx, targetClassificationSQL(lock), recordID))
	found := err == nil
	if err != nil && !errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		return enveloperead.Envelope{}, fmt.Errorf("classify record target: %w", err)
	}
	if err := ClassifyTargetPresence(found, envelope.DeletedAt); err != nil {
		return envelope, err
	}
	return envelope, nil
}

// targetClassificationSQL is the envelope read for one lock mode.
func targetClassificationSQL(lock TargetLock) string {
	switch lock {
	case TargetLockShare:
		return enveloperead.SelectSQL + "\n WHERE record_id = $1\n FOR SHARE"
	case TargetLockNoKeyUpdate:
		return enveloperead.SelectSQL + "\n WHERE record_id = $1\n FOR NO KEY UPDATE"
	case TargetLockUpdate:
		return enveloperead.SelectSQL + "\n WHERE record_id = $1\n FOR UPDATE"
	default:
		return enveloperead.SelectSQL + "\n WHERE record_id = $1"
	}
}
