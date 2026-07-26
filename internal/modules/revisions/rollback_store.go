package revisions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var (
	ErrRollbackTargetNotFound      = errors.New("revisions: rollback target not found")
	ErrRollbackPreconditionFailed  = errors.New("revisions: rollback precondition failed")
	ErrUnsupportedRollbackMutation = errors.New("revisions: unsupported rollback mutation")
)

type RollbackPreconditionError struct {
	ReasonCode string
}

func (e *RollbackPreconditionError) Error() string {
	return ErrRollbackPreconditionFailed.Error()
}

func (e *RollbackPreconditionError) Unwrap() error {
	return ErrRollbackPreconditionFailed
}

type RollbackResult struct {
	Payload     map[string]any
	StatusCode  int
	IncidentID  uuid.UUID
	ClientTxnID string
	Changes     []RollbackRecordChange
	Replayed    bool
}

type RollbackRecordChange struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type rollbackRecordEnvelope struct {
	IncidentID      uuid.UUID
	RecordID        uuid.UUID
	RecordType      string
	RowVersion      int64
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type rollbackMutationTarget struct {
	ChangeSetID   uuid.UUID
	CreatedAt     time.Time
	Source        string
	SequenceNo    int
	TargetKind    string
	TargetID      string
	OperationKind string
	BeforeValue   map[string]any
	AfterValue    map[string]any
}

type rollbackPlan struct {
	Target            rollbackMutationTarget
	Targets           []rollbackMutationTarget
	Companion         []rollbackMutationTarget
	Affected          []uuid.UUID
	Addressed         uuid.UUID
	RecordType        string
	WholeSet          bool
	RequiresChangeSet bool
	RestoreRevisionNo int64
	RestoreSnapshot   map[string]any
	DeferredErr       error
}

type rollbackApplyResult struct {
	ChangeSetID uuid.UUID
	Changes     []RollbackRecordChange
}

type rollbackProtectedSet struct {
	Affected    []uuid.UUID
	DeferredErr error
}

func (s *commandStore) RollbackRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request RollbackRequest, requestHash []byte, requestID string, now time.Time) (RollbackResult, error) {
	authStore := authn.NewStore(s.db)
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    rollbackRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return RollbackResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredRollbackPayload(existing.ResponseJSON)
		if err != nil {
			return RollbackResult{}, err
		}
		return rollbackResultFromPayload(payload, request.ClientTxnID), nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return RollbackResult{}, fmt.Errorf("query rollback idempotency: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return RollbackResult{}, fmt.Errorf("begin rollback transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackResult{}, err
	}

	protected, err := s.loadRollbackProtectedSetTx(ctx, tx, record, request.Target)
	if err != nil {
		return RollbackResult{}, err
	}
	if err := lockDestructiveOperationRecordsNowaitTx(ctx, tx, protected.Affected); err != nil {
		return RollbackResult{}, err
	}
	record, err = loadRollbackRecordEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
		return RollbackResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, record.IncidentID); err != nil {
		return RollbackResult{}, err
	}
	if record.DeletedAt != nil {
		return RollbackResult{}, ErrRecordDeletedUseRestore
	}
	if record.RowVersion != request.BaseRowVersion {
		return RollbackResult{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: record.RowVersion}
	}
	if protected.DeferredErr != nil {
		return RollbackResult{}, protected.DeferredErr
	}

	var plan rollbackPlan
	switch request.Target.Kind {
	case "history_entry":
		plan, err = s.loadHistoryEntryRollbackPlanTx(ctx, tx, record, request.Target.HistoryEntryRef)
	case "change_set":
		plan, err = s.loadChangeSetRollbackPlanTx(ctx, tx, record, request.Target.ChangeSetID)
	case "row_restore":
		plan, err = s.loadRowRestorePlanTx(ctx, tx, record, request.Target.RestoreToRevisionNo)
	default:
		err = &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err != nil {
		return RollbackResult{}, err
	}
	if err := s.validateRollbackPlan(plan); err != nil {
		return RollbackResult{}, err
	}
	if request.Target.Kind != "row_restore" {
		if err := ensureNoLaterRollbackPlanMutationTx(ctx, tx, plan); err != nil {
			return RollbackResult{}, err
		}
	}

	var applied rollbackApplyResult
	if request.Target.Kind == "row_restore" {
		applied, err = s.applyRowRestorePlanTx(ctx, tx, actor, record, plan, request, requestID, now.UTC())
	} else {
		applied, err = s.applyRollbackPlanTx(ctx, tx, actor, record, plan, request, requestID, now.UTC())
	}
	if err != nil {
		return RollbackResult{}, err
	}
	rowVersion := rollbackPayloadRowVersion(record.RecordID, record.RowVersion, applied.Changes)
	payload := buildRollbackPayload(record.IncidentID, record.RecordID, rowVersion, request.Target, plan.Target.ChangeSetID, applied.ChangeSetID, plan.Affected)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return RollbackResult{}, authn.ErrClientTxnConflict
		}
		return RollbackResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RollbackResult{}, fmt.Errorf("commit rollback transaction: %w", err)
	}
	return RollbackResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  record.IncidentID,
		ClientTxnID: request.ClientTxnID,
		Changes:     applied.Changes,
	}, nil
}

