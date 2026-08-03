package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type rollbackQueryRepository struct {
	store *commandStore
}

func (r rollbackQueryRepository) loadRollbackRecordEnvelopesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]rollbackRecordEnvelope, error) {
	envelopes := make(map[uuid.UUID]rollbackRecordEnvelope, len(recordIDs))
	for _, recordID := range canonicalRecordIDs(recordIDs) {
		envelope, err := r.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		envelopes[recordID] = envelope
	}
	return envelopes, nil
}

func (r rollbackQueryRepository) loadRollbackProtectedSetTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, target RollbackTarget) (rollbackProtectedSet, error) {
	fallback := rollbackProtectedSet{Affected: []uuid.UUID{record.RecordID}}
	switch target.Kind {
	case "history_entry":
		mutation, err := loadHistoryEntryRollbackTargetTx(ctx, tx, record, target.HistoryEntryRef)
		if errors.Is(err, ErrRollbackTargetNotFound) {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		if err != nil {
			return rollbackProtectedSet{}, err
		}
		affected, err := r.affectedRecordsForRollbackTargetTx(ctx, tx, record.IncidentID, mutation, record.RecordID)
		if err != nil {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		return rollbackProtectedSet{Affected: affected}, nil
	case "change_set":
		plan, err := r.loadChangeSetRollbackPlanTx(ctx, tx, record, target.ChangeSetID)
		if errors.Is(err, ErrRollbackTargetNotFound) {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		if err != nil {
			return rollbackProtectedSet{}, err
		}
		return rollbackProtectedSet{Affected: plan.Affected}, nil
	case "row_restore":
		return fallback, nil
	default:
		fallback.DeferredErr = &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		return fallback, nil
	}
}

func (r rollbackQueryRepository) loadHistoryEntryRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (rollbackPlan, error) {
	target, err := loadHistoryEntryRollbackTargetTx(ctx, tx, record, historyEntryRef)
	if err != nil {
		return rollbackPlan{}, err
	}
	affected, err := r.affectedRecordsForRollbackTargetTx(ctx, tx, record.IncidentID, target, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan := rollbackPlan{
		Target:     target,
		Affected:   affected,
		Addressed:  record.RecordID,
		RecordType: record.RecordType,
	}
	if provider, ok := r.store.nonRowRollbackProviders.Provider(target.TargetKind); ok {
		_, describeErr := provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: nonRowContractTarget(record.IncidentID, target), AddressedRecordID: record.RecordID})
		if describeErr != nil {
			adapted := adaptRowRollbackProviderError(describeErr)
			if !deferableRollbackProviderError(adapted) {
				return rollbackPlan{}, adapted
			}
			plan.DeferredErr = adapted
		}
	}
	if target.TargetKind == "entity_mention" {
		companion, err := loadRollbackMentionCompanionLinkTargetsTx(ctx, tx, target)
		if err != nil {
			return rollbackPlan{}, err
		}
		plan.Companion = companion
	}
	requiresChangeSet, err := r.historyEntryRequiresChangeSetTx(ctx, tx, record.IncidentID, target)
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
          AND (
              visible.target_id = $3
              OR (
                  visible.target_kind = 'record_link'
                  AND (
                      visible.before_value ->> 'src_record_id' = $3
                      OR visible.before_value ->> 'dst_record_id' = $3
                      OR visible.after_value ->> 'src_record_id' = $3
                      OR visible.after_value ->> 'dst_record_id' = $3
                  )
              )
              OR (
                  visible.target_kind = 'entity_mention'
                  AND (
                      visible.before_value ->> 'source_record_id' = $3
                      OR visible.after_value ->> 'source_record_id' = $3
                  )
              )
              OR (
                  visible.target_kind = 'record_tag'
                  AND (
                      visible.before_value ->> 'record_id' = $3
                      OR visible.after_value ->> 'record_id' = $3
                  )
              )
              OR (
                  visible.target_kind = 'indicator_observation'
                  AND (
                      visible.before_value ->> 'source_record_id' = $3
                      OR visible.after_value ->> 'source_record_id' = $3
                      OR visible.before_value ->> 'resolved_indicator_record_id' = $3
                      OR visible.after_value ->> 'resolved_indicator_record_id' = $3
                  )
              )
              OR (
                  visible.target_kind = 'indicator_state_interval'
                  AND (
                      visible.before_value ->> 'indicator_record_id' = $3
                      OR visible.after_value ->> 'indicator_record_id' = $3
                  )
              )
          )
   )
 ORDER BY csm.sequence_no ASC
`, changeSetID, record.IncidentID, record.RecordID.String())
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
		provider, ok := r.store.nonRowRollbackProviders.Provider(target.TargetKind)
		if !ok {
			continue
		}
		_, describeErr := provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: nonRowContractTarget(record.IncidentID, target), AddressedRecordID: record.RecordID})
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

func (r rollbackQueryRepository) loadRollbackRecordEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, forUpdate bool) (rollbackRecordEnvelope, error) {
	record, err := r.store.envelopes.LoadEnvelopeTx(ctx, tx, recordID, forUpdate)
	if err != nil {
		return rollbackRecordEnvelope{}, adaptEnvelopeError(err)
	}
	return record, nil
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
	provider, ok := r.store.rowRollbackProviders.Provider(record.RecordType)
	if !ok {
		return rollbackPlan{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := provider.ValidateRollbackValue(snapshot); err != nil {
		return rollbackPlan{}, adaptRowRollbackProviderError(err)
	}
	return rollbackPlan{
		Target: rollbackMutationTarget{
			ChangeSetID: changeSetID,
			TargetKind:  rollbackMutationTargetKindForRecordType(record.RecordType),
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

func loadRollbackMentionCompanionLinkTargetsTx(ctx context.Context, tx pgx.Tx, target rollbackMutationTarget) ([]rollbackMutationTarget, error) {
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
  FROM change_set_mutations csm
  JOIN change_sets cs
    ON cs.change_set_id = csm.change_set_id
 WHERE csm.change_set_id = $1
   AND csm.sequence_no <> $2
   AND csm.target_kind = 'record_link'
 ORDER BY csm.sequence_no ASC
`, target.ChangeSetID, target.SequenceNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var companions []rollbackMutationTarget
	for rows.Next() {
		var (
			companion rollbackMutationTarget
			beforeRaw []byte
			afterRaw  []byte
		)
		if err := rows.Scan(&companion.ChangeSetID, &companion.CreatedAt, &companion.Source, &companion.SequenceNo, &companion.TargetKind, &companion.TargetID, &companion.OperationKind, &beforeRaw, &afterRaw); err != nil {
			return nil, err
		}
		before, err := decodeRollbackValue(beforeRaw)
		if err != nil {
			return nil, err
		}
		after, err := decodeRollbackValue(afterRaw)
		if err != nil {
			return nil, err
		}
		companion.BeforeValue = before
		companion.AfterValue = after
		companions = append(companions, companion)
	}
	return companions, rows.Err()
}

func (r rollbackQueryRepository) historyEntryRequiresChangeSetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, target rollbackMutationTarget) (bool, error) {
	targets := []rollbackMutationTarget{target}
	rows, err := tx.Query(ctx, `
SELECT sequence_no, target_kind, target_id, operation_kind, before_value, after_value
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND sequence_no <> $2
 ORDER BY sequence_no
`, target.ChangeSetID, target.SequenceNo)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var candidate rollbackMutationTarget
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(&candidate.SequenceNo, &candidate.TargetKind, &candidate.TargetID, &candidate.OperationKind, &beforeRaw, &afterRaw); err != nil {
			return false, err
		}
		candidate.ChangeSetID = target.ChangeSetID
		candidate.BeforeValue, err = decodeRollbackValue(beforeRaw)
		if err != nil {
			return false, err
		}
		candidate.AfterValue, err = decodeRollbackValue(afterRaw)
		if err != nil {
			return false, err
		}
		targets = append(targets, candidate)
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	for _, candidate := range targets {
		provider, ok := r.store.nonRowRollbackProviders.Provider(candidate.TargetKind)
		if !ok {
			continue
		}
		descriptor, err := provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: nonRowContractTarget(incidentID, candidate)})
		if err != nil {
			adapted := adaptRowRollbackProviderError(err)
			var precondition *RollbackPreconditionError
			if !errors.Is(adapted, ErrRollbackTargetNotFound) && !errors.As(adapted, &precondition) {
				return false, adapted
			}
		}
		if descriptor.RequiresWholeChangeSet || (firstClassRollbackTargetKind(target.TargetKind) && candidate.SequenceNo != target.SequenceNo && len(descriptor.AtomicCompanions) > 0) {
			return true, nil
		}
	}
	return false, nil
}

