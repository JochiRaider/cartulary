package indicators

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Store) CreateIndicatorRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, command CreateCommand, requestHash []byte, requestID string, now time.Time) (CreateResult, error) {
	return s.createService.createIndicatorRow(ctx, actor, incidentID, command, requestHash, requestID, now)
}

type indicatorCreateService struct {
	owner *Store
}

func (service indicatorCreateService) createIndicatorRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, command CreateCommand, requestHash []byte, requestID string, now time.Time) (CreateResult, error) {
	s := service.owner
	scopeKey := incidentID.String() + ":" + ViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    indicatorCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: command.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return CreateResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return CreateResult{}, fmt.Errorf("decode replayed indicator create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return CreateResult{}, err
		}
		changeSetID, err := extractUUIDFromPayload(payload, "change_set_id")
		if err != nil {
			return CreateResult{}, err
		}
		row, ok := payload["row"].(map[string]any)
		if !ok {
			return CreateResult{}, fmt.Errorf("decode replayed indicator create row")
		}
		rowVersion, err := extractInt64FromPayload(payload, "row", "row_version")
		if err != nil {
			return CreateResult{}, err
		}
		return CreateResult{
			Outcome: CreateOutcomeReplayed, CanonicalRow: row,
			RecordID: recordID, ChangeSetID: changeSetID, RowVersion: rowVersion,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return CreateResult{}, fmt.Errorf("query indicator create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateResult{}, fmt.Errorf("begin indicator create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return CreateResult{}, err
	}
	beforeSnapshot, err := s.captureIndicatorSnapshotBeforeUpsertTx(ctx, tx, incidentID, command)
	if err != nil {
		return CreateResult{}, err
	}
	record, beforeRow, operationKind, _, err := s.upsertIndicatorTx(ctx, tx, actor, incidentID, command, now)
	if err != nil {
		return CreateResult{}, err
	}
	afterSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return CreateResult{}, err
	}
	afterRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, record.RecordID)
	if err != nil {
		return CreateResult{}, err
	}

	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      indicatorCreateRouteKey,
		ClientTxnID: &command.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return CreateResult{}, err
	}

	var beforeVersionID *string
	if beforeRow != nil {
		beforeVersion := record.RowVersion
		if !jsonEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("indicator", record.RecordID, beforeVersion)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("indicator", record.RecordID, record.RowVersion)
	if err := s.revisionsStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "indicator",
		RecordID:        record.RecordID,
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return CreateResult{}, err
	}
	if beforeRow == nil || !jsonEqual(beforeRow, afterRow) {
		if err := s.revisionsStore.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID:    changeSetID,
			RecordID:       record.RecordID,
			RowVersion:     record.RowVersion,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			LiveChange: revisions.LiveRecordChange{
				BeforeValue: beforeRow,
				AfterValue:  afterRow,
			},
		}); err != nil {
			return CreateResult{}, err
		}
	}

	outcome := CreateOutcomeCreated
	statusCode := httpStatusCreated
	if beforeRow != nil {
		outcome = CreateOutcomeReused
		statusCode = httpStatusOK
		if !jsonEqual(beforeRow, afterRow) {
			outcome = CreateOutcomeUpdated
		}
	}
	payload := buildStoredCreateResponse(changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, statusCode, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return CreateResult{}, authn.ErrClientTxnConflict
		}
		return CreateResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateResult{}, fmt.Errorf("commit indicator create transaction: %w", err)
	}

	return CreateResult{
		Outcome: outcome, CanonicalRow: afterRow, RecordID: record.RecordID,
		ChangeSetID: changeSetID, RowVersion: record.RowVersion,
	}, nil
}

