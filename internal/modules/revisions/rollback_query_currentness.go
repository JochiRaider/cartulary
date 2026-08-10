package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func (r rollbackQueryRepository) validateCanonicalRowTargetsTx(ctx context.Context, tx pgx.Tx, plan rollbackPlan) error {
	targets := []rollbackMutationTarget{plan.Target}
	if plan.WholeSet {
		targets = plan.Targets
	}
	for _, target := range targets {
		dispatch, err := r.store.targetSemantics.dispatchClass(target.TargetKind)
		if err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		if dispatch == rollbackcontract.DispatchNonRow {
			continue
		}
		if dispatch != rollbackcontract.DispatchRow {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		recordID, err := uuid.Parse(target.TargetID)
		if err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		envelope, err := r.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return err
		}
		if _, err := r.store.targetSemantics.rowProvider(target.TargetKind, envelope.RecordType); err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		if err := r.store.appender.snapshotCaptures.validatePersisted(recordID, envelope.RecordType, target.BeforeValue); err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		if err := r.store.appender.snapshotCaptures.validatePersisted(recordID, envelope.RecordType, target.AfterValue); err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	return nil
}

func (r rollbackQueryRepository) ensurePlanCurrentTx(ctx context.Context, tx pgx.Tx, plan rollbackPlan) error {
	if plan.WholeSet {
		for _, target := range plan.Targets {
			if err := ensureNoLaterRollbackTargetMutationTx(ctx, tx, target, true); err != nil {
				return err
			}
		}
		return nil
	}
	if err := ensureNoLaterRollbackTargetMutationTx(ctx, tx, plan.Target, false); err != nil {
		return err
	}
	for _, companion := range plan.Companion {
		if err := ensureNoLaterRollbackTargetMutationTx(ctx, tx, companion, false); err != nil {
			return err
		}
	}
	return nil
}

func ensureNoLaterRollbackTargetMutationTx(ctx context.Context, tx pgx.Tx, target rollbackMutationTarget, ignoreSameChangeSet bool) error {
	rows, err := tx.Query(ctx, `
SELECT cs.source
  FROM change_sets cs
  JOIN change_set_mutations csm
    ON csm.change_set_id = cs.change_set_id
 WHERE csm.target_kind = $1
   AND csm.target_id = $2
   AND (
       cs.created_at > $3
       OR (cs.created_at = $3 AND csm.change_set_id = $4 AND csm.sequence_no > $5)
   )
   AND ($6::boolean = false OR csm.change_set_id <> $4)
 ORDER BY cs.created_at DESC, csm.sequence_no DESC
 LIMIT 1
`, target.TargetKind, target.TargetID, target.CreatedAt.UTC(), target.ChangeSetID, target.SequenceNo, ignoreSameChangeSet)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return rows.Err()
	}
	var source string
	if err := rows.Scan(&source); err != nil {
		return err
	}
	if source == "rollback" {
		return &RollbackPreconditionError{ReasonCode: "stale_target"}
	}
	return &RollbackPreconditionError{ReasonCode: "dependent_later_changes"}
}
