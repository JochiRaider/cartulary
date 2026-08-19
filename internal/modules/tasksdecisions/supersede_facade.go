package tasksdecisions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	recordsmodule "github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
)

const DecisionsViewSchemaID = "cartulary.view.decisions.v1"

type SupersedeRequest struct {
	BaseRowVersion      int64
	ClientTxnID         string
	Reason              string
	ReplacementRecordID *uuid.UUID
}

type SupersedeCommand struct {
	ActorUserID    uuid.UUID
	TargetRecordID uuid.UUID
	Request        SupersedeRequest
	RequestHash    []byte
	RequestID      string
	RouteKey       string
	Now            time.Time
}

type SupersedeMutationResult struct {
	Row                     map[string]any
	Replayed                bool
	IncidentID              uuid.UUID
	RecordID                uuid.UUID
	ChangeSetID             uuid.UUID
	ClientTxnID             string
	RowVersion              int64
	ViewSchemaID            string
	ChangedFieldKeys        []string
	AdditionalRecordChanges []SupersedeMutationResult
	Facts                   SupersedeFacts
}

type SupersedeFacts struct {
	TargetRecordID        uuid.UUID
	SupersedingRecordID   uuid.UUID
	TargetRowVersion      int64
	SupersedingRowVersion int64
	TargetStatus          string
	Reason                string
}

type SupersedeRowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *SupersedeRowVersionConflictError) Error() string {
	return "tasksdecisions: row version conflict"
}