func (s *commandStore) loadRollbackProtectedSetTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, target RollbackTarget) (rollbackProtectedSet, error) {
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
		affected, err := s.affectedRecordsForRollbackTargetTx(ctx, tx, record.IncidentID, mutation, record.RecordID)
		if err != nil {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		return rollbackProtectedSet{Affected: affected}, nil
	case "change_set":
		plan, err := s.loadChangeSetRollbackPlanTx(ctx, tx, record, target.ChangeSetID)
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

func (s *commandStore) loadHistoryEntryRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (rollbackPlan, error) {
	target, err := loadHistoryEntryRollbackTargetTx(ctx, tx, record, historyEntryRef)
	if err != nil {
		return rollbackPlan{}, err
	}
	affected, err := s.affectedRecordsForRollbackTargetTx(ctx, tx, record.IncidentID, target, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan := rollbackPlan{
		Target:     target,
		Affected:   affected,
		Addressed:  record.RecordID,
		RecordType: record.RecordType,
	}
	if provider, ok := s.nonRowRollbackProviders.Provider(target.TargetKind); ok {
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
	requiresChangeSet, err := s.historyEntryRequiresChangeSetTx(ctx, tx, record.IncidentID, target)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan.RequiresChangeSet = requiresChangeSet
	return plan, nil
}

func (s *commandStore) loadChangeSetRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, raw string) (rollbackPlan, error) {
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
	affected, err := s.affectedRecordsForRollbackTargetsTx(ctx, tx, record.IncidentID, targets, record.RecordID)
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
		provider, ok := s.nonRowRollbackProviders.Provider(target.TargetKind)
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

func loadRollbackRecordEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, forUpdate bool) (rollbackRecordEnvelope, error) {
	query := `
SELECT incident_id, record_id, record_type, row_version, deleted_at, deleted_by_user_id::text
  FROM records
 WHERE record_id = $1`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	var (
		record       rollbackRecordEnvelope
		deletedAt    sql.NullTime
		deletedByRaw sql.NullString
	)
	if err := tx.QueryRow(ctx, query, recordID).Scan(&record.IncidentID, &record.RecordID, &record.RecordType, &record.RowVersion, &deletedAt, &deletedByRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackRecordEnvelope{}, ErrRecordNotFound
		}
		return rollbackRecordEnvelope{}, err
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	if deletedByRaw.Valid {
		parsed, err := uuid.Parse(deletedByRaw.String)
		if err != nil {
			return rollbackRecordEnvelope{}, err
		}
		record.DeletedByUserID = &parsed
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

func (s *commandStore) loadRowRestorePlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, revisionNo int64) (rollbackPlan, error) {
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
	provider, ok := s.rowRollbackProviders.Provider(record.RecordType)
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

func ensureNoLaterRollbackPlanMutationTx(ctx context.Context, tx pgx.Tx, plan rollbackPlan) error {
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

func (s *commandStore) applyRollbackPlanTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, record rollbackRecordEnvelope, plan rollbackPlan, request RollbackRequest, requestID string, now time.Time) (rollbackApplyResult, error) {
	changeSetID, err := s.appender.AppendChangeSetTx(ctx, tx, AppendChangeSetParams{
		IncidentID:  record.IncidentID,
		ActorUserID: actor.ID,
		Source:      "rollback",
		Reason:      request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return rollbackApplyResult{}, err
	}
	sequenceNo := 1
	var changes []RollbackRecordChange
	if plan.WholeSet {
		changes, err = s.applyChangeSetRollbackPlanTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		return rollbackApplyResult{ChangeSetID: changeSetID, Changes: changes}, nil
	}
	switch plan.Target.TargetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		change, err := s.applyRowBackedRollbackTx(ctx, tx, actor, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, change)
	case "record_link", "entity_alias", "entity_preserved_identifier", "indicator_observation", "indicator_state_interval":
		linkChanges, err := s.applyRecordLinkRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, linkChanges...)
	case "entity_mention":
		mentionChanges, err := s.applyMentionRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, mentionChanges...)
	case "record_tag":
		tagChanges, err := s.applyRecordTagRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, tagChanges...)
	default:
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return rollbackApplyResult{ChangeSetID: changeSetID, Changes: changes}, nil
}

func (s *commandStore) applyRowRestorePlanTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, record rollbackRecordEnvelope, plan rollbackPlan, request RollbackRequest, requestID string, now time.Time) (rollbackApplyResult, error) {
	provider, ok := s.deleteRestoreProviders.Provider(record.RecordType)
	if !ok {
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := s.snapshotRecordTx(ctx, tx, record.RecordID, provider)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, record.RecordID, actor.ID, now)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	if err := s.restoreRollbackSourceTx(ctx, tx, record.RecordType, record.RecordID, actor.ID, now, nextRowVersion, plan.RestoreSnapshot); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := s.rebuildProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return rollbackApplyResult{}, err
	}
	afterSnapshot, err := s.snapshotRecordTx(ctx, tx, record.RecordID, provider)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	changeSetID, err := s.appender.AppendChangeSetTx(ctx, tx, AppendChangeSetParams{
		IncidentID:  record.IncidentID,
		ActorUserID: actor.ID,
		Source:      "rollback",
		Reason:      request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return rollbackApplyResult{}, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", record.RecordID, record.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", record.RecordID, nextRowVersion)
	targetKind := rollbackMutationTargetKindForRecordType(record.RecordType)
	if targetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", targetKind, record.RecordID, record.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", targetKind, record.RecordID, nextRowVersion)
	}
	if err := s.appender.AppendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      targetKind,
		TargetID:        record.RecordID.String(),
		OperationKind:   "row_restore",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := s.appender.AppendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    record.RecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	change := RollbackRecordChange{
		RecordID:         record.RecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot, afterSnapshot),
	}
	return rollbackApplyResult{ChangeSetID: changeSetID, Changes: []RollbackRecordChange{change}}, nil
}

