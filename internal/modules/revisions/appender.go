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
	"github.com/JochiRaider/cartulary/internal/modules/records"
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
	historicalPolicy HistoricalIntentPolicy
	intents          IntentAppender
}

func NewAppender(
	recordViews *RecordViewCatalog,
	historicalPolicy HistoricalIntentPolicy,
	intents IntentAppender,
) (*Appender, error) {
	if recordViews == nil {
		return nil, errors.New("revisions: record/view catalog is required")
	}
	if historicalPolicy == nil {
		return nil, errors.New("revisions: historical intent policy is required")
	}
	if intents == nil {
		return nil, errors.New("revisions: Collaboration intent appender is required")
	}
	return &Appender{recordViews: recordViews, historicalPolicy: historicalPolicy, intents: intents}, nil
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

type AppendMutationParams struct {
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

type AppendRecordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
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

func (*Appender) AppendMutationTx(ctx context.Context, tx pgx.Tx, params AppendMutationParams) error {
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
    after_value
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, params.ChangeSetID, params.SequenceNo, params.TargetKind, params.TargetID, params.OperationKind, params.BeforeVersionID, params.AfterVersionID, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)); err != nil {
		return fmt.Errorf("append change-set mutation: %w", err)
	}
	return nil
}

func (a *Appender) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	if err := a.AppendRecordRevisionOnlyTx(ctx, tx, params); err != nil {
		return err
	}
	return a.appendRecordRevisionIntentTx(ctx, tx, params)
}

// AppendRecordRevisionOnlyTx persists history for a source owner that appends
// its own typed Collaboration intent later in the same transaction.
func (*Appender) AppendRecordRevisionOnlyTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO record_revisions (
    change_set_id,
    record_id,
    row_version,
    before_json,
    after_json
)
VALUES ($1, $2, $3, $4, $5)
`, params.ChangeSetID, params.RecordID, params.RowVersion, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)); err != nil {
		return fmt.Errorf("append record revision: %w", err)
	}
	return nil
}

func (a *Appender) appendRecordRevisionIntentTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	suppressed, err := a.historicalPolicy.IsSuppressedTx(ctx, tx)
	if err != nil {
		return err
	}
	if suppressed {
		return nil
	}

	envelope, err := records.NewStore().LoadEnvelopeTx(ctx, tx, params.RecordID, false)
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
`, params.ChangeSetID).Scan(
		&actorUserID,
		&clientTxnID,
		&source,
		&createdAt,
	); err != nil {
		return fmt.Errorf("load record revision collaboration identity: %w", err)
	}
	beforeRow, err := collaborationRow(params.BeforeValue)
	if err != nil {
		return fmt.Errorf("decode record revision before row: %w", err)
	}
	afterRow, err := collaborationRow(params.AfterValue)
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
`, params.ChangeSetID, params.RecordID.String()).Scan(&mutationOrdinal); err != nil {
		return fmt.Errorf("load record revision collaboration ordinal: %w", err)
	}
	clientTxn := ""
	if clientTxnID != nil {
		clientTxn = *clientTxnID
	}
	intent, err := collaboration.NewRecordChangeIntent(collaboration.RecordChange{
		IncidentID:       envelope.IncidentID,
		RecordID:         params.RecordID,
		RowVersion:       params.RowVersion,
		ChangeSetID:      params.ChangeSetID,
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