func (s *Store) captureIndicatorSnapshotBeforeUpsertTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, command CreateCommand) (*revisions.RecordSnapshot, error) {
	input, err := indicatorInputFromCreateCommand(command)
	if err != nil {
		return nil, err
	}
	current, matched, err := s.sources.loadByDedupeTx(ctx, tx, incidentID, input.IndicatorType, input.DedupeKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if !matched {
		return nil, nil
	}
	snapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *Store) FindOrCreateIndicatorParticipantTx(ctx context.Context, tx pgx.Tx, command IndicatorFindOrCreateParticipantCommand) (IndicatorFindOrCreateParticipantResult, error) {
	if command.IncidentID == uuid.Nil {
		return IndicatorFindOrCreateParticipantResult{}, &IndicatorCreateValidationError{Field: "incident_id", ReasonCode: "missing_required_field"}
	}
	if command.Actor.ID == uuid.Nil {
		return IndicatorFindOrCreateParticipantResult{}, &IndicatorCreateValidationError{Field: "actor_user_id", ReasonCode: "missing_required_field"}
	}
	if strings.TrimSpace(command.OperationContext) == "" {
		return IndicatorFindOrCreateParticipantResult{}, &IndicatorCreateValidationError{Field: "operation_context", ReasonCode: "missing_required_field"}
	}
	if command.OperationOccurred.IsZero() {
		return IndicatorFindOrCreateParticipantResult{}, &IndicatorCreateValidationError{Field: "operation_occurred", ReasonCode: "missing_required_field"}
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return IndicatorFindOrCreateParticipantResult{}, err
	}

	record, _, operationKind, _, err := s.upsertIndicatorTx(ctx, tx, command.Actor, command.IncidentID, CreateCommand{
		IndicatorType:   command.IndicatorType,
		ValueKind:       command.ValueKind,
		DisplayValue:    command.DisplayValue,
		NormalizedValue: command.NormalizedValue,
	}, command.OperationOccurred.UTC())
	if err != nil {
		return IndicatorFindOrCreateParticipantResult{}, err
	}
	if err := s.refreshProjectionRowTx(ctx, tx, record.RecordID); err != nil {
		return IndicatorFindOrCreateParticipantResult{}, err
	}
	status := "reused"
	if operationKind == "create" {
		status = "created"
	}
	return IndicatorFindOrCreateParticipantResult{
		SchemaID:  IndicatorFindOrCreateParticipantV1,
		Status:    status,
		Indicator: referenceFromRecord(record),
	}, nil
}

func (s *Store) upsertIndicatorTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, command CreateCommand, now time.Time) (indicatorRecord, map[string]any, string, int, error) {
	input, err := indicatorInputFromCreateCommand(command)
	if err != nil {
		return indicatorRecord{}, nil, "", 0, err
	}
	current, matched, err := s.sources.loadByDedupeTx(ctx, tx, incidentID, input.IndicatorType, input.DedupeKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return indicatorRecord{}, nil, "", 0, err
	}
	if !matched {
		record := indicatorRecord{
			IncidentID:      incidentID,
			IndicatorType:   input.IndicatorType,
			ValueKind:       input.ValueKind,
			DisplayValue:    input.DisplayValue,
			NormalizedValue: cloneStringPointer(input.NormalizedValue),
			DedupeKey:       input.DedupeKey,
			DefangedValue:   cloneStringPointer(input.DefangedValue),
			HashAlgorithm:   cloneStringPointer(input.HashAlgorithm),
			HashValue:       cloneStringPointer(input.HashValue),
			STIXPattern:     cloneStringPointer(input.STIXPattern),
			RowVersion:      1,
			CreatedAt:       now.UTC(),
			UpdatedAt:       now.UTC(),
			CreatedByUser:   actor.ID,
			UpdatedByUser:   actor.ID,
		}
		recordID, err := s.recordStore.InsertTx(ctx, tx, records.InsertParams{
			IncidentID:      incidentID,
			RecordType:      "indicator",
			CreatedByUserID: actor.ID,
			CreatedAt:       record.CreatedAt,
			UpdatedByUserID: actor.ID,
			UpdatedAt:       record.UpdatedAt,
			RowVersion:      record.RowVersion,
		})
		if err != nil {
			return indicatorRecord{}, nil, "", 0, err
		}
		record.RecordID = recordID
		if err := s.sources.insertTx(ctx, tx, &record); err != nil {
			return indicatorRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	beforeRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, current.RecordID)
	if err != nil {
		return indicatorRecord{}, nil, "", 0, err
	}

	next := current
	fieldChanged := false
	if next.DefangedValue == nil && input.DefangedValue != nil {
		next.DefangedValue = cloneStringPointer(input.DefangedValue)
		fieldChanged = true
	}
	if next.HashAlgorithm == nil && input.HashAlgorithm != nil {
		next.HashAlgorithm = cloneStringPointer(input.HashAlgorithm)
		fieldChanged = true
	}
	if next.HashValue == nil && input.HashValue != nil {
		next.HashValue = cloneStringPointer(input.HashValue)
		fieldChanged = true
	}
	if next.STIXPattern == nil && input.STIXPattern != nil {
		next.STIXPattern = cloneStringPointer(input.STIXPattern)
		fieldChanged = true
	}
	if fieldChanged {
		next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			return indicatorRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := s.sources.updateTx(ctx, tx, next); err != nil {
			return indicatorRecord{}, nil, "", 0, err
		}
	}
	return next, beforeRow, "patch", httpStatusOK, nil
}