func (s *commandStore) applyChangeSetRollbackPlanTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	beforeSnapshots, err := s.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	nextVersions := make(map[uuid.UUID]int64, len(plan.Affected))
	for _, recordID := range plan.Affected {
		nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, recordID, actor.ID, now)
		if err != nil {
			return nil, err
		}
		nextVersions[recordID] = nextRowVersion
	}

	for i := len(plan.Targets) - 1; i >= 0; i-- {
		target := plan.Targets[i]
		switch target.TargetKind {
		case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
			_, _, err := s.applyRowBackedRollbackMutationTx(ctx, tx, actor, target, changeSetID, sequenceNo, now, nextVersions)
			if err != nil {
				return nil, err
			}
		case "record_link", "entity_mention", "record_tag", "entity_preserved_identifier", "entity_alias", "indicator_observation", "indicator_state_interval":
			_, err := s.applyNonRowRollbackMutationTx(ctx, tx, actor, incidentID, target, changeSetID, sequenceNo, now)
			if err != nil {
				return nil, err
			}
		default:
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	if err := s.rebuildProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}

	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := s.insertRollbackRecordRevisionSnapshotTx(ctx, tx, recordID, changeSetID, beforeSnapshots[recordID], nextVersions[recordID])
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (s *commandStore) applyRowBackedRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time, nextVersions map[uuid.UUID]int64) (uuid.UUID, []string, error) {
	targetRecordID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return uuid.UUID{}, nil, ErrRollbackTargetNotFound
	}
	targetRecord, err := loadRollbackRecordEnvelopeTx(ctx, tx, targetRecordID, false)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	recordType, err := rollbackRecordTypeForTarget(target, targetRecord.RecordType)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	provider, ok := s.deleteRestoreProviders.Provider(recordType)
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := provider.SnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	nextRowVersion, ok := nextVersions[targetRecordID]
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := s.restoreRollbackSourceTx(ctx, tx, recordType, targetRecordID, actor.ID, now, nextRowVersion, target.BeforeValue); err != nil {
		return uuid.UUID{}, nil, err
	}
	afterSnapshot, err := provider.SnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, nextRowVersion)
	if target.TargetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	}
	if err := s.appender.AppendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      *sequenceNo,
		TargetKind:      target.TargetKind,
		TargetID:        target.TargetID,
		OperationKind:   "rollback",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
	}); err != nil {
		return uuid.UUID{}, nil, err
	}
	(*sequenceNo)++
	return targetRecordID, rollbackChangedFieldKeys(target.BeforeValue, target.AfterValue), nil
}

