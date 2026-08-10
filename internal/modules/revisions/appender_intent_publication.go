package revisions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

func (a *Appender) AppendRecordRevisionAndIntentTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	if err := a.AppendRecordRevisionTx(ctx, tx, params); err != nil {
		return err
	}
	return a.appendRecordRevisionIntentTx(ctx, tx, params.ChangeSetID, params.RecordID, params.RowVersion, params.LiveChange)
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