func (r rollbackQueryRepository) affectedRecordsForRollbackTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, target rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	provider, ok := r.store.nonRowRollbackProviders.Provider(target.TargetKind)
	if !ok {
		return affectedRecordsForRollbackTarget(target, fallback)
	}
	descriptor, err := provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{
		Target:            nonRowContractTarget(incidentID, target),
		AddressedRecordID: fallback,
	})
	if err != nil {
		adapted := adaptRowRollbackProviderError(err)
		var precondition *RollbackPreconditionError
		if !errors.Is(adapted, ErrRollbackTargetNotFound) && !errors.As(adapted, &precondition) {
			return nil, adapted
		}
		if affected := canonicalRecordIDs(descriptor.AffectedRecordIDs); len(affected) > 0 {
			return affected, nil
		}
		return []uuid.UUID{fallback}, nil
	}
	affected := canonicalRecordIDs(descriptor.AffectedRecordIDs)
	if len(affected) == 0 {
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return affected, nil
}

func (r rollbackQueryRepository) affectedRecordsForRollbackTargetsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, targets []rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	recordIDs := map[uuid.UUID]struct{}{fallback: {}}
	for _, target := range targets {
		affected, err := r.affectedRecordsForRollbackTargetTx(ctx, tx, incidentID, target, fallback)
		if err != nil {
			return nil, err
		}
		for _, recordID := range affected {
			recordIDs[recordID] = struct{}{}
		}
	}
	values := make([]uuid.UUID, 0, len(recordIDs))
	for recordID := range recordIDs {
		if recordID != uuid.Nil {
			values = append(values, recordID)
		}
	}
	return canonicalRecordIDs(values), nil
}