func (s *commandStore) applyNonRowRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time) (rollbackcontract.ApplyInverseResult, error) {
	result, err := s.executeNonRowInverseTx(ctx, tx, actor, incidentID, target, now)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, result.BeforeValue, result.AfterValue); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return result, nil
}

func (s *commandStore) executeNonRowInverseTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, target rollbackMutationTarget, now time.Time) (rollbackcontract.ApplyInverseResult, error) {
	provider, ok := s.nonRowRollbackProviders.Provider(target.TargetKind)
	if !ok {
		return rollbackcontract.ApplyInverseResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	contractTarget := nonRowContractTarget(incidentID, target)
	descriptor, err := provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: contractTarget})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, adaptRowRollbackProviderError(err)
	}
	result, err := provider.ApplyInverseTx(ctx, tx, rollbackcontract.ApplyInverseRequest{
		Target:      contractTarget,
		ActorUserID: actor.ID,
		Now:         now,
	})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, adaptRowRollbackProviderError(err)
	}
	if err := validateNonRowApplyResult(descriptor, result); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return result, nil
}

func (s *commandStore) insertRollbackRecordRevisionSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changeSetID uuid.UUID, beforeSnapshot map[string]any, rowVersion int64) (RollbackRecordChange, error) {
	record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	provider, ok := s.deleteRestoreProviders.Provider(record.RecordType)
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	afterSnapshot, err := s.snapshotRecordTx(ctx, tx, recordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := s.appender.AppendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         recordID,
		RowVersion:       rowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot, afterSnapshot),
	}, nil
}

func (s *commandStore) applyRowBackedRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) (RollbackRecordChange, error) {
	target := plan.Target
	targetRecordID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return RollbackRecordChange{}, ErrRollbackTargetNotFound
	}
	targetRecord, err := loadRollbackRecordEnvelopeTx(ctx, tx, targetRecordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	recordType, err := rollbackRecordTypeForTarget(target, targetRecord.RecordType)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	provider, ok := s.deleteRestoreProviders.Provider(recordType)
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := s.snapshotRecordTx(ctx, tx, targetRecordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, targetRecordID, actor.ID, now)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := s.restoreRollbackSourceTx(ctx, tx, recordType, targetRecordID, actor.ID, now, nextRowVersion, target.BeforeValue); err != nil {
		return RollbackRecordChange{}, err
	}
	if err := s.rebuildProjectionsTx(ctx, tx, targetRecord.IncidentID); err != nil {
		return RollbackRecordChange{}, err
	}
	afterSnapshot, err := s.snapshotRecordTx(ctx, tx, targetRecordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, nextRowVersion)
	if target.TargetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	}
	if err := s.appender.AppendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      *sequenceNo,
		TargetKind:      target.TargetKind,
		TargetID:        target.TargetID,
		OperationKind:   "rollback",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	(*sequenceNo)++
	if err := s.appender.AppendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    targetRecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         targetRecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot, afterSnapshot),
	}, nil
}