func (f *MutationFacade) SupersedeDecision(ctx context.Context, command SupersedeCommand) (SupersedeMutationResult, error) {
	request := command.Request
	if request.ReplacementRecordID == nil {
		return SupersedeMutationResult{}, &ValidationError{Field: "replacement_record_id", ReasonCode: "missing_required_field"}
	}
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.TargetRecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return SupersedeMutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationDecisionSupersession {
			return SupersedeMutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.DecisionSupersessionResult()
		if !ok || stored.ViewSchemaID != DecisionsViewSchemaID || stored.Facts.TargetRecordID != command.TargetRecordID {
			return SupersedeMutationResult{}, ErrStoredMutationKindMismatch
		}
		return SupersedeMutationResult{Replayed: true, RecordID: command.TargetRecordID, ChangeSetID: stored.ChangeSetID, ViewSchemaID: DecisionsViewSchemaID, ClientTxnID: request.ClientTxnID, RowVersion: stored.Facts.TargetRowVersion, Facts: stored.Facts}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return SupersedeMutationResult{}, fmt.Errorf("query decision supersede idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SupersedeMutationResult{}, fmt.Errorf("begin decision supersede transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	targetMeta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, f.recordStore, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if targetMeta.RecordType != "decision" {
		return SupersedeMutationResult{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, targetMeta.IncidentID); err != nil {
		return SupersedeMutationResult{}, err
	}
	if targetMeta.RowVersion != request.BaseRowVersion {
		return SupersedeMutationResult{}, &SupersedeRowVersionConflictError{RecordID: command.TargetRecordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: targetMeta.RowVersion}
	}

	sourceRecordID := *request.ReplacementRecordID
	if sourceRecordID == command.TargetRecordID {
		return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("superseding_decision_must_be_different")
	}
	sourceMeta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, f.recordStore, sourceRecordID)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
		return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("superseding_decision_must_be_active_same_incident_decision")
	}
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if sourceMeta.RecordType != "decision" || sourceMeta.IncidentID != targetMeta.IncidentID {
		return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("superseding_decision_must_be_active_same_incident_decision")
	}

	targetState, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	sourceState, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := policy.ValidateDecisionMachineState(targetState); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := policy.ValidateDecisionMachineState(sourceState); err != nil {
		return SupersedeMutationResult{}, err
	}
	if sourceState.Status != "approved" && sourceState.Status != "executed" {
		return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("superseding_decision_must_be_approved_or_executed")
	}
	if targetState.Status != "proposed" && targetState.Status != "approved" && targetState.Status != "executed" {
		return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("target_decision_must_be_proposed_approved_or_executed")
	}
	if targetState.IncomingSupersedes > 0 {
		return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("target_must_not_have_active_replacement")
	}

	if err := f.projectionRows.RefreshDecisionTx(ctx, tx, command.TargetRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.projectionRows.RefreshDecisionTx(ctx, tx, sourceRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	beforeTargetRow, err := f.projectionRows.LoadDecisionTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	beforeSourceRow, err := f.projectionRows.LoadDecisionTx(ctx, tx, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	beforeTargetSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	beforeSourceSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}

	now := command.Now.UTC()
	link, err := f.linkStore.InsertSupersedesCommandTx(ctx, tx, links.InsertSupersedesCommand{
		IncidentID:          targetMeta.IncidentID,
		ReplacementRecordID: sourceRecordID,
		SupersededRecordID:  command.TargetRecordID,
		OwnerUserID:         command.ActorUserID,
		Now:                 now,
	})
	if err != nil {
		if tasksource.IsUniqueViolation(err) {
			return SupersedeMutationResult{}, policy.DecisionSupersedeValidationError("target_must_not_have_active_replacement")
		}
		return SupersedeMutationResult{}, err
	}
	linkID := link.RecordLinkID

	sourceVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, sourceRecordID, command.ActorUserID, now)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	targetVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.TargetRecordID, command.ActorUserID, now)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := tasksource.TouchSupersedingDecisionTx(ctx, tx, sourceRecordID, now); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := tasksource.MarkSupersededDecisionTx(ctx, tx, command.TargetRecordID, now); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.projectionRows.RefreshDecisionTx(ctx, tx, sourceRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.projectionRows.RefreshDecisionTx(ctx, tx, command.TargetRecordID); err != nil {
		return SupersedeMutationResult{}, err
	}
	afterSourceRow, err := f.projectionRows.LoadDecisionTx(ctx, tx, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	afterTargetRow, err := f.projectionRows.LoadDecisionTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	afterSourceSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, sourceRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	afterTargetSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.TargetRecordID)
	if err != nil {
		return SupersedeMutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  targetMeta.IncidentID,
		ActorUserID: command.ActorUserID,
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
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		RecordID:        sourceRecordID,
		OperationKind:   "patch",
		BeforeVersionID: &sourceBeforeVersionID,
		AfterVersionID:  &sourceAfterVersionID,
		BeforeSnapshot:  &beforeSourceSnapshot,
		AfterSnapshot:   &afterSourceSnapshot,
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	targetBeforeVersionID := supersedeVersionID(command.TargetRecordID, targetMeta.RowVersion)
	targetAfterVersionID := supersedeVersionID(command.TargetRecordID, targetVersion)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      2,
		TargetKind:      "record",
		RecordID:        command.TargetRecordID,
		OperationKind:   "patch",
		BeforeVersionID: &targetBeforeVersionID,
		AfterVersionID:  &targetAfterVersionID,
		BeforeSnapshot:  &beforeTargetSnapshot,
		AfterSnapshot:   &afterTargetSnapshot,
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.revisions.AppendMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
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
	if err := f.revisions.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       sourceRecordID,
		RowVersion:     sourceVersion,
		BeforeSnapshot: &beforeSourceSnapshot,
		AfterSnapshot:  &afterSourceSnapshot,
		LiveChange:     revisions.LiveRecordChange{BeforeValue: beforeSourceRow, AfterValue: afterSourceRow},
	}); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := f.revisions.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       command.TargetRecordID,
		RowVersion:     targetVersion,
		BeforeSnapshot: &beforeTargetSnapshot,
		AfterSnapshot:  &afterTargetSnapshot,
		LiveChange:     revisions.LiveRecordChange{BeforeValue: beforeTargetRow, AfterValue: afterTargetRow},
	}); err != nil {
		return SupersedeMutationResult{}, err
	}

	targetStatus := decisionRowStatus(afterTargetRow)
	storedResult := NewStoredDecisionSupersessionResult(StoredDecisionSupersessionResult{
		ViewSchemaID: DecisionsViewSchemaID,
		ChangeSetID:  changeSetID,
		Facts: SupersedeFacts{
			TargetRecordID: command.TargetRecordID, SupersedingRecordID: sourceRecordID,
			TargetRowVersion: targetVersion, SupersedingRowVersion: sourceVersion,
			TargetStatus: targetStatus, Reason: request.Reason,
		},
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return SupersedeMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SupersedeMutationResult{}, fmt.Errorf("commit decision supersede transaction: %w", err)
	}

	targetChange := SupersedeMutationResult{
		Row:              afterTargetRow,
		IncidentID:       targetMeta.IncidentID,
		RecordID:         command.TargetRecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       targetVersion,
		ViewSchemaID:     DecisionsViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeTargetRow, afterTargetRow),
	}
	sourceChange := SupersedeMutationResult{
		Row:              afterSourceRow,
		IncidentID:       targetMeta.IncidentID,
		RecordID:         sourceRecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       sourceVersion,
		ViewSchemaID:     DecisionsViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeSourceRow, afterSourceRow),
	}
	return SupersedeMutationResult{
		IncidentID:              targetMeta.IncidentID,
		RecordID:                command.TargetRecordID,
		ChangeSetID:             changeSetID,
		ClientTxnID:             request.ClientTxnID,
		RowVersion:              targetVersion,
		ViewSchemaID:            DecisionsViewSchemaID,
		ChangedFieldKeys:        targetChange.ChangedFieldKeys,
		AdditionalRecordChanges: []SupersedeMutationResult{targetChange, sourceChange},
		Facts: SupersedeFacts{
			TargetRecordID: command.TargetRecordID, SupersedingRecordID: sourceRecordID,
			TargetRowVersion: targetVersion, SupersedingRowVersion: sourceVersion,
			TargetStatus: targetStatus, Reason: request.Reason,
		},
	}, nil
}

type supersedeRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func loadSupersedeRecordMetaForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	envelopes RecordEnvelopeCapability,
	recordID uuid.UUID,
) (supersedeRecordMeta, error) {
	envelope, err := envelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if errors.Is(err, recordsmodule.ErrEnvelopeNotFound) {
		return supersedeRecordMeta{}, pgx.ErrNoRows
	}
	if err != nil {
		return supersedeRecordMeta{}, err
	}
	if envelope.DeletedAt != nil {
		return supersedeRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return supersedeRecordMeta{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		RowVersion: envelope.RowVersion,
	}, nil
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
