package revisions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (a *Appender) AppendNonRowMutationTx(ctx context.Context, tx pgx.Tx, params AppendNonRowMutationParams) error {
	return a.appendMutationValuesTx(ctx, tx, params)
}

func (a *Appender) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params AppendRecordMutationParams) error {
	beforeValue, afterValue, err := recordSnapshotPair(params.RecordID, params.BeforeSnapshot, params.AfterSnapshot)
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
