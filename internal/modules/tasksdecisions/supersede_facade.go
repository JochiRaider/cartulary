package tasksdecisions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const decisionsViewSchemaID = "cartulary.view.decisions.v1"

type SupersedeFacade struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	rowProjector   *projectionadapters.RowProjector
	recordStore    *records.Store
	revisionStore  revisionAppendPort
	taskStore      *Store
}

type SupersedeRequest struct {
	BaseRowVersion      int64
	ClientTxnID         string
	Reason              string
	ReplacementRecordID *uuid.UUID
}

type SupersedeCommand struct {
	Actor          authn.UserRecord
	TargetRecordID uuid.UUID
	Request        SupersedeRequest
	RequestHash    []byte
	RequestID      string
	RouteKey       string
	Now            time.Time
}

type SupersedeMutationResult struct {
	Payload                 map[string]any
	StatusCode              int
	Replayed                bool
	IncidentID              uuid.UUID
	RecordID                uuid.UUID
	ChangeSetID             uuid.UUID
	ClientTxnID             string
	RowVersion              int64
	ViewSchemaID            string
	ChangedFieldKeys        []string
	AdditionalRecordChanges []SupersedeMutationResult
}

type SupersedeRowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *SupersedeRowVersionConflictError) Error() string {
	return "tasksdecisions: row version conflict"
}

func NewSupersedeFacade(pool postgres.DB) *SupersedeFacade {
	return &SupersedeFacade{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: incidents.NewAccess(pool),
		rowProjector:   projectionadapters.NewRowProjector(pool),
		recordStore:    records.NewStore(),
		revisionStore:  newRevisionAppendAdapter(),
		taskStore:      NewStore(),
	}
}

