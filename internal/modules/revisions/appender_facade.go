package revisions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

type HistoricalIntentPolicy interface {
	IsSuppressedTx(context.Context, pgx.Tx) (bool, error)
}

type IntentAppender interface {
	AppendIntentTx(context.Context, pgx.Tx, collaboration.EventIntent) error
}

// Appender is the composition-scoped Revisions write facade. Callers retain
// transaction ownership and can acquire no history-query or command
// capabilities through it.
type Appender struct {
	recordViews      *RecordViewCatalog
	recordEnvelopes  RecordEnvelopeTxReader
	snapshotCaptures *RecordSnapshotCaptureCatalog
	targetSemantics  *TargetSemanticsCatalog
	historicalPolicy HistoricalIntentPolicy
	intents          IntentAppender
}

func NewAppender(
	recordViews *RecordViewCatalog,
	recordEnvelopes RecordEnvelopeTxReader,
	snapshotCaptures *RecordSnapshotCaptureCatalog,
	targetSemantics *TargetSemanticsCatalog,
	historicalPolicy HistoricalIntentPolicy,
	intents IntentAppender,
) (*Appender, error) {
	if recordViews == nil {
		return nil, errors.New("revisions: record/view catalog is required")
	}
	if recordEnvelopes == nil {
		return nil, errors.New("revisions: record envelope reader is required")
	}
	if snapshotCaptures == nil {
		return nil, errors.New("revisions: snapshot capture catalog is required")
	}
	if targetSemantics == nil {
		return nil, errors.New("revisions: target semantics catalog is required")
	}
	if historicalPolicy == nil {
		return nil, errors.New("revisions: historical intent policy is required")
	}
	if intents == nil {
		return nil, errors.New("revisions: Collaboration intent appender is required")
	}
	return &Appender{
		recordViews:      recordViews,
		recordEnvelopes:  recordEnvelopes,
		snapshotCaptures: snapshotCaptures,
		targetSemantics:  targetSemantics,
		historicalPolicy: historicalPolicy,
		intents:          intents,
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

type LiveRecordChange struct {
	BeforeValue any
	AfterValue  any
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

type AppendRecordRevisionParams struct {
	ChangeSetID    uuid.UUID
	RecordID       uuid.UUID
	RowVersion     int64
	BeforeSnapshot *RecordSnapshot
	AfterSnapshot  *RecordSnapshot
	LiveChange     LiveRecordChange
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
