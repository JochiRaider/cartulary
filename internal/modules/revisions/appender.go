package revisions

import (
	"context"
	"encoding/json"
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
// snapshot. Record history must use the captured-record APIs below.
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

type AppendCapturedRecordMutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	RecordID        uuid.UUID
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeSnapshot  *CapturedRecordSnapshot
	AfterSnapshot   *CapturedRecordSnapshot
}

type AppendCapturedRecordRevisionParams struct {
	ChangeSetID    uuid.UUID
	RecordID       uuid.UUID
	RowVersion     int64
	BeforeSnapshot *CapturedRecordSnapshot
	AfterSnapshot  *CapturedRecordSnapshot
	LiveChange     LiveRecordChange
}

func (a *Appender) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (CapturedRecordSnapshot, error) {
	envelope, err := a.recordEnvelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
		return CapturedRecordSnapshot{}, fmt.Errorf("load record snapshot envelope: %w", err)
	}
	return a.snapshotCaptures.captureTx(ctx, tx, envelope)
}

func (*Appender) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params AppendChangeSetParams) (uuid.UUID, error) {
	if params.ChangeSetID != nil {
		if _, err := tx.Exec(ctx, `
INSERT INTO change_sets (
    change_set_id,
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, *params.ChangeSetID, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()); err != nil {
			return uuid.UUID{}, fmt.Errorf("append change set: %w", err)
		}
		return *params.ChangeSetID, nil
	}
	var changeSetID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO change_sets (
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING change_set_id
`, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()).Scan(&changeSetID); err != nil {
		return uuid.UUID{}, fmt.Errorf("append change set: %w", err)
	}
	return changeSetID, nil
}

func (a *Appender) AppendNonRowMutationTx(ctx context.Context, tx pgx.Tx, params AppendNonRowMutationParams) error {
	return a.appendMutationValuesTx(ctx, tx, params)
}

func (a *Appender) AppendCapturedRecordMutationTx(ctx context.Context, tx pgx.Tx, params AppendCapturedRecordMutationParams) error {
	beforeValue, afterValue, err := capturedSnapshotPair(params.RecordID, params.BeforeSnapshot, params.AfterSnapshot)
	if err != nil {
		return err
	}
	return a.appendMutationValuesTx(ctx, tx, AppendNonRowMutationParams{
		ChangeSetID:     params.ChangeSetID,
		SequenceNo:      params.SequenceNo,
		TargetKind:      params.TargetKind,
		TargetID:        params.RecordID.String(),
		OperationKind:   params.OperationKind,
		BeforeVersionID: params.BeforeVersionID,
		AfterVersionID:  params.AfterVersionID,
		BeforeValue:     beforeValue,
		AfterValue:      afterValue,
	})
}

func (a *Appender) appendMutationValuesTx(ctx context.Context, tx pgx.Tx, params AppendNonRowMutationParams) error {
	description, err := a.targetSemantics.DescribeValues(params.TargetKind, params.TargetID, params.BeforeValue, params.AfterValue)
	if err != nil {
		return fmt.Errorf("describe change-set mutation history: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO change_set_mutations (
    change_set_id,
    sequence_no,
    target_kind,
    target_id,
    operation_kind,
    before_version_id,
    after_version_id,
    before_value,
    after_value,
    history_record_ids,
    history_entry_record_ids
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, params.ChangeSetID, params.SequenceNo, params.TargetKind, params.TargetID, params.OperationKind, params.BeforeVersionID, params.AfterVersionID, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue), description.HistoryRecordIDs, description.HistoryEntryRecordIDs); err != nil {
		return fmt.Errorf("append change-set mutation: %w", err)
	}
	return nil
}

func (a *Appender) AppendCapturedRecordRevisionTx(ctx context.Context, tx pgx.Tx, params AppendCapturedRecordRevisionParams) error {
	if err := a.AppendCapturedRecordRevisionOnlyTx(ctx, tx, params); err != nil {
		return err
	}
	return a.appendRecordRevisionIntentTx(ctx, tx, params.ChangeSetID, params.RecordID, params.RowVersion, params.LiveChange)
}

func (*Appender) AppendCapturedRecordRevisionOnlyTx(ctx context.Context, tx pgx.Tx, params AppendCapturedRecordRevisionParams) error {
	beforeValue, afterValue, err := capturedSnapshotPair(params.RecordID, params.BeforeSnapshot, params.AfterSnapshot)
	if err != nil {
		return err
	}
	revisionID, err := appendRecordRevisionValuesTx(ctx, tx, appendRecordRevisionValuesParams{
		ChangeSetID: params.ChangeSetID,
		RecordID:    params.RecordID,
		RowVersion:  params.RowVersion,
		BeforeValue: beforeValue,
		AfterValue:  afterValue,
	})
	if err != nil {
		return err
	}
	facts, err := recordRevisionConflictFacts(params.LiveChange)
	if err != nil {
		return fmt.Errorf("derive record revision conflict facts: %w", err)
	}
	for _, fact := range facts {
		beforeValue, err := revisionConflictFactValue(fact.BeforeValue, fact.BeforePresent)
		if err != nil {
			return fmt.Errorf("encode record revision conflict fact before value: %w", err)
		}
		afterValue, err := revisionConflictFactValue(fact.AfterValue, fact.AfterPresent)
		if err != nil {
			return fmt.Errorf("encode record revision conflict fact after value: %w", err)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO record_revision_conflict_facts (
    revision_id,
    field_key,
    before_present,
    before_value,
    after_present,
    after_value
)
VALUES ($1, $2, $3, $4, $5, $6)
`, revisionID, fact.FieldKey, fact.BeforePresent, beforeValue, fact.AfterPresent, afterValue); err != nil {
			return fmt.Errorf("append record revision conflict fact: %w", err)
		}
	}
	return nil
}

type appendRecordRevisionValuesParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue map[string]any
	AfterValue  map[string]any
}

func appendRecordRevisionValuesTx(ctx context.Context, tx pgx.Tx, params appendRecordRevisionValuesParams) (int64, error) {
	var revisionID int64
	if err := tx.QueryRow(ctx, `
INSERT INTO record_revisions (
    change_set_id,
    record_id,
    row_version,
    before_json,
    after_json
)
VALUES ($1, $2, $3, $4, $5)
RETURNING revision_id
`, params.ChangeSetID, params.RecordID, params.RowVersion, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)).Scan(&revisionID); err != nil {
		return 0, fmt.Errorf("append record revision: %w", err)
	}
	return revisionID, nil
}

type recordRevisionConflictFact struct {
	FieldKey      string
	BeforePresent bool
	BeforeValue   any
	AfterPresent  bool
	AfterValue    any
}

func revisionConflictFactValue(value any, present bool) (any, error) {
	if !present {
		return nil, nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func recordRevisionConflictFacts(change LiveRecordChange) ([]recordRevisionConflictFact, error) {
	beforeRow, err := collaborationRow(change.BeforeValue)
	if err != nil {
		return nil, err
	}
	afterRow, err := collaborationRow(change.AfterValue)
	if err != nil {
		return nil, err
	}
	changedFieldKeys, err := collaboration.ChangedCellKeys(beforeRow, afterRow)
	if err != nil {
		return nil, err
	}
	beforeCells, err := revisionConflictCells(beforeRow)
	if err != nil {
		return nil, err
	}
	afterCells, err := revisionConflictCells(afterRow)
	if err != nil {
		return nil, err
	}
	facts := make([]recordRevisionConflictFact, 0, len(changedFieldKeys))
	for _, fieldKey := range changedFieldKeys {
		beforeValue, beforePresent := beforeCells[fieldKey]
		afterValue, afterPresent := afterCells[fieldKey]
		facts = append(facts, recordRevisionConflictFact{
			FieldKey:      fieldKey,
			BeforePresent: beforePresent,
			BeforeValue:   beforeValue,
			AfterPresent:  afterPresent,
			AfterValue:    afterValue,
		})
	}
	return facts, nil
}

func revisionConflictCells(row map[string]any) (map[string]any, error) {
	if row == nil || row["cells"] == nil {
		return map[string]any{}, nil
	}
	cells, ok := row["cells"].(map[string]any)
	if ok {
		return cells, nil
	}
	encoded, err := json.Marshal(row["cells"])
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	return decoded, nil
}

func capturedSnapshotPair(recordID uuid.UUID, before *CapturedRecordSnapshot, after *CapturedRecordSnapshot) (map[string]any, map[string]any, error) {
	if before == nil && after == nil {
		return nil, nil, fmt.Errorf("%w: record %s has no before or after value", ErrInvalidCapturedSnapshot, recordID)
	}
	beforeValue, err := capturedSnapshotValue(before, recordID)
	if err != nil {
		return nil, nil, err
	}
	afterValue, err := capturedSnapshotValue(after, recordID)
	if err != nil {
		return nil, nil, err
	}
	if before != nil && after != nil && (before.recordType != after.recordType || before.snapshotSchemaID != after.snapshotSchemaID) {
		return nil, nil, fmt.Errorf("%w: record %s before/after schema mismatch", ErrInvalidCapturedSnapshot, recordID)
	}
	return beforeValue, afterValue, nil
}

func (a *Appender) appendRecordRevisionIntentTx(
	ctx context.Context,
	tx pgx.Tx,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
	rowVersion int64,
	liveChange LiveRecordChange,
) error {
	suppressed, err := a.historicalPolicy.IsSuppressedTx(ctx, tx)
	if err != nil {
		return err
	}
	if suppressed {
		return nil
	}

	envelope, err := a.recordEnvelopes.LoadEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return fmt.Errorf("load record revision collaboration envelope: %w", err)
	}
	var (
		actorUserID uuid.UUID
		clientTxnID *string
		source      string
		createdAt   time.Time
	)
	if err := tx.QueryRow(ctx, `
SELECT actor_user_id, client_txn_id, source, created_at
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(
		&actorUserID,
		&clientTxnID,
		&source,
		&createdAt,
	); err != nil {
		return fmt.Errorf("load record revision collaboration identity: %w", err)
	}
	beforeRow, err := collaborationRow(liveChange.BeforeValue)
	if err != nil {
		return fmt.Errorf("decode record revision before row: %w", err)
	}
	afterRow, err := collaborationRow(liveChange.AfterValue)
	if err != nil {
		return fmt.Errorf("decode record revision after row: %w", err)
	}
	row := afterRow
	if row == nil {
		row = beforeRow
	}
	viewSchemaID, err := a.recordViews.Resolve(envelope.RecordType, row)
	if err != nil {
		return err
	}
	changedFieldKeys, err := collaboration.ChangedCellKeys(beforeRow, afterRow)
	if err != nil {
		return err
	}
	changeKind := ""
	switch {
	case envelope.DeletedAt != nil:
		changeKind = "remove"
	case source == "records.restore" || source == "rollback":
		changeKind = "invalidate"
	}
	var mutationOrdinal int
	if err := tx.QueryRow(ctx, `
SELECT GREATEST(COALESCE(min(sequence_no), 1) - 1, 0)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_id = $2
`, changeSetID, recordID.String()).Scan(&mutationOrdinal); err != nil {
		return fmt.Errorf("load record revision collaboration ordinal: %w", err)
	}
	clientTxn := ""
	if clientTxnID != nil {
		clientTxn = *clientTxnID
	}
	intent, err := collaboration.NewRecordChangeIntent(collaboration.RecordChange{
		IncidentID:       envelope.IncidentID,
		RecordID:         recordID,
		RowVersion:       rowVersion,
		ChangeSetID:      changeSetID,
		ClientTxnID:      clientTxn,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: changedFieldKeys,
		ViewSchemaID:     viewSchemaID,
		ChangeKind:       changeKind,
		Row:              afterRow,
	}, mutationOrdinal, createdAt)
	if err != nil {
		return err
	}
	if err := a.intents.AppendIntentTx(ctx, tx, intent); err != nil {
		return fmt.Errorf("append record revision collaboration intent: %w", err)
	}
	return nil
}

func collaborationRow(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	if row, ok := value.(map[string]any); ok {
		return row, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var row map[string]any
	if err := json.Unmarshal(encoded, &row); err != nil {
		return nil, err
	}
	return row, nil
}