func (s *commandStore) insertRollbackMutationTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, sequenceNo *int, target rollbackMutationTarget, beforeValue any, afterValue any) error {
	if err := s.appender.AppendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:   changeSetID,
		SequenceNo:    *sequenceNo,
		TargetKind:    target.TargetKind,
		TargetID:      target.TargetID,
		OperationKind: "rollback",
		BeforeValue:   beforeValue,
		AfterValue:    afterValue,
	}); err != nil {
		return err
	}
	(*sequenceNo)++
	return nil
}

func (s *commandStore) applyRecordLinkRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	return s.applyNonRowRollbackTx(ctx, tx, actor, incidentID, plan, changeSetID, sequenceNo, now)
}

func (s *commandStore) applyMentionRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	target := plan.Target
	beforeSnapshots, err := s.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	mentionResult, err := s.executeNonRowInverseTx(ctx, tx, actor, incidentID, target, now)
	if err != nil {
		return nil, err
	}
	if err := validateNonRowApplyResult(rollbackcontract.TargetDescriptor{AffectedRecordIDs: plan.Affected}, mentionResult); err != nil {
		return nil, err
	}
	companionResults := make([]rollbackcontract.ApplyInverseResult, 0, len(plan.Companion))
	for _, companion := range plan.Companion {
		result, err := s.executeNonRowInverseTx(ctx, tx, actor, incidentID, companion, now)
		if err != nil {
			return nil, err
		}
		companionResults = append(companionResults, result)
	}
	nextVersions, err := s.advanceRollbackAffectedRecordsTx(ctx, tx, actor, plan.Affected, now)
	if err != nil {
		return nil, err
	}
	if err := s.rebuildProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, mentionResult.BeforeValue, mentionResult.AfterValue); err != nil {
		return nil, err
	}
	for index, companion := range plan.Companion {
		result := companionResults[index]
		if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, companion, result.BeforeValue, result.AfterValue); err != nil {
			return nil, err
		}
	}
	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := s.insertRollbackRecordRevisionSnapshotTx(ctx, tx, recordID, changeSetID, beforeSnapshots[recordID], nextVersions[recordID])
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (s *commandStore) applyRecordTagRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	return s.applyNonRowRollbackTx(ctx, tx, actor, incidentID, plan, changeSetID, sequenceNo, now)
}

func (s *commandStore) applyNonRowRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	beforeSnapshots, err := s.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	result, err := s.applyNonRowRollbackMutationTx(ctx, tx, actor, incidentID, plan.Target, changeSetID, sequenceNo, now)
	if err != nil {
		return nil, err
	}
	if err := validateNonRowApplyResult(rollbackcontract.TargetDescriptor{AffectedRecordIDs: plan.Affected}, result); err != nil {
		return nil, err
	}
	nextVersions, err := s.advanceRollbackAffectedRecordsTx(ctx, tx, actor, plan.Affected, now)
	if err != nil {
		return nil, err
	}
	if err := s.rebuildProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := s.insertRollbackRecordRevisionSnapshotTx(ctx, tx, recordID, changeSetID, beforeSnapshots[recordID], nextVersions[recordID])
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (s *commandStore) snapshotRollbackAffectedRecordsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	snapshots := make(map[uuid.UUID]map[string]any, len(recordIDs))
	for _, recordID := range recordIDs {
		record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		provider, ok := s.deleteRestoreProviders.Provider(record.RecordType)
		if !ok {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		snapshot, err := s.snapshotRecordTx(ctx, tx, recordID, provider)
		if err != nil {
			return nil, err
		}
		snapshots[recordID] = snapshot
	}
	return snapshots, nil
}

