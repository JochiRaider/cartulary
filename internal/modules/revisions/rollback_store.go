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

	"github.com/JochiRaider/cartulary/internal/modules/projections"
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
}

type rollbackApplyResult struct {
	ChangeSetID uuid.UUID
	Changes     []RollbackRecordChange
}

type rollbackProtectedSet struct {
	Affected    []uuid.UUID
	DeferredErr error
}

func (s *Store) RollbackRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request RollbackRequest, requestHash []byte, requestID string, now time.Time) (RollbackResult, error) {
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

	protected, err := loadRollbackProtectedSetTx(ctx, tx, record, request.Target)
	if err != nil {
		return RollbackResult{}, err
	}
	if err := LockRecordEnvelopesNowaitTx(ctx, tx, protected.Affected); err != nil {
		return RollbackResult{}, err
	}
	record, err = loadRollbackRecordEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
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
		plan, err = loadHistoryEntryRollbackPlanTx(ctx, tx, record, request.Target.HistoryEntryRef)
	case "change_set":
		plan, err = loadChangeSetRollbackPlanTx(ctx, tx, record, request.Target.ChangeSetID)
	case "row_restore":
		plan, err = loadRowRestorePlanTx(ctx, tx, record, request.Target.RestoreToRevisionNo)
	default:
		err = &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err != nil {
		return RollbackResult{}, err
	}
	if err := validateRollbackPlan(plan); err != nil {
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

func loadRollbackProtectedSetTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, target RollbackTarget) (rollbackProtectedSet, error) {
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
		affected, err := affectedRecordsForRollbackTarget(mutation, record.RecordID)
		if err != nil {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		return rollbackProtectedSet{Affected: affected}, nil
	case "change_set":
		plan, err := loadChangeSetRollbackPlanTx(ctx, tx, record, target.ChangeSetID)
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

func loadHistoryEntryRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (rollbackPlan, error) {
	target, err := loadHistoryEntryRollbackTargetTx(ctx, tx, record, historyEntryRef)
	if err != nil {
		return rollbackPlan{}, err
	}
	affected, err := affectedRecordsForRollbackTarget(target, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan := rollbackPlan{
		Target:     target,
		Affected:   affected,
		Addressed:  record.RecordID,
		RecordType: record.RecordType,
	}
	if target.TargetKind == "entity_mention" {
		companion, err := loadRollbackMentionCompanionLinkTargetsTx(ctx, tx, target)
		if err != nil {
			return rollbackPlan{}, err
		}
		plan.Companion = companion
	}
	requiresChangeSet, err := historyEntryRequiresChangeSetTx(ctx, tx, target)
	if err != nil {
		return rollbackPlan{}, err
	}
	plan.RequiresChangeSet = requiresChangeSet
	return plan, nil
}

func loadChangeSetRollbackPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, raw string) (rollbackPlan, error) {
	changeSetID, err := uuid.Parse(raw)
	if err != nil {
		return rollbackPlan{}, ErrRollbackTargetNotFound
	}
	rows, err := tx.Query(ctx, `
SELECT csm.change_set_id,
       cs.created_at,
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
		if err := rows.Scan(&target.ChangeSetID, &target.CreatedAt, &target.SequenceNo, &target.TargetKind, &target.TargetID, &target.OperationKind, &beforeRaw, &afterRaw); err != nil {
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
	affected, err := affectedRecordsForRollbackTargets(targets, record.RecordID)
	if err != nil {
		return rollbackPlan{}, err
	}
	return rollbackPlan{
		Target:     targets[0],
		Targets:    targets,
		Affected:   affected,
		Addressed:  record.RecordID,
		RecordType: record.RecordType,
		WholeSet:   true,
	}, nil
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
	if err := row.Scan(&target.ChangeSetID, &target.CreatedAt, &target.SequenceNo, &target.TargetKind, &target.TargetID, &target.OperationKind, &beforeRaw, &afterRaw); err != nil {
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

func loadRowRestorePlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, revisionNo int64) (rollbackPlan, error) {
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
	if _, err := rollbackSourceForRecordType(record.RecordType, snapshot); err != nil {
		return rollbackPlan{}, err
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

func (s *Store) applyRollbackPlanTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, record rollbackRecordEnvelope, plan rollbackPlan, request RollbackRequest, requestID string, now time.Time) (rollbackApplyResult, error) {
	changeSetID, err := s.InsertChangeSetTx(ctx, tx, ChangeSetParams{
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
	case "record_link":
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
	default:
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return rollbackApplyResult{ChangeSetID: changeSetID, Changes: changes}, nil
}

func (s *Store) applyRowRestorePlanTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, record rollbackRecordEnvelope, plan rollbackPlan, request RollbackRequest, requestID string, now time.Time) (rollbackApplyResult, error) {
	adapter, ok := deleteRestoreAdapters[record.RecordType]
	if !ok {
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := adapter.snapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	source, err := rollbackSourceForRecordType(record.RecordType, plan.RestoreSnapshot)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, record.RecordID, actor.ID, now)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	if err := updateSourceFromRollbackSourceTx(ctx, tx, record.RecordType, record.RecordID, actor.ID, now, nextRowVersion, source); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := rebuildRollbackProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return rollbackApplyResult{}, err
	}
	afterSnapshot, err := adapter.snapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	changeSetID, err := s.InsertChangeSetTx(ctx, tx, ChangeSetParams{
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
	if err := s.InsertMutationTx(ctx, tx, MutationParams{
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
	if err := s.InsertRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    record.RecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	viewSchemaID, err := adapter.viewSchemaID(ctx, tx, record.RecordID)
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

func (s *Store) applyChangeSetRollbackPlanTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	beforeSnapshots, err := snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
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

	sourceUpdated := map[uuid.UUID]bool{}
	changedKeys := map[uuid.UUID]map[string]struct{}{}
	for i := len(plan.Targets) - 1; i >= 0; i-- {
		target := plan.Targets[i]
		switch target.TargetKind {
		case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
			recordID, keys, err := s.applyRowBackedRollbackMutationTx(ctx, tx, actor, target, changeSetID, sequenceNo, now, nextVersions)
			if err != nil {
				return nil, err
			}
			sourceUpdated[recordID] = true
			addRollbackChangedKeys(changedKeys, recordID, keys)
		case "record_link":
			affected, err := s.applyRecordLinkRollbackMutationTx(ctx, tx, actor, incidentID, target, changeSetID, sequenceNo, now)
			if err != nil {
				return nil, err
			}
			for _, recordID := range affected {
				keys, err := rollbackRecordLinkChangedFieldKeysTx(ctx, tx, target, recordID)
				if err != nil {
					return nil, err
				}
				addRollbackChangedKeys(changedKeys, recordID, keys)
			}
		case "entity_mention":
			recordID, key, err := s.applyMentionRollbackMutationTx(ctx, tx, actor, target, changeSetID, sequenceNo, now)
			if err != nil {
				return nil, err
			}
			addRollbackChangedKeys(changedKeys, recordID, []string{key})
		default:
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}

	for _, recordID := range plan.Affected {
		if sourceUpdated[recordID] {
			continue
		}
		record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		if err := touchRollbackSourceRowTx(ctx, tx, record.RecordType, recordID, actor.ID, now, nextVersions[recordID]); err != nil {
			return nil, err
		}
	}
	if err := rebuildRollbackProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}

	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := s.insertRollbackRecordRevisionSnapshotTx(ctx, tx, recordID, changeSetID, beforeSnapshots[recordID], nextVersions[recordID])
		if err != nil {
			return nil, err
		}
		change.ChangedFieldKeys = sortedRollbackChangedKeys(changedKeys[recordID])
		changes = append(changes, change)
	}
	return changes, nil
}

func (s *Store) applyRowBackedRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time, nextVersions map[uuid.UUID]int64) (uuid.UUID, []string, error) {
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
	adapter, ok := deleteRestoreAdapters[recordType]
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := adapter.snapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	source, err := rollbackSourceForRecordType(recordType, target.BeforeValue)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	nextRowVersion, ok := nextVersions[targetRecordID]
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := updateSourceFromRollbackSourceTx(ctx, tx, recordType, targetRecordID, actor.ID, now, nextRowVersion, source); err != nil {
		return uuid.UUID{}, nil, err
	}
	afterSnapshot, err := adapter.snapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, nextRowVersion)
	if target.TargetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	}
	if err := s.InsertMutationTx(ctx, tx, MutationParams{
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

func (s *Store) applyRecordLinkRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]uuid.UUID, error) {
	linkID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return nil, ErrRollbackTargetNotFound
	}
	affected, err := affectedRecordsForRollbackTarget(target, uuid.UUID{})
	if err != nil {
		return nil, err
	}
	linkBefore, err := loadRollbackRecordLinkValueTx(ctx, tx, linkID)
	if err != nil {
		return nil, err
	}
	switch target.OperationKind {
	case "create":
		if err := tombstoneRollbackRecordLinkTx(ctx, tx, incidentID, linkID, actor.ID, now); err != nil {
			return nil, err
		}
	case "delete":
		if err := restoreRollbackRecordLinkTx(ctx, tx, incidentID, linkID, target.BeforeValue, actor.ID, now); err != nil {
			return nil, err
		}
	default:
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	linkAfter, err := loadRollbackRecordLinkValueTx(ctx, tx, linkID)
	if err != nil {
		return nil, err
	}
	if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, linkBefore, linkAfter); err != nil {
		return nil, err
	}
	return affected, nil
}

func (s *Store) applyMentionRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time) (uuid.UUID, string, error) {
	mentionID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return uuid.UUID{}, "", ErrRollbackTargetNotFound
	}
	sourceRecordID, ok := stringFromMap(target.BeforeValue, "source_record_id")
	if !ok {
		return uuid.UUID{}, "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	sourceID, err := uuid.Parse(sourceRecordID)
	if err != nil {
		return uuid.UUID{}, "", ErrRollbackTargetNotFound
	}
	mentionBefore, err := loadRollbackMentionValueTx(ctx, tx, mentionID)
	if err != nil {
		return uuid.UUID{}, "", err
	}
	if err := restoreRollbackMentionTx(ctx, tx, mentionID, target.BeforeValue, actor.ID, now); err != nil {
		return uuid.UUID{}, "", err
	}
	mentionAfter, err := loadRollbackMentionValueTx(ctx, tx, mentionID)
	if err != nil {
		return uuid.UUID{}, "", err
	}
	if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, mentionBefore, mentionAfter); err != nil {
		return uuid.UUID{}, "", err
	}
	return sourceID, rollbackMentionFieldKey(target.BeforeValue), nil
}

func (s *Store) insertRollbackRecordRevisionSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changeSetID uuid.UUID, beforeSnapshot map[string]any, rowVersion int64) (RollbackRecordChange, error) {
	record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	adapter, ok := deleteRestoreAdapters[record.RecordType]
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	afterSnapshot, err := adapter.snapshotTx(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := s.InsertRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := adapter.viewSchemaID(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{RecordID: recordID, RowVersion: rowVersion, ChangeSetID: changeSetID, ViewSchemaID: viewSchemaID}, nil
}

func (s *Store) applyRowBackedRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) (RollbackRecordChange, error) {
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
	adapter, ok := deleteRestoreAdapters[recordType]
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := adapter.snapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	source, err := rollbackSourceForRecordType(recordType, target.BeforeValue)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, targetRecordID, actor.ID, now)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := updateSourceFromRollbackSourceTx(ctx, tx, recordType, targetRecordID, actor.ID, now, nextRowVersion, source); err != nil {
		return RollbackRecordChange{}, err
	}
	if err := rebuildRollbackProjectionsTx(ctx, tx, targetRecord.IncidentID); err != nil {
		return RollbackRecordChange{}, err
	}
	afterSnapshot, err := adapter.snapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, nextRowVersion)
	if target.TargetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	}
	if err := s.InsertMutationTx(ctx, tx, MutationParams{
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
	if err := s.InsertRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    targetRecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := adapter.viewSchemaID(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         targetRecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(target.BeforeValue, target.AfterValue),
	}, nil
}

func (s *Store) insertRollbackRecordRevisionForAffectedTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, recordID uuid.UUID, changeSetID uuid.UUID, now time.Time, beforeSnapshot map[string]any) (RollbackRecordChange, error) {
	record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	adapter, ok := deleteRestoreAdapters[record.RecordType]
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	nextRowVersion, err := updateRollbackRecordEnvelopeTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := touchRollbackSourceRowTx(ctx, tx, record.RecordType, recordID, actor.ID, now, nextRowVersion); err != nil {
		return RollbackRecordChange{}, err
	}
	afterSnapshot, err := adapter.snapshotTx(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := s.InsertRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := adapter.viewSchemaID(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{RecordID: recordID, RowVersion: nextRowVersion, ChangeSetID: changeSetID, ViewSchemaID: viewSchemaID}, nil
}

func (s *Store) insertRollbackMutationTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, sequenceNo *int, target rollbackMutationTarget, beforeValue any, afterValue any) error {
	if err := s.InsertMutationTx(ctx, tx, MutationParams{
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

func (s *Store) applyRecordLinkRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	target := plan.Target
	linkID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return nil, ErrRollbackTargetNotFound
	}
	beforeSnapshots, err := snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	linkBefore, err := loadRollbackRecordLinkValueTx(ctx, tx, linkID)
	if err != nil {
		return nil, err
	}
	switch target.OperationKind {
	case "create":
		if err := tombstoneRollbackRecordLinkTx(ctx, tx, incidentID, linkID, actor.ID, now); err != nil {
			return nil, err
		}
	case "delete":
		if err := restoreRollbackRecordLinkTx(ctx, tx, incidentID, linkID, target.BeforeValue, actor.ID, now); err != nil {
			return nil, err
		}
	default:
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := rebuildRollbackProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	linkAfter, err := loadRollbackRecordLinkValueTx(ctx, tx, linkID)
	if err != nil {
		return nil, err
	}
	if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, linkBefore, linkAfter); err != nil {
		return nil, err
	}
	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := s.insertRollbackRecordRevisionForAffectedTx(ctx, tx, actor, recordID, changeSetID, now, beforeSnapshots[recordID])
		if err != nil {
			return nil, err
		}
		keys, err := rollbackRecordLinkChangedFieldKeysTx(ctx, tx, target, recordID)
		if err != nil {
			return nil, err
		}
		change.ChangedFieldKeys = keys
		changes = append(changes, change)
	}
	return changes, nil
}

func (s *Store) applyMentionRollbackTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	target := plan.Target
	mentionID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return nil, ErrRollbackTargetNotFound
	}
	sourceRecordID, ok := stringFromMap(target.BeforeValue, "source_record_id")
	if !ok {
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	sourceID, err := uuid.Parse(sourceRecordID)
	if err != nil {
		return nil, ErrRollbackTargetNotFound
	}
	beforeSnapshots, err := snapshotRollbackAffectedRecordsTx(ctx, tx, []uuid.UUID{sourceID})
	if err != nil {
		return nil, err
	}
	mentionBefore, err := loadRollbackMentionValueTx(ctx, tx, mentionID)
	if err != nil {
		return nil, err
	}
	if err := restoreRollbackMentionTx(ctx, tx, mentionID, target.BeforeValue, actor.ID, now); err != nil {
		return nil, err
	}
	for _, companion := range plan.Companion {
		companionID, err := uuid.Parse(companion.TargetID)
		if err != nil {
			return nil, ErrRollbackTargetNotFound
		}
		switch companion.OperationKind {
		case "create":
			if err := tombstoneRollbackRecordLinkTx(ctx, tx, incidentID, companionID, actor.ID, now); err != nil {
				return nil, err
			}
		case "delete":
			if err := restoreRollbackRecordLinkTx(ctx, tx, incidentID, companionID, companion.BeforeValue, actor.ID, now); err != nil {
				return nil, err
			}
		default:
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	if err := rebuildRollbackProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	mentionAfter, err := loadRollbackMentionValueTx(ctx, tx, mentionID)
	if err != nil {
		return nil, err
	}
	if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, mentionBefore, mentionAfter); err != nil {
		return nil, err
	}
	for _, companion := range plan.Companion {
		linkID, err := uuid.Parse(companion.TargetID)
		if err != nil {
			return nil, ErrRollbackTargetNotFound
		}
		linkAfter, err := loadRollbackRecordLinkValueTx(ctx, tx, linkID)
		if err != nil {
			return nil, err
		}
		if err := s.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, companion, companion.AfterValue, linkAfter); err != nil {
			return nil, err
		}
	}
	change, err := s.insertRollbackRecordRevisionForAffectedTx(ctx, tx, actor, sourceID, changeSetID, now, beforeSnapshots[sourceID])
	if err != nil {
		return nil, err
	}
	change.ChangedFieldKeys = []string{rollbackMentionFieldKey(target.BeforeValue)}
	return []RollbackRecordChange{change}, nil
}

func snapshotRollbackAffectedRecordsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	snapshots := make(map[uuid.UUID]map[string]any, len(recordIDs))
	for _, recordID := range recordIDs {
		record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		adapter, ok := deleteRestoreAdapters[record.RecordType]
		if !ok {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		snapshot, err := adapter.snapshotTx(ctx, tx, recordID)
		if err != nil {
			return nil, err
		}
		snapshots[recordID] = snapshot
	}
	return snapshots, nil
}

func validateRollbackPlan(plan rollbackPlan) error {
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
			if err := validateRollbackTarget(target); err != nil {
				return err
			}
		}
		return nil
	}
	return validateRollbackTarget(plan.Target)
}

func validateRollbackTarget(target rollbackMutationTarget) error {
	switch target.TargetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		if target.OperationKind != "patch" && target.OperationKind != "field_update" && target.OperationKind != "hostname_update" && target.OperationKind != "state_update" {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		if target.BeforeValue == nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	case "record_link":
		if target.OperationKind != "create" && target.OperationKind != "delete" {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		value := target.BeforeValue
		if target.OperationKind == "create" {
			value = target.AfterValue
		}
		if _, err := rollbackRecordLinkIdentity(value); err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	case "entity_mention":
		if target.OperationKind != "patch" || target.BeforeValue == nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		for _, key := range []string{"source_record_id", "source_field_key", "resolution_status"} {
			if _, ok := stringFromMap(target.BeforeValue, key); !ok {
				return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
			}
		}
	case "record_tag":
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	default:
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return nil
}

func rollbackRecordTypeForTarget(target rollbackMutationTarget, currentRecordType string) (string, error) {
	switch target.TargetKind {
	case "record":
		if currentRecordType == "" {
			return "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		return currentRecordType, nil
	case "timeline_record":
		return "timeline_event", nil
	case "host", "identity", "indicator", "assessment", "evidence":
		return target.TargetKind, nil
	default:
		return "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
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

func updateHostFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) error {
	displayName, ok := stringFromMap(source, "display_name")
	if !ok || displayName == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	hostState, ok := stringFromMap(source, "host_state")
	if !ok || hostState == "" {
		hostState = "canonical"
	}
	mergedInto, err := uuidPointerFromMap(source, "merged_into_record_id")
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE hosts
   SET display_name = $2,
       hostname = $3,
       aad_device_id = $4,
       fqdn = $5,
       host_state = $6,
       merged_into_record_id = $7,
       row_version = $8,
       updated_at = $9,
       updated_by_user_id = $10
 WHERE record_id = $1
`, recordID, displayName, nullableStringAny(source, "hostname"), nullableStringAny(source, "aad_device_id"), nullableStringAny(source, "fqdn"), hostState, mergedInto, rowVersion, now.UTC(), actorUserID)
	if err != nil {
		return fmt.Errorf("update host rollback source: %w", err)
	}
	return nil
}

func updateSourceFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordType string, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) error {
	switch recordType {
	case "host":
		return updateHostFromRollbackSourceTx(ctx, tx, recordID, actorUserID, now, rowVersion, source)
	case "identity":
		return updateIdentityFromRollbackSourceTx(ctx, tx, recordID, actorUserID, now, rowVersion, source)
	case "timeline_event":
		return updateTimelineFromRollbackSourceTx(ctx, tx, recordID, actorUserID, now, rowVersion, source)
	case "indicator":
		return updateIndicatorFromRollbackSourceTx(ctx, tx, recordID, actorUserID, now, rowVersion, source)
	case "evidence":
		return updateEvidenceFromRollbackSourceTx(ctx, tx, recordID, now, source)
	case "party":
		return updateGenericWorkbookSourceTx(ctx, tx, "parties", recordID, now, source, []string{"display_name", "party_kind", "organization_name", "role_title", "primary_email", "timezone_name", "external_ref", "notes", "updated_at"})
	case "task_request":
		return updateGenericWorkbookSourceTx(ctx, tx, "task_requests", recordID, now, source, []string{"title", "status", "owner_user_id", "priority", "task_kind", "workstream", "due_at", "requester_party_text", "requester_party_id", "blocked_reason", "completed_at", "external_ticket_ref", "closure_summary", "decision_record_id", "updated_at"})
	case "decision":
		return updateGenericWorkbookSourceTx(ctx, tx, "decisions", recordID, now, source, []string{"summary", "status", "owner_user_id", "decision_type", "decided_at", "rationale", "supersedes_record_id", "updated_at"})
	case "artifact":
		return updateGenericWorkbookSourceTx(ctx, tx, "artifacts", recordID, now, source, []string{"title", "body", "timestamp_utc", "updated_at", "comm_id", "comm_type", "audience", "channel_or_meeting", "summary", "next_report_at", "privilege_tag", "handoff_id", "outgoing_owner_user_id", "incoming_owner_user_id", "current_state_summary", "next_checks", "acknowledged_at", "status_review_id", "review_owner_user_id", "active_risks_summary", "lesson_id", "owner_user_id", "closure_state", "created_by_user_id"})
	case "assessment":
		return updateAssessmentFromRollbackSourceTx(ctx, tx, recordID, now, source)
	default:
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
}

func touchRollbackSourceRowTx(ctx context.Context, tx pgx.Tx, recordType string, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64) error {
	switch recordType {
	case "timeline_event":
		_, err := tx.Exec(ctx, `UPDATE timeline_events SET row_version = $2, edited_at = $3, updated_by_user_id = $4 WHERE record_id = $1`, recordID, rowVersion, now.UTC(), actorUserID)
		return err
	case "host":
		_, err := tx.Exec(ctx, `UPDATE hosts SET row_version = $2, updated_at = $3, updated_by_user_id = $4 WHERE record_id = $1`, recordID, rowVersion, now.UTC(), actorUserID)
		return err
	case "identity":
		_, err := tx.Exec(ctx, `UPDATE identities SET row_version = $2, updated_at = $3, updated_by_user_id = $4 WHERE record_id = $1`, recordID, rowVersion, now.UTC(), actorUserID)
		return err
	case "indicator":
		_, err := tx.Exec(ctx, `UPDATE indicators SET row_version = $2, updated_at = $3, updated_by_user_id = $4 WHERE record_id = $1`, recordID, rowVersion, now.UTC(), actorUserID)
		return err
	default:
		adapter, ok := deleteRestoreAdapters[recordType]
		if !ok {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		_, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET updated_at = $2 WHERE %s = $1`, adapter.SourceTable, adapter.SourceRecordCol), recordID, now.UTC())
		return err
	}
}

func updateIdentityFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) error {
	displayName, ok := stringFromMap(source, "display_name")
	if !ok || displayName == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	state, ok := stringFromMap(source, "identity_state")
	if !ok || state == "" {
		state = "canonical"
	}
	mergedInto, err := uuidPointerFromMap(source, "merged_into_record_id")
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE identities
   SET display_name = $2,
       upn = $3,
       email = $4,
       sam_account_name = $5,
       aad_object_id = $6,
       sid = $7,
       identity_state = $8,
       merged_into_record_id = $9,
       row_version = $10,
       updated_at = $11,
       updated_by_user_id = $12
 WHERE record_id = $1
`, recordID, displayName, nullableStringAny(source, "upn"), nullableStringAny(source, "email"), nullableStringAny(source, "sam_account_name"), nullableStringAny(source, "aad_object_id"), nullableStringAny(source, "sid"), state, mergedInto, rowVersion, now.UTC(), actorUserID)
	return err
}

func updateTimelineFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) error {
	captureState, ok := stringFromMap(source, "capture_state")
	if !ok || captureState == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	_, err := tx.Exec(ctx, `
UPDATE timeline_events
   SET occurred_at = $2,
       summary = $3,
       details = $4,
       source_text = $5,
       capture_state = $6,
       row_version = $7,
       edited_at = $8,
       updated_by_user_id = $9,
       reviewed_by_user_id = $10,
       reviewed_at = $11,
       superseded_by_user_id = $12,
       superseded_at = $13
 WHERE record_id = $1
`, recordID, nullableAny(source, "occurred_at"), nullableStringAny(source, "summary"), nullableStringAny(source, "details"), nullableStringAny(source, "source_text"), captureState, rowVersion, now.UTC(), actorUserID, nullableUUIDAny(source, "reviewed_by_user_id"), nullableAny(source, "reviewed_at"), nullableUUIDAny(source, "superseded_by_user_id"), nullableAny(source, "superseded_at"))
	return err
}

func updateIndicatorFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) error {
	indicatorType, ok := stringFromMap(source, "indicator_type")
	if !ok || indicatorType == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	valueKind, ok := stringFromMap(source, "value_kind")
	if !ok || valueKind == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	displayValue, ok := stringFromMap(source, "display_value")
	if !ok || displayValue == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	dedupeKey, ok := stringFromMap(source, "dedupe_key")
	if !ok || dedupeKey == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	_, err := tx.Exec(ctx, `
UPDATE indicators
   SET indicator_type = $2,
       value_kind = $3,
       display_value = $4,
       normalized_value = $5,
       dedupe_key = $6,
       defanged_value = $7,
       hash_algorithm = $8,
       hash_value = $9,
       stix_pattern = $10,
       row_version = $11,
       updated_at = $12,
       updated_by_user_id = $13,
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_id = $1
`, recordID, indicatorType, valueKind, displayValue, nullableStringAny(source, "normalized_value"), dedupeKey, nullableStringAny(source, "defanged_value"), nullableStringAny(source, "hash_algorithm"), nullableStringAny(source, "hash_value"), nullableStringAny(source, "stix_pattern"), rowVersion, now.UTC(), actorUserID)
	return err
}

func updateEvidenceFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time, source map[string]any) error {
	lifecycle, ok := stringFromMap(source, "lifecycle_state")
	if !ok || lifecycle == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	_, err := tx.Exec(ctx, `
UPDATE evidence
   SET title = $2,
       lifecycle_state = $3,
       requested_at = $4,
       received_at = $5,
       storage_ref = $6,
       blob_hash = $7,
       collector_party_text = $8,
       collector_party_id = $9,
       source_party_text = $10,
       source_party_id = $11,
       upload_state = $12,
       object_blob_id = $13,
       updated_at = $14
 WHERE record_id = $1
`, recordID, nullableStringAny(source, "title"), lifecycle, nullableAny(source, "requested_at"), nullableAny(source, "received_at"), nullableStringAny(source, "storage_ref"), nullableStringAny(source, "blob_hash"), nullableStringAny(source, "collector_party_text"), nullableUUIDAny(source, "collector_party_id"), nullableStringAny(source, "source_party_text"), nullableUUIDAny(source, "source_party_id"), stringDefault(source, "upload_state", "pending"), nullableUUIDAny(source, "object_blob_id"), now.UTC())
	return err
}

func updateAssessmentFromRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time, source map[string]any) error {
	state, ok := stringFromMap(source, "assessment_state")
	if !ok || state == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	subjectType, ok := stringFromMap(source, "subject_type")
	if !ok || subjectType == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	_, err := tx.Exec(ctx, `
UPDATE assessments
   SET subject_record_id = $2,
       subject_type = $3,
       assessment_state = $4,
       confidence_score = $5,
       rationale = $6,
       assessor_user_id = $7,
       assessed_at = $8,
       updated_at = $9,
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_id = $1
`, recordID, nullableUUIDAny(source, "subject_record_id"), subjectType, state, nullableAny(source, "confidence_score"), stringDefault(source, "rationale", ""), nullableUUIDAny(source, "assessor_user_id"), nullableAny(source, "assessed_at"), now.UTC())
	return err
}

