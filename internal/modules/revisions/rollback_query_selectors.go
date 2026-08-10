package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r rollbackQueryRepository) loadHistoryEntryRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (rollbackPlan, error) {
	target, err := loadHistoryEntryRollbackTargetTx(ctx, tx, record, historyEntryRef)
	if err != nil {
		return rollbackPlan{}, err
	}
	siblings, err := r.loadRollbackSiblingTargetsTx(ctx, tx, target)
	if err != nil {
		return rollbackPlan{}, err
	}
	descriptor, describeErr := r.describeRollbackTargetTx(ctx, tx, record.IncidentID, target, record.RecordID, siblings)
	affected, deferredErr, err := rollbackDescriptorAffected(descriptor, describeErr, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan := rollbackPlan{
		Target:      target,
		Affected:    affected,
		Addressed:   record.RecordID,
		RecordType:  record.RecordType,
		DeferredErr: deferredErr,
	}
	plan.Companion, err = selectRollbackCompanions(descriptor.SingleEntryCompanions, siblings)
	if err != nil {
		return rollbackPlan{}, err
	}
	requiresChangeSet, err := r.historyEntryRequiresChangeSetTx(ctx, tx, record.IncidentID, target, descriptor, siblings)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan.RequiresChangeSet = requiresChangeSet
	return plan, nil
}

func (r rollbackQueryRepository) loadChangeSetRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, raw string) (rollbackPlan, error) {
	changeSetID, err := uuid.Parse(raw)
	if err != nil {
		return rollbackPlan{}, ErrRollbackTargetNotFound
	}
	rows, err := tx.Query(ctx, `
SELECT csm.change_set_id,
       cs.created_at,
       cs.source,
       csm.sequence_no,
       csm.target_kind,
       csm.target_id,
       csm.operation_kind,
       csm.before_value,
       csm.after_value
  FROM change_sets cs
  JOIN change_set_mutations csm
    ON csm.change_set_id = cs.change_set_id
 WHERE cs.change_set_id = $1
   AND cs.incident_id = $2
   AND EXISTS (
       SELECT 1
         FROM change_set_mutations visible
        WHERE visible.change_set_id = cs.change_set_id
          AND $3::uuid = ANY(visible.history_record_ids)
   )
 ORDER BY csm.sequence_no ASC
`, changeSetID, record.IncidentID, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	defer rows.Close()
	targets := make([]rollbackMutationTarget, 0)
	for rows.Next() {
		var (
			target    rollbackMutationTarget
			beforeRaw []byte
			afterRaw  []byte
		)
		if err := rows.Scan(&target.ChangeSetID, &target.CreatedAt, &target.Source, &target.SequenceNo, &target.TargetKind, &target.TargetID, &target.OperationKind, &beforeRaw, &afterRaw); err != nil {
			return rollbackPlan{}, err
		}
		before, err := decodeRollbackValue(beforeRaw)
		if err != nil {
			return rollbackPlan{}, err
		}
		after, err := decodeRollbackValue(afterRaw)
		if err != nil {
			return rollbackPlan{}, err
		}
		target.BeforeValue = before
		target.AfterValue = after
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return rollbackPlan{}, err
	}
	if len(targets) == 0 {
		return rollbackPlan{}, ErrRollbackTargetNotFound
	}
	affected, err := r.affectedRecordsForRollbackTargetsTx(ctx, tx, record.IncidentID, targets, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan := rollbackPlan{
		Target:     targets[0],
		Targets:    targets,
		Affected:   affected,
		Addressed:  record.RecordID,
		RecordType: record.RecordType,
		WholeSet:   true,
	}
	for _, target := range targets {
		siblings := rollbackTargetsExcept(targets, target.SequenceNo)
		_, describeErr := r.describeRollbackTargetTx(ctx, tx, record.IncidentID, target, record.RecordID, siblings)
		if describeErr == nil {
			continue
		}
		adapted := adaptRowRollbackProviderError(describeErr)
		if !deferableRollbackProviderError(adapted) {
			return rollbackPlan{}, adapted
		}
		if plan.DeferredErr == nil {
			plan.DeferredErr = adapted
		}
	}
	return plan, nil
}

func loadHistoryEntryRollbackTargetTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (rollbackMutationTarget, error) {
	row := tx.QueryRow(ctx, `
SELECT csm.change_set_id,
       cs.created_at,
       cs.source,
       csm.sequence_no,
       csm.target_kind,
       csm.target_id,
       csm.operation_kind,
       csm.before_value,
       csm.after_value
  FROM record_history_entry_refs href
  JOIN change_sets cs
    ON cs.change_set_id = href.change_set_id
  JOIN change_set_mutations csm
    ON csm.change_set_id = href.change_set_id
   AND csm.sequence_no = href.mutation_sequence_no
 WHERE href.record_id = $1
   AND href.history_entry_ref = $2
   AND cs.incident_id = $3
`, record.RecordID, historyEntryRef, record.IncidentID)
	var (
		target    rollbackMutationTarget
		beforeRaw []byte
		afterRaw  []byte
	)
	if err := row.Scan(&target.ChangeSetID, &target.CreatedAt, &target.Source, &target.SequenceNo, &target.TargetKind, &target.TargetID, &target.OperationKind, &beforeRaw, &afterRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackMutationTarget{}, ErrRollbackTargetNotFound
		}
		return rollbackMutationTarget{}, err
	}
	before, err := decodeRollbackValue(beforeRaw)
	if err != nil {
		return rollbackMutationTarget{}, err
	}
	after, err := decodeRollbackValue(afterRaw)
	if err != nil {
		return rollbackMutationTarget{}, err
	}
	target.BeforeValue = before
	target.AfterValue = after
	return target, nil
}

func (r rollbackQueryRepository) loadRowRestorePlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, revisionNo int64) (rollbackPlan, error) {
	row := tx.QueryRow(ctx, `
SELECT change_set_id, after_json
  FROM record_revisions
 WHERE record_id = $1
   AND row_version = $2
`, record.RecordID, revisionNo)
	var (
		changeSetID uuid.UUID
		afterRaw    []byte
	)
	if err := row.Scan(&changeSetID, &afterRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackPlan{}, ErrRollbackTargetNotFound
		}
		return rollbackPlan{}, err
	}
	snapshot, err := decodeRollbackValue(afterRaw)
	if err != nil {
		return rollbackPlan{}, err
	}
	if err := r.store.appender.snapshotCaptures.validatePersisted(record.RecordID, record.RecordType, snapshot); err != nil {
		return rollbackPlan{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	targetKind, err := r.store.targetSemantics.defaultRowTargetKind(record.RecordType)
	if err != nil {
		return rollbackPlan{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	provider, err := r.store.targetSemantics.rowProvider(targetKind, record.RecordType)
	if err != nil {
		return rollbackPlan{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := provider.ValidateRollbackValue(snapshot); err != nil {
		return rollbackPlan{}, adaptRowRollbackProviderError(err)
	}
	return rollbackPlan{
		Target: rollbackMutationTarget{
			ChangeSetID: changeSetID,
			TargetKind:  targetKind,
			TargetID:    record.RecordID.String(),
			AfterValue:  snapshot,
		},
		Affected:          []uuid.UUID{record.RecordID},
		Addressed:         record.RecordID,
		RecordType:        record.RecordType,
		RestoreRevisionNo: revisionNo,
		RestoreSnapshot:   snapshot,
	}, nil
}