func (s *commandStore) advanceRollbackAffectedRecordsTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, recordIDs []uuid.UUID, now time.Time) (map[uuid.UUID]int64, error) {
	nextVersions := make(map[uuid.UUID]int64, len(recordIDs))
	for _, recordID := range recordIDs {
		nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, recordID, actor.ID, now)
		if err != nil {
			return nil, err
		}
		nextVersions[recordID] = nextRowVersion
	}
	return nextVersions, nil
}

func (s *commandStore) validateRollbackPlan(plan rollbackPlan) error {
	if plan.DeferredErr != nil {
		return plan.DeferredErr
	}
	if plan.RestoreRevisionNo > 0 {
		if len(plan.Affected) != 1 || plan.RestoreSnapshot == nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		return nil
	}
	if plan.RequiresChangeSet {
		return &RollbackPreconditionError{ReasonCode: "entry_requires_change_set"}
	}
	if plan.WholeSet {
		if len(plan.Targets) == 0 {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		for _, target := range plan.Targets {
			if err := s.validateRollbackTarget(target); err != nil {
				return err
			}
		}
		return nil
	}
	return s.validateRollbackTarget(plan.Target)
}

func deferableRollbackProviderError(err error) bool {
	if errors.Is(err, ErrRollbackTargetNotFound) {
		return true
	}
	var precondition *RollbackPreconditionError
	return errors.As(err, &precondition)
}

func (s *commandStore) validateRollbackTarget(target rollbackMutationTarget) error {
	if firstClassRollbackTargetKind(target.TargetKind) {
		if target.BeforeValue == nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		return nil
	}
	if _, ok := s.nonRowRollbackProviders.Provider(target.TargetKind); ok {
		return nil
	}
	return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
}

func rollbackRecordTypeForTarget(target rollbackMutationTarget, currentRecordType string) (string, error) {
	if currentRecordType == "" {
		return "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if target.TargetKind != "record" && rollbackMutationTargetKindForRecordType(currentRecordType) != target.TargetKind {
		return "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return currentRecordType, nil
}

func rollbackMutationTargetKindForRecordType(recordType string) string {
	switch recordType {
	case "timeline_event":
		return "timeline_record"
	case "host", "identity", "indicator", "assessment", "evidence":
		return recordType
	default:
		return "record"
	}
}

func rollbackPayloadRowVersion(recordID uuid.UUID, fallback int64, changes []RollbackRecordChange) int64 {
	for _, change := range changes {
		if change.RecordID == recordID {
			return change.RowVersion
		}
	}
	return fallback
}

func updateRollbackRecordEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	var rowVersion int64
	if err := tx.QueryRow(ctx, `
UPDATE records
   SET row_version = row_version + 1,
       updated_at = $2,
       updated_by_user_id = $3
 WHERE record_id = $1
RETURNING row_version
`, recordID, now.UTC(), actorUserID).Scan(&rowVersion); err != nil {
		return 0, err
	}
	return rowVersion, nil
}

func (s *commandStore) restoreRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordType string, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, retainedValue map[string]any) error {
	provider, ok := s.rowRollbackProviders.Provider(recordType)
	if !ok {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := provider.ValidateRollbackValue(retainedValue); err != nil {
		return adaptRowRollbackProviderError(err)
	}
	return adaptRowRollbackProviderError(provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
		RecordID:       recordID,
		ActorUserID:    actorUserID,
		Now:            now.UTC(),
		NextRowVersion: rowVersion,
		RetainedValue:  retainedValue,
	}))
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

