package revisions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Appender is the composition-scoped Revisions write facade. Callers retain
// transaction ownership and can acquire no history-query or command
// capabilities through it.
type Appender struct {
	recordEnvelopes  RecordEnvelopeTxReader
	snapshotCaptures *RecordSnapshotCaptureCatalog
	targetSemantics  *TargetSemanticsCatalog
}

func NewAppender(
	recordEnvelopes RecordEnvelopeTxReader,
	snapshotCaptures *RecordSnapshotCaptureCatalog,
	targetSemantics *TargetSemanticsCatalog,
) (*Appender, error) {
	if recordEnvelopes == nil {
		return nil, errors.New("revisions: record envelope reader is required")
	}
	if snapshotCaptures == nil {
		return nil, errors.New("revisions: snapshot capture catalog is required")
	}
	if targetSemantics == nil {
		return nil, errors.New("revisions: target semantics catalog is required")
	}
	return &Appender{
		recordEnvelopes:  recordEnvelopes,
		snapshotCaptures: snapshotCaptures,
		targetSemantics:  targetSemantics,
	}, nil
}

type AppendChangeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

// AppendNonRowMutationParams is the explicit persistence surface for target
// kinds whose owner semantics are not represented by a canonical record
// snapshot. Record history must use the record-snapshot APIs below.
type AppendNonRowMutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type AppendRecordMutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	RecordID        uuid.UUID
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeSnapshot  *RecordSnapshot
	AfterSnapshot   *RecordSnapshot
}

type RevisionConflictFact struct {
	FieldKey      string
	BeforePresent bool
	BeforeValue   any
	AfterPresent  bool
	AfterValue    any
}

type LiveRevisionInput struct {
	ChangeSetID    uuid.UUID
	RecordID       uuid.UUID
	RowVersion     int64
	BeforeSnapshot *RecordSnapshot
	AfterSnapshot  *RecordSnapshot
	ConflictFacts  []RevisionConflictFact
}

type HistoricalRevisionInput struct {
	ChangeSetID    uuid.UUID
	RecordID       uuid.UUID
	RowVersion     int64
	BeforeSnapshot *RecordSnapshot
	AfterSnapshot  *RecordSnapshot
}

func (a *Appender) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (RecordSnapshot, error) {
	envelope, err := a.recordEnvelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
		return RecordSnapshot{}, fmt.Errorf("load record snapshot envelope: %w", err)
	}
	return a.snapshotCaptures.captureTx(ctx, tx, envelope)
}

func recordSnapshotPair(recordID uuid.UUID, before *RecordSnapshot, after *RecordSnapshot) (map[string]any, map[string]any, error) {
	if before == nil && after == nil {
		return nil, nil, fmt.Errorf("%w: record %s has no before or after value", ErrInvalidRecordSnapshot, recordID)
	}
	beforeValue, err := recordSnapshotValue(before, recordID)
	if err != nil {
		return nil, nil, err
	}
	afterValue, err := recordSnapshotValue(after, recordID)
	if err != nil {
		return nil, nil, err
	}
	if before != nil && after != nil && (before.recordType != after.recordType || before.snapshotSchemaID != after.snapshotSchemaID) {
		return nil, nil, fmt.Errorf("%w: record %s before/after schema mismatch", ErrInvalidRecordSnapshot, recordID)
	}
	return beforeValue, afterValue, nil
}