func (f *SupersedeFacade) SupersedeDecision(ctx context.Context, command SupersedeCommand) (SupersedeMutationResult, error) {
	request := command.Request
	if request.ReplacementRecordID == nil {
		return SupersedeMutationResult{}, &ValidationError{Field: "replacement_record_id", ReasonCode: "missing_required_field"}
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.TargetRecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return SupersedeMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeSupersedeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return SupersedeMutationResult{}, fmt.Errorf("decode replayed decision supersede payload: %w", err)
		}
		return SupersedeMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: command.TargetRecordID, ViewSchemaID: decisionsViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return SupersedeMutationResult{}, fmt.Errorf("query decision supersede idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SupersedeMutationResult{}, fmt.Errorf("begin decision supersede transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	targetMeta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if targetMeta.RecordType != "decision" {
		return SupersedeMutationResult{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, targetMeta.IncidentID); err != nil {
		return SupersedeMutationResult{}, err
	}
	if targetMeta.RowVersion != request.BaseRowVersion {
		return SupersedeMutationResult{}, &SupersedeRowVersionConflictError{RecordID: command.TargetRecordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: targetMeta.RowVersion}
	}

	sourceRecordID := *request.ReplacementRecordID
	if sourceRecordID == command.TargetRecordID {
		return SupersedeMutationResult{}, DecisionSupersedeValidationError("superseding_decision_must_be_different")
	}
	sourceMeta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, sourceRecordID)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
		return SupersedeMutationResult{}, DecisionSupersedeValidationError("superseding_decision_must_be_active_same_incident_decision")
	}
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if sourceMeta.RecordType != "decision" || sourceMeta.IncidentID != targetMeta.IncidentID {
		return SupersedeMutationResult{}, DecisionSupersedeValidationError("superseding_decision_must_be_active_same_incident_decision")
	}

	targetState, err := f.taskStore.LoadDecisionMachineStateForUpdateTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	sourceState, err := f.taskStore.LoadDecisionMachineStateForUpdateTx(ctx, tx, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := ValidateDecisionMachineState(targetState); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := ValidateDecisionMachineState(sourceState); err != nil {
		return SupersedeMutationResult{}, err
	}
	if sourceState.Status != "approved" && sourceState.Status != "executed" {
		return SupersedeMutationResult{}, DecisionSupersedeValidationError("superseding_decision_must_be_approved_or_executed")
	}
	if targetState.Status != "proposed" && targetState.Status != "approved" && targetState.Status != "executed" {
		return SupersedeMutationResult{}, DecisionSupersedeValidationError("target_decision_must_be_proposed_approved_or_executed")
	}
	if targetState.IncomingSupersedes > 0 {
		return SupersedeMutationResult{}, DecisionSupersedeValidationError("target_must_not_have_active_replacement")
	}

	if err := f.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.DecisionsViewSchemaID, command.TargetRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.DecisionsViewSchemaID, sourceRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	beforeTargetRow, err := f.rowProjector.LoadRowTx(ctx, tx, decisionsViewSchemaID, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	beforeSourceRow, err := f.rowProjector.LoadRowTx(ctx, tx, decisionsViewSchemaID, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}

	now := command.Now.UTC()
	linkID, err := f.taskStore.InsertDecisionSupersedesLinkTx(ctx, tx, targetMeta.IncidentID, sourceRecordID, command.TargetRecordID, command.Actor.ID, now)
	if err != nil {
		if authn.IsUniqueViolation(err) {
			return SupersedeMutationResult{}, DecisionSupersedeValidationError("target_must_not_have_active_replacement")
		}
		return SupersedeMutationResult{}, err
	}

	sourceVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, sourceRecordID, command.Actor.ID, now)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	targetVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.TargetRecordID, command.Actor.ID, now)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.taskStore.TouchSupersedingDecisionTx(ctx, tx, sourceRecordID, now); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.taskStore.MarkSupersededDecisionTx(ctx, tx, command.TargetRecordID, now); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.DecisionsViewSchemaID, sourceRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.DecisionsViewSchemaID, command.TargetRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	afterSourceRow, err := f.rowProjector.LoadRowTx(ctx, tx, decisionsViewSchemaID, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	afterTargetRow, err := f.rowProjector.LoadRowTx(ctx, tx, decisionsViewSchemaID, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	changeSetID, err := f.revisionStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  targetMeta.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		Reason:      &request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	sourceBeforeVersionID := supersedeVersionID(sourceRecordID, sourceMeta.RowVersion)
	sourceAfterVersionID := supersedeVersionID(sourceRecordID, sourceVersion)
	if err := f.revisionStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        sourceRecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &sourceBeforeVersionID,
		AfterVersionID:  &sourceAfterVersionID,
		BeforeValue:     beforeSourceRow,
		AfterValue:      afterSourceRow,
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	targetBeforeVersionID := supersedeVersionID(command.TargetRecordID, targetMeta.RowVersion)
	targetAfterVersionID := supersedeVersionID(command.TargetRecordID, targetVersion)
	if err := f.revisionStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      2,
		TargetKind:      "record",
		TargetID:        command.TargetRecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &targetBeforeVersionID,
		AfterVersionID:  &targetAfterVersionID,
		BeforeValue:     beforeTargetRow,
		AfterValue:      afterTargetRow,
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.revisionStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:   changeSetID,
		SequenceNo:    3,
		TargetKind:    "record_link",
		TargetID:      linkID.String(),
		OperationKind: "create",
		AfterValue: map[string]any{
			"record_link_id": linkID.String(),
			"incident_id":    targetMeta.IncidentID.String(),
			"src_record_id":  sourceRecordID.String(),
			"dst_record_id":  command.TargetRecordID.String(),
			"link_type":      "supersedes",
		},
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.revisionStore.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    sourceRecordID,
		RowVersion:  sourceVersion,
		BeforeValue: beforeSourceRow,
		AfterValue:  afterSourceRow,
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.revisionStore.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    command.TargetRecordID,
		RowVersion:  targetVersion,
		BeforeValue: beforeTargetRow,
		AfterValue:  afterTargetRow,
	}); err != nil {
		return SupersedeMutationResult{}, err
	}

	targetStatus := decisionRowStatus(afterTargetRow)
	payload := map[string]any{
		"view_schema_id":          decisionsViewSchemaID,
		"change_set_id":           changeSetID.String(),
		"target_record_id":        command.TargetRecordID.String(),
		"superseding_record_id":   sourceRecordID.String(),
		"target_row_version":      targetVersion,
		"superseding_row_version": sourceVersion,
		"target_status":           targetStatus,
		"reason":                  request.Reason,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return SupersedeMutationResult{}, authn.ErrClientTxnConflict
		}
		return SupersedeMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SupersedeMutationResult{}, fmt.Errorf("commit decision supersede transaction: %w", err)
	}

	targetChange := SupersedeMutationResult{
		Payload:          map[string]any{"row": afterTargetRow},
		StatusCode:       http.StatusOK,
		IncidentID:       targetMeta.IncidentID,
		RecordID:         command.TargetRecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       targetVersion,
		ViewSchemaID:     decisionsViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeTargetRow, afterTargetRow),
	}
	sourceChange := SupersedeMutationResult{
		Payload:          map[string]any{"row": afterSourceRow},
		StatusCode:       http.StatusOK,
		IncidentID:       targetMeta.IncidentID,
		RecordID:         sourceRecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       sourceVersion,
		ViewSchemaID:     decisionsViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeSourceRow, afterSourceRow),
	}
	return SupersedeMutationResult{
		Payload:                 payload,
		StatusCode:              http.StatusOK,
		IncidentID:              targetMeta.IncidentID,
		RecordID:                command.TargetRecordID,
		ChangeSetID:             changeSetID,
		ClientTxnID:             request.ClientTxnID,
		RowVersion:              targetVersion,
		ViewSchemaID:            decisionsViewSchemaID,
		ChangedFieldKeys:        targetChange.ChangedFieldKeys,
		AdditionalRecordChanges: []SupersedeMutationResult{targetChange, sourceChange},
	}, nil
}

type supersedeRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func loadSupersedeRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (supersedeRecordMeta, error) {
	var meta supersedeRecordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return supersedeRecordMeta{}, err
	}
	if deletedAt.Valid {
		return supersedeRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

func decisionRowStatus(row map[string]any) string {
	cells, _ := row["cells"].(map[string]any)
	cell, _ := cells["decision.status"].(map[string]any)
	status, _ := cell["value"].(string)
	return status
}

func supersedeVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func decodeSupersedeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