func (s *commandStore) historyEntryRequiresChangeSetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, target rollbackMutationTarget) (bool, error) {
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
		provider, ok := s.nonRowRollbackProviders.Provider(candidate.TargetKind)
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

func (s *commandStore) affectedRecordsForRollbackTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, target rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	provider, ok := s.nonRowRollbackProviders.Provider(target.TargetKind)
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

func (s *commandStore) affectedRecordsForRollbackTargetsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, targets []rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	recordIDs := map[uuid.UUID]struct{}{fallback: {}}
	for _, target := range targets {
		affected, err := s.affectedRecordsForRollbackTargetTx(ctx, tx, incidentID, target, fallback)
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

func nonRowContractTarget(incidentID uuid.UUID, target rollbackMutationTarget) rollbackcontract.NonRowTarget {
	return rollbackcontract.NonRowTarget{
		IncidentID:    incidentID,
		ChangeSetID:   target.ChangeSetID,
		SequenceNo:    target.SequenceNo,
		TargetKind:    target.TargetKind,
		TargetID:      target.TargetID,
		OperationKind: target.OperationKind,
		BeforeValue:   target.BeforeValue,
		AfterValue:    target.AfterValue,
	}
}

func affectedRecordsForRollbackTarget(target rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	if !firstClassRollbackTargetKind(target.TargetKind) {
		return []uuid.UUID{fallback}, nil
	}
	recordID, err := uuid.Parse(target.TargetID)
	if err != nil || recordID == uuid.Nil {
		return nil, ErrRollbackTargetNotFound
	}
	return []uuid.UUID{recordID}, nil
}

func firstClassRollbackTargetKind(targetKind string) bool {
	switch targetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		return true
	default:
		return false
	}
}

func buildRollbackPayload(incidentID uuid.UUID, recordID uuid.UUID, rowVersion int64, target RollbackTarget, targetChangeSetID uuid.UUID, rollbackChangeSetID uuid.UUID, affected []uuid.UUID) map[string]any {
	affectedText := make([]string, 0, len(affected))
	for _, id := range affected {
		affectedText = append(affectedText, id.String())
	}
	sort.Strings(affectedText)
	return map[string]any{
		"incident_id":            incidentID.String(),
		"record_id":              recordID.String(),
		"row_version":            rowVersion,
		"target":                 target.Normalized(),
		"target_change_set_id":   targetChangeSetID.String(),
		"rollback_change_set_id": rollbackChangeSetID.String(),
		"affected_record_ids":    affectedText,
	}
}

func decodeRollbackValue(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("decode rollback value: %w", err)
	}
	return value, nil
}

func decodeStoredRollbackPayload(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode rollback idempotency payload: %w", err)
	}
	return payload, nil
}

func rollbackResultFromPayload(payload map[string]any, clientTxnID string) RollbackResult {
	result := RollbackResult{Payload: payload, StatusCode: http.StatusOK, ClientTxnID: clientTxnID, Replayed: true}
	if raw, ok := payload["incident_id"].(string); ok {
		result.IncidentID, _ = uuid.Parse(raw)
	}
	if raw, ok := payload["record_id"].(string); ok {
		recordID, _ := uuid.Parse(raw)
		changeSetID := uuid.UUID{}
		if changeSetRaw, ok := payload["rollback_change_set_id"].(string); ok {
			changeSetID, _ = uuid.Parse(changeSetRaw)
		}
		var rowVersion int64
		switch value := payload["row_version"].(type) {
		case float64:
			rowVersion = int64(value)
		case int64:
			rowVersion = value
		}
		result.Changes = []RollbackRecordChange{{RecordID: recordID, RowVersion: rowVersion, ChangeSetID: changeSetID}}
	}
	return result
}

func objectMap(parent map[string]any, key string) (map[string]any, bool) {
	if parent == nil {
		return nil, false
	}
	value, ok := parent[key].(map[string]any)
	return value, ok
}

func rollbackChangedFieldKeys(before map[string]any, after map[string]any) []string {
	beforeCells, _ := objectMap(before, "cells")
	afterCells, _ := objectMap(after, "cells")
	keys := make([]string, 0)
	seen := map[string]struct{}{}
	for key := range beforeCells {
		seen[key] = struct{}{}
	}
	for key := range afterCells {
		seen[key] = struct{}{}
	}
	for key := range seen {
		if !jsonishEqual(beforeCells[key], afterCells[key]) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func jsonishEqual(left any, right any) bool {
	leftRaw, _ := json.Marshal(left)
	rightRaw, _ := json.Marshal(right)
	return bytes.Equal(leftRaw, rightRaw)
}