func updateGenericWorkbookSourceTx(ctx context.Context, tx pgx.Tx, table string, recordID uuid.UUID, now time.Time, source map[string]any, columns []string) error {
	assignments := make([]string, 0, len(columns))
	args := []any{recordID}
	for _, column := range columns {
		value := nullableAny(source, column)
		if column == "updated_at" {
			value = now.UTC()
		}
		args = append(args, value)
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if len(assignments) == 0 {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	_, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET %s WHERE record_id = $1`, table, joinSQLAssignments(assignments)), args...)
	return err
}

func rebuildRollbackProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	store := projections.NewStore(nil)
	if err := store.RebuildIncidentTimelineTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := store.RebuildIncidentHostsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := store.RebuildIncidentIdentitiesTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := store.RebuildIncidentIndicatorsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := store.RebuildIncidentAssessmentsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	return nil
}

func loadRollbackMentionCompanionLinkTargetsTx(ctx context.Context, tx pgx.Tx, target rollbackMutationTarget) ([]rollbackMutationTarget, error) {
	rows, err := tx.Query(ctx, `
SELECT csm.change_set_id,
       cs.created_at,
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
		if err := rows.Scan(&companion.ChangeSetID, &companion.CreatedAt, &companion.SequenceNo, &companion.TargetKind, &companion.TargetID, &companion.OperationKind, &beforeRaw, &afterRaw); err != nil {
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

func historyEntryRequiresChangeSetTx(ctx context.Context, tx pgx.Tx, target rollbackMutationTarget) (bool, error) {
	switch target.TargetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
	default:
		return false, nil
	}
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM change_set_mutations csm
     WHERE csm.change_set_id = $1
       AND csm.sequence_no <> $2
       AND csm.target_kind = 'record_link'
       AND COALESCE(csm.before_value ->> 'link_type', csm.after_value ->> 'link_type') = 'attached_evidence'
)
`, target.ChangeSetID, target.SequenceNo).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func loadRollbackRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	var raw []byte
	err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'record_link_id', record_link_id::text,
    'incident_id', incident_id::text,
    'src_record_id', src_record_id::text,
    'dst_record_id', dst_record_id::text,
    'link_type', link_type,
    'field_key', field_key,
    'provenance', provenance,
    'confidence', confidence,
    'owner_user_id', owner_user_id::text,
    'created_by_user_id', created_by_user_id::text,
    'decided_at', decided_at,
    'created_at', created_at,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM record_links
 WHERE record_link_id = $1
`, recordLinkID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRollbackTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func tombstoneRollbackRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $3,
       deleted_by_user_id = $4
 WHERE record_link_id = $1
   AND incident_id = $2
   AND deleted_at IS NULL
`, recordLinkID, incidentID, now.UTC(), actorUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return &RollbackPreconditionError{ReasonCode: "stale_target"}
	}
	return nil
}

func restoreRollbackRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, value map[string]any, actorUserID uuid.UUID, now time.Time) error {
	identity, err := rollbackRecordLinkIdentity(value)
	if err != nil {
		return err
	}
	if identity.IncidentID != incidentID || identity.RecordLinkID != recordLinkID {
		return ErrRollbackTargetNotFound
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET src_record_id = $3,
       dst_record_id = $4,
       link_type = $5,
       field_key = $6,
       provenance = $7,
       confidence = $8,
       owner_user_id = $9,
       decided_at = COALESCE($10, decided_at),
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_link_id = $1
   AND incident_id = $2
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, nullableStringAny(value, "field_key"), stringDefault(value, "provenance", "rollback"), nullableAny(value, "confidence"), uuidAnyDefault(value, "owner_user_id", actorUserID), nullableAny(value, "decided_at"))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	_, err = tx.Exec(ctx, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, COALESCE($10, $11), $11)
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, nullableStringAny(value, "field_key"), stringDefault(value, "provenance", "rollback"), nullableAny(value, "confidence"), uuidAnyDefault(value, "owner_user_id", actorUserID), nullableAny(value, "decided_at"), now.UTC())
	return err
}

type rollbackLinkIdentity struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
}

func rollbackRecordLinkIdentity(value map[string]any) (rollbackLinkIdentity, error) {
	var identity rollbackLinkIdentity
	var err error
	if identity.RecordLinkID, err = uuidFromMap(value, "record_link_id"); err != nil {
		return identity, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if identity.IncidentID, err = uuidFromMap(value, "incident_id"); err != nil {
		return identity, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if identity.SrcRecordID, err = uuidFromMap(value, "src_record_id"); err != nil {
		return identity, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if identity.DstRecordID, err = uuidFromMap(value, "dst_record_id"); err != nil {
		return identity, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	linkType, ok := stringFromMap(value, "link_type")
	if !ok || linkType == "" {
		return identity, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	identity.LinkType = linkType
	return identity, nil
}

func loadRollbackMentionValueTx(ctx context.Context, tx pgx.Tx, mentionID uuid.UUID) (map[string]any, error) {
	var raw []byte
	err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'entity_mention_id', entity_mention_id::text,
    'source_record_id', source_record_id::text,
    'entity_type', entity_type,
    'source_field_key', source_field_key,
    'origin_kind', origin_kind,
    'origin_locator', origin_locator,
    'raw_text', raw_text,
    'normalized_text', normalized_text,
    'resolution_status', resolution_status,
    'row_version', row_version,
    'ordinal', ordinal,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'resolved_record_id', resolved_record_id::text,
    'resolved_by_user_id', resolved_by_user_id::text,
    'resolved_at', resolved_at,
    'resolution_method', resolution_method
)
  FROM entity_mentions
 WHERE entity_mention_id = $1
`, mentionID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRollbackTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func restoreRollbackMentionTx(ctx context.Context, tx pgx.Tx, mentionID uuid.UUID, value map[string]any, actorUserID uuid.UUID, now time.Time) error {
	sourceID, err := uuidFromMap(value, "source_record_id")
	if err != nil {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	status, ok := stringFromMap(value, "resolution_status")
	if !ok || status == "" {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	rowVersion := int64FromRollbackAny(value["row_version"])
	if rowVersion <= 0 {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	tag, err := tx.Exec(ctx, `
UPDATE entity_mentions
   SET source_record_id = $2,
       entity_type = $3,
       source_field_key = $4,
       origin_kind = $5,
       origin_locator = $6,
       raw_text = $7,
       normalized_text = $8,
       resolution_status = $9,
       row_version = row_version + 1,
       resolved_record_id = $10,
       resolved_by_user_id = $11,
       resolved_at = $12,
       resolution_method = $13
 WHERE entity_mention_id = $1
`, mentionID, sourceID, stringDefault(value, "entity_type", "host"), stringDefault(value, "source_field_key", ""), stringDefault(value, "origin_kind", "rollback"), stringDefault(value, "origin_locator", mentionID.String()), stringDefault(value, "raw_text", ""), stringDefault(value, "normalized_text", ""), status, nullableAny(value, "resolved_record_id"), nullableAny(value, "resolved_by_user_id"), nullableAny(value, "resolved_at"), nullableStringAny(value, "resolution_method"))
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrRollbackTargetNotFound
	}
	_ = actorUserID
	_ = now
	return nil
}

func affectedRecordsForRollbackTarget(target rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	recordIDs := map[uuid.UUID]struct{}{}
	add := func(value string) error {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return err
		}
		recordIDs[parsed] = struct{}{}
		return nil
	}
	switch target.TargetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		if err := add(target.TargetID); err != nil {
			return nil, ErrRollbackTargetNotFound
		}
	case "entity_mention":
		if source, ok := stringFromMap(target.BeforeValue, "source_record_id"); ok {
			if err := add(source); err != nil {
				return nil, ErrRollbackTargetNotFound
			}
		} else if source, ok := stringFromMap(target.AfterValue, "source_record_id"); ok {
			if err := add(source); err != nil {
				return nil, ErrRollbackTargetNotFound
			}
		} else {
			recordIDs[fallback] = struct{}{}
		}
	case "record_link":
		value := target.BeforeValue
		if target.OperationKind == "create" {
			value = target.AfterValue
		}
		for _, key := range []string{"src_record_id", "dst_record_id"} {
			if value, ok := stringFromMap(value, key); ok {
				if err := add(value); err != nil {
					return nil, ErrRollbackTargetNotFound
				}
			}
		}
		if len(recordIDs) == 0 {
			recordIDs[fallback] = struct{}{}
		}
	default:
		recordIDs[fallback] = struct{}{}
	}
	affected := make([]uuid.UUID, 0, len(recordIDs))
	for recordID := range recordIDs {
		affected = append(affected, recordID)
	}
	sort.Slice(affected, func(i, j int) bool { return affected[i].String() < affected[j].String() })
	return affected, nil
}

func affectedRecordsForRollbackTargets(targets []rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	recordIDs := map[uuid.UUID]struct{}{fallback: {}}
	for _, target := range targets {
		affected, err := affectedRecordsForRollbackTarget(target, fallback)
		if err != nil {
			recordIDs[fallback] = struct{}{}
			continue
		}
		for _, recordID := range affected {
			if recordID == uuid.Nil {
				continue
			}
			recordIDs[recordID] = struct{}{}
		}
		for _, value := range []map[string]any{target.BeforeValue, target.AfterValue} {
			for _, key := range []string{"record_id", "source_record_id", "src_record_id", "dst_record_id"} {
				text, ok := stringFromMap(value, key)
				if !ok {
					continue
				}
				parsed, err := uuid.Parse(text)
				if err == nil {
					recordIDs[parsed] = struct{}{}
				}
			}
		}
		if parsed, err := uuid.Parse(target.TargetID); err == nil && firstClassRollbackTargetKind(target.TargetKind) {
			recordIDs[parsed] = struct{}{}
		}
	}
	delete(recordIDs, uuid.Nil)
	affected := make([]uuid.UUID, 0, len(recordIDs))
	for recordID := range recordIDs {
		affected = append(affected, recordID)
	}
	sort.Slice(affected, func(i, j int) bool { return affected[i].String() < affected[j].String() })
	return affected, nil
}

func firstClassRollbackTargetKind(targetKind string) bool {
	switch targetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		return true
	default:
		return false
	}
}

func addRollbackChangedKeys(changed map[uuid.UUID]map[string]struct{}, recordID uuid.UUID, keys []string) {
	if recordID == uuid.Nil {
		return
	}
	if _, ok := changed[recordID]; !ok {
		changed[recordID] = map[string]struct{}{}
	}
	for _, key := range keys {
		if key == "" {
			continue
		}
		changed[recordID][key] = struct{}{}
	}
}

func sortedRollbackChangedKeys(keys map[string]struct{}) []string {
	if len(keys) == 0 {
		return nil
	}
	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
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

func stringFromMap(values map[string]any, key string) (string, bool) {
	if values == nil {
		return "", false
	}
	value, ok := values[key]
	if !ok || value == nil {
		return "", false
	}
	text, ok := value.(string)
	return text, ok
}

func nullableStringAny(values map[string]any, key string) any {
	if value, ok := stringFromMap(values, key); ok {
		return value
	}
	return nil
}

func uuidPointerFromMap(values map[string]any, key string) (*uuid.UUID, error) {
	value, ok := stringFromMap(values, key)
	if !ok || value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func uuidFromMap(values map[string]any, key string) (uuid.UUID, error) {
	value, ok := stringFromMap(values, key)
	if !ok || value == "" {
		return uuid.UUID{}, fmt.Errorf("missing uuid field %s", key)
	}
	return uuid.Parse(value)
}

func nullableUUIDAny(values map[string]any, key string) any {
	parsed, err := uuidPointerFromMap(values, key)
	if err != nil || parsed == nil {
		return nil
	}
	return *parsed
}

func uuidAnyDefault(values map[string]any, key string, fallback uuid.UUID) any {
	parsed, err := uuidPointerFromMap(values, key)
	if err != nil || parsed == nil {
		return fallback
	}
	return *parsed
}

func nullableAny(values map[string]any, key string) any {
	if values == nil {
		return nil
	}
	value, ok := values[key]
	if !ok {
		return nil
	}
	return value
}

func stringDefault(values map[string]any, key string, fallback string) string {
	if value, ok := stringFromMap(values, key); ok && value != "" {
		return value
	}
	return fallback
}

func int64FromRollbackAny(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	default:
		return 0
	}
}

func joinSQLAssignments(assignments []string) string {
	if len(assignments) == 0 {
		return ""
	}
	out := assignments[0]
	for _, assignment := range assignments[1:] {
		out += ", " + assignment
	}
	return out
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

func rollbackMentionFieldKey(value map[string]any) string {
	if fieldKey, ok := stringFromMap(value, "source_field_key"); ok {
		return fieldKey
	}
	return "timeline.has_unresolved_mentions"
}

func rollbackRecordLinkChangedFieldKeysTx(ctx context.Context, tx pgx.Tx, target rollbackMutationTarget, recordID uuid.UUID) ([]string, error) {
	value := target.BeforeValue
	if target.OperationKind == "create" {
		value = target.AfterValue
	}
	linkType, ok := stringFromMap(value, "link_type")
	if !ok || linkType != "attached_evidence" {
		return nil, nil
	}
	srcText, _ := stringFromMap(value, "src_record_id")
	dstText, _ := stringFromMap(value, "dst_record_id")
	record, err := loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return nil, err
	}
	switch {
	case srcText == recordID.String():
		switch record.RecordType {
		case "timeline_event":
			return []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"}, nil
		case "host":
			return []string{"host.evidence_count"}, nil
		case "identity":
			return []string{"identity.evidence_count"}, nil
		}
	case dstText == recordID.String() && record.RecordType == "evidence":
		return []string{"evidence.linked_record_count"}, nil
	}
	return nil, nil
}

func rollbackSourceForRecordType(recordType string, value map[string]any) (map[string]any, error) {
	if source, ok := objectMap(value, "source"); ok {
		return source, nil
	}
	cells, ok := objectMap(value, "cells")
	if !ok {
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	source := map[string]any{}
	var mapping map[string]string
	switch recordType {
	case "timeline_event":
		mapping = map[string]string{
			"timeline.occurred_at":           "occurred_at",
			"timeline.summary":               "summary",
			"timeline.details":               "details",
			"timeline.source_text":           "source_text",
			"timeline.capture_state":         "capture_state",
			"timeline.replacement_record_id": "replacement_record_id",
			"timeline.reviewed_at":           "reviewed_at",
			"timeline.superseded_at":         "superseded_at",
		}
		source["capture_state"] = "rough"
	case "host":
		mapping = map[string]string{
			"host.display_name":  "display_name",
			"host.hostname":      "hostname",
			"host.aad_device_id": "aad_device_id",
			"host.fqdn":          "fqdn",
			"host.host_state":    "host_state",
		}
		source["host_state"] = "canonical"
	case "identity":
		mapping = map[string]string{
			"identity.display_name":     "display_name",
			"identity.upn":              "upn",
			"identity.email":            "email",
			"identity.sam_account_name": "sam_account_name",
			"identity.aad_object_id":    "aad_object_id",
			"identity.sid":              "sid",
			"identity.identity_state":   "identity_state",
		}
		source["identity_state"] = "canonical"
	case "indicator":
		mapping = map[string]string{
			"indicator.indicator_type":   "indicator_type",
			"indicator.value_kind":       "value_kind",
			"indicator.display_value":    "display_value",
			"indicator.normalized_value": "normalized_value",
			"indicator.defanged_value":   "defanged_value",
			"indicator.hash_algorithm":   "hash_algorithm",
			"indicator.hash_value":       "hash_value",
			"indicator.stix_pattern":     "stix_pattern",
			"indicator.dedupe_key":       "dedupe_key",
		}
	case "evidence":
		mapping = map[string]string{
			"evidence.title":                "title",
			"evidence.lifecycle_state":      "lifecycle_state",
			"evidence.requested_at":         "requested_at",
			"evidence.received_at":          "received_at",
			"evidence.storage_ref":          "storage_ref",
			"evidence.blob_hash":            "blob_hash",
			"evidence.collector_party_text": "collector_party_text",
			"evidence.collector_party_id":   "collector_party_id",
			"evidence.source_party_text":    "source_party_text",
			"evidence.source_party_id":      "source_party_id",
			"evidence.upload_state":         "upload_state",
		}
		source["lifecycle_state"] = "requested"
		source["upload_state"] = "pending"
	case "assessment":
		mapping = map[string]string{
			"assessment.subject_ref":      "subject_record_id",
			"assessment.subject_type":     "subject_type",
			"assessment.assessment_state": "assessment_state",
			"assessment.confidence_score": "confidence_score",
			"assessment.rationale":        "rationale",
			"assessment.assessor":         "assessor_user_id",
			"assessment.assessed_at":      "assessed_at",
		}
	case "party":
		mapping = map[string]string{
			"party.display_name":      "display_name",
			"party.party_kind":        "party_kind",
			"party.organization_name": "organization_name",
			"party.role_title":        "role_title",
			"party.primary_email":     "primary_email",
			"party.timezone_name":     "timezone_name",
			"party.external_ref":      "external_ref",
			"party.notes":             "notes",
		}
	case "task_request":
		mapping = map[string]string{
			"task.title":                "title",
			"task.status":               "status",
			"task.owner_user_id":        "owner_user_id",
			"task.priority":             "priority",
			"task.task_kind":            "task_kind",
			"task.workstream":           "workstream",
			"task.due_at":               "due_at",
			"task.requester_party_text": "requester_party_text",
			"task.requester_party_id":   "requester_party_id",
			"task.blocked_reason":       "blocked_reason",
			"task.completed_at":         "completed_at",
			"task.external_ticket_ref":  "external_ticket_ref",
			"task.closure_summary":      "closure_summary",
			"task.decision_record_id":   "decision_record_id",
		}
	case "decision":
		mapping = map[string]string{
			"decision.summary":              "summary",
			"decision.status":               "status",
			"decision.owner_user_id":        "owner_user_id",
			"decision.decision_type":        "decision_type",
			"decision.decided_at":           "decided_at",
			"decision.rationale":            "rationale",
			"decision.supersedes_record_id": "supersedes_record_id",
		}
	case "artifact":
		mapping = map[string]string{
			"note.title":                          "title",
			"note.body":                           "body",
			"comm_log.summary":                    "summary",
			"comm_log.comm_type":                  "comm_type",
			"comm_log.audience":                   "audience",
			"comm_log.channel_or_meeting":         "channel_or_meeting",
			"handoff.current_state_summary":       "current_state_summary",
			"status_review.current_state_summary": "current_state_summary",
			"lesson.summary":                      "summary",
			"finding.title":                       "title",
		}
	default:
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	for fieldKey, sourceKey := range mapping {
		cell, ok := objectMap(cells, fieldKey)
		if !ok {
			continue
		}
		source[sourceKey] = cell["value"]
	}
	if len(source) == 0 {
		return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return source, nil
}
