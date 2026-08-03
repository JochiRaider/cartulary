package indicators

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	observationCreateSource  = "indicators.observations.capture"
	observationResolveSource = "indicators.observations.resolve"
	lifecycleAppendSource    = "indicators.lifecycle.append"
)

var (
	ErrIndicatorNotFound            = errors.New("indicators: indicator not found")
	ErrIndicatorObservationNotFound = errors.New("indicators: indicator observation not found")
)

type incidentLifecycleAccess interface {
	EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

func newIncidentLifecycleAccess(pool postgres.DB) incidentLifecycleAccess {
	return incidents.NewAccess(pool)
}

type IndicatorCreateValidationError struct {
	Field      string
	ReasonCode string
}

func (e *IndicatorCreateValidationError) Error() string {
	return fmt.Sprintf("invalid indicator create payload: %s %s", e.Field, e.ReasonCode)
}

type IndicatorObservationCreateParams struct {
	IncidentID                uuid.UUID
	SourceRecordID            uuid.UUID
	SourceFieldKey            string
	OriginKind                string
	OriginLocator             string
	ObservedText              string
	ParsedIndicatorType       *string
	NormalizedCandidate       *string
	ResolvedIndicatorRecordID *uuid.UUID
	ResolutionMethod          *string
	RequestID                 *string
	ClientTxnID               *string
	MutationSource            string
	CreatedAt                 time.Time
}

type IndicatorObservationResolveParams struct {
	ObservationID             uuid.UUID
	ResolvedIndicatorRecordID uuid.UUID
	RequestID                 *string
	ClientTxnID               *string
	MutationSource            string
	ResolvedAt                time.Time
}

type IndicatorLifecycleAppendParams struct {
	IncidentID        uuid.UUID
	IndicatorRecordID uuid.UUID
	LifecycleState    string
	ValidFrom         time.Time
	ValidTo           *time.Time
	Confidence        *int
	Rationale         *string
	SupportRefs       []string
	Assessor          *string
	RequestID         *string
	ClientTxnID       *string
	MutationSource    string
	CreatedAt         time.Time
}

type IndicatorObservationRecord struct {
	ObservationID             uuid.UUID
	IncidentID                uuid.UUID
	SourceRecordID            uuid.UUID
	SourceFieldKey            string
	OriginKind                string
	OriginLocator             string
	ObservedText              string
	ParsedIndicatorType       *string
	NormalizedCandidate       *string
	ResolutionStatus          string
	ResolvedIndicatorRecordID *uuid.UUID
	RowVersion                int64
	CreatedByUserID           uuid.UUID
	CreatedAt                 time.Time
	ResolvedByUserID          *uuid.UUID
	ResolvedAt                *time.Time
	ResolutionMethod          *string
	DeletedAt                 *time.Time
	DeletedByUserID           *uuid.UUID
}

type IndicatorLifecycleIntervalRecord struct {
	IntervalID        uuid.UUID
	IncidentID        uuid.UUID
	IndicatorRecordID uuid.UUID
	LifecycleState    string
	ValidFrom         time.Time
	ValidTo           *time.Time
	Confidence        *int
	Rationale         *string
	SupportRefs       []string
	Assessor          *string
	AssessedAt        time.Time
	RowVersion        int64
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
}

type indicatorUpsertInput = identity.Canonical

func (s *Store) CreateIndicatorRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, command CreateCommand, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + ViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    indicatorCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: command.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed indicator create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: httpStatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query indicator create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin indicator create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	record, beforeRow, operationKind, statusCode, err := s.upsertIndicatorTx(ctx, tx, actor, incidentID, command, now)
	if err != nil {
		return MutationResult{}, err
	}
	projected, err := refreshIndicatorProjectionTx(ctx, tx, record.RecordID)
	if err != nil {
		return MutationResult{}, err
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
		return MutationResult{}, err
	}

	afterRow := BuildIndicatorRow(projected)
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
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "indicator",
		TargetID:        record.RecordID.String(),
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if beforeRow == nil || !jsonEqual(beforeRow, afterRow) {
		if err := s.revisionsStore.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    record.RecordID,
			RowVersion:  record.RowVersion,
			BeforeValue: beforeRow,
			AfterValue:  afterRow,
		}); err != nil {
			return MutationResult{}, err
		}
	}

	payload := BuildMutationPayload(changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, statusCode, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit indicator create transaction: %w", err)
	}

	return MutationResult{
		Payload:     payload,
		StatusCode:  statusCode,
		RecordID:    record.RecordID,
		ChangeSetID: changeSetID,
		RowVersion:  record.RowVersion,
	}, nil
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
	if _, err := refreshIndicatorProjectionTx(ctx, tx, record.RecordID); err != nil {
		return IndicatorFindOrCreateParticipantResult{}, err
	}
	status := "reused"
	if operationKind == "create" {
		status = "created"
	}
	return IndicatorFindOrCreateParticipantResult{
		SchemaID:  IndicatorFindOrCreateParticipantV1,
		Status:    status,
		Indicator: record,
	}, nil
}

func (s *Store) CreateIndicatorObservation(ctx context.Context, actor authn.UserRecord, params IndicatorObservationCreateParams) (IndicatorObservationRecord, uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("begin indicator observation transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	createdAt := params.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	record, err := insertIndicatorObservationTx(ctx, tx, actor.ID, params, createdAt)
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}

	changeSource := params.MutationSource
	if strings.TrimSpace(changeSource) == "" {
		changeSource = observationCreateSource
	}
	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  params.IncidentID,
		ActorUserID: actor.ID,
		Source:      changeSource,
		ClientTxnID: params.ClientTxnID,
		RequestID:   params.RequestID,
		CreatedAt:   createdAt,
	})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	afterValue := buildIndicatorObservationValue(record)
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "indicator_observation",
		TargetID:       record.ObservationID.String(),
		OperationKind:  "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_observation:%s:%d", record.ObservationID.String(), record.RowVersion)),
		AfterValue:     afterValue,
	}); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if record.ResolvedIndicatorRecordID != nil {
		if _, err := refreshIndicatorProjectionTx(ctx, tx, *record.ResolvedIndicatorRecordID); err != nil {
			return IndicatorObservationRecord{}, uuid.UUID{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("commit indicator observation transaction: %w", err)
	}
	return record, changeSetID, nil
}

func (s *Store) ResolveIndicatorObservation(ctx context.Context, actor authn.UserRecord, params IndicatorObservationResolveParams) (IndicatorObservationRecord, uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("begin indicator observation resolve transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := loadIndicatorObservationTx(ctx, tx, params.ObservationID, true)
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, current.IncidentID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := validateIndicatorRecordIncidentTx(ctx, tx, current.IncidentID, params.ResolvedIndicatorRecordID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}

	resolvedAt := params.ResolvedAt.UTC()
	if resolvedAt.IsZero() {
		resolvedAt = time.Now().UTC()
	}
	next := current
	next.ResolutionStatus = "resolved"
	next.ResolvedIndicatorRecordID = &params.ResolvedIndicatorRecordID
	next.ResolvedByUserID = &actor.ID
	next.ResolvedAt = &resolvedAt
	method := params.MutationSource
	if strings.TrimSpace(method) == "" {
		method = observationResolveSource
	}
	next.ResolutionMethod = &method
	next.RowVersion = current.RowVersion + 1
	tag, err := tx.Exec(ctx, `
UPDATE indicator_observations
   SET resolution_status = 'resolved',
       resolved_indicator_record_id = $2,
       resolved_by_user_id = $3,
       resolved_at = $4,
       resolution_method = $5,
       row_version = $6
 WHERE indicator_observation_id = $1
   AND deleted_at IS NULL
`, next.ObservationID, params.ResolvedIndicatorRecordID, actor.ID, resolvedAt, method, next.RowVersion)
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("resolve indicator observation: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return IndicatorObservationRecord{}, uuid.UUID{}, ErrIndicatorObservationNotFound
	}

	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  current.IncidentID,
		ActorUserID: actor.ID,
		Source:      method,
		ClientTxnID: params.ClientTxnID,
		RequestID:   params.RequestID,
		CreatedAt:   resolvedAt,
	})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "indicator_observation",
		TargetID:        next.ObservationID.String(),
		OperationKind:   "resolve",
		BeforeVersionID: stringPointer(fmt.Sprintf("indicator_observation:%s:%d", current.ObservationID.String(), current.RowVersion)),
		AfterVersionID:  stringPointer(fmt.Sprintf("indicator_observation:%s:%d", next.ObservationID.String(), next.RowVersion)),
		BeforeValue:     buildIndicatorObservationValue(current),
		AfterValue:      buildIndicatorObservationValue(next),
	}); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	projectionIDs := []uuid.UUID{params.ResolvedIndicatorRecordID}
	if current.ResolvedIndicatorRecordID != nil && *current.ResolvedIndicatorRecordID != params.ResolvedIndicatorRecordID {
		projectionIDs = append(projectionIDs, *current.ResolvedIndicatorRecordID)
	}
	for _, recordID := range projectionIDs {
		if _, err := refreshIndicatorProjectionTx(ctx, tx, recordID); err != nil {
			return IndicatorObservationRecord{}, uuid.UUID{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("commit indicator observation resolve transaction: %w", err)
	}
	return next, changeSetID, nil
}

func (s *Store) AppendIndicatorLifecycleInterval(ctx context.Context, actor authn.UserRecord, params IndicatorLifecycleAppendParams) (IndicatorLifecycleIntervalRecord, uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, fmt.Errorf("begin indicator lifecycle transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := validateIndicatorRecordIncidentTx(ctx, tx, params.IncidentID, params.IndicatorRecordID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	createdAt := params.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	record, err := insertIndicatorLifecycleIntervalTx(ctx, tx, actor.ID, params, createdAt)
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}

	changeSource := params.MutationSource
	if strings.TrimSpace(changeSource) == "" {
		changeSource = lifecycleAppendSource
	}
	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  params.IncidentID,
		ActorUserID: actor.ID,
		Source:      changeSource,
		ClientTxnID: params.ClientTxnID,
		RequestID:   params.RequestID,
		CreatedAt:   createdAt,
	})
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "indicator_state_interval",
		TargetID:       record.IntervalID.String(),
		OperationKind:  "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_state_interval:%s:%d", record.IntervalID.String(), record.RowVersion)),
		AfterValue:     buildIndicatorLifecycleValue(record),
	}); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if _, err := refreshIndicatorProjectionTx(ctx, tx, params.IndicatorRecordID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, fmt.Errorf("commit indicator lifecycle transaction: %w", err)
	}
	return record, changeSetID, nil
}

func (s *Store) upsertIndicatorTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, command CreateCommand, now time.Time) (IndicatorRecord, map[string]any, string, int, error) {
	input, err := indicatorInputFromCreateCommand(command)
	if err != nil {
		return IndicatorRecord{}, nil, "", 0, err
	}
	current, matched, err := loadIndicatorByDedupeTx(ctx, tx, incidentID, input.IndicatorType, input.DedupeKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return IndicatorRecord{}, nil, "", 0, err
	}
	if !matched {
		record := IndicatorRecord{
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
			return IndicatorRecord{}, nil, "", 0, err
		}
		record.RecordID = recordID
		if err := insertIndicatorTx(ctx, tx, &record); err != nil {
			return IndicatorRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	beforeProjected, err := refreshIndicatorProjectionTx(ctx, tx, current.RecordID)
	if err != nil {
		return IndicatorRecord{}, nil, "", 0, err
	}
	beforeRow := BuildIndicatorRow(beforeProjected)

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
			return IndicatorRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := updateIndicatorTx(ctx, tx, next); err != nil {
			return IndicatorRecord{}, nil, "", 0, err
		}
	}
	return next, beforeRow, "patch", httpStatusOK, nil
}

func indicatorInputFromCreateCommand(command CreateCommand) (indicatorUpsertInput, error) {
	input, err := identity.Canonicalize(identity.Input{
		IndicatorType:   command.IndicatorType,
		ValueKind:       command.ValueKind,
		DisplayValue:    command.DisplayValue,
		NormalizedValue: command.NormalizedValue,
		DefangedValue:   command.DefangedValue,
		HashAlgorithm:   command.HashAlgorithm,
		HashValue:       command.HashValue,
		STIXPattern:     command.STIXPattern,
	})
	if err == nil {
		return input, nil
	}
	var validationError *identity.ValidationError
	if errors.As(err, &validationError) {
		return indicatorUpsertInput{}, &IndicatorCreateValidationError{
			Field:      "indicator." + validationError.Field,
			ReasonCode: validationError.ReasonCode,
		}
	}
	return indicatorUpsertInput{}, err
}

func loadIndicatorByDedupeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, indicatorType string, dedupeKey string) (IndicatorRecord, bool, error) {
	record, err := scanIndicatorRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id,
    r.deleted_at,
    r.deleted_by_user_id
  FROM indicators i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND i.indicator_type = $2
   AND i.dedupe_key = $3
   AND r.deleted_at IS NULL
 LIMIT 1
 FOR UPDATE OF i, r
`, incidentID, indicatorType, dedupeKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorRecord{}, false, nil
	}
	if err != nil {
		return IndicatorRecord{}, false, fmt.Errorf("load indicator by dedupe: %w", err)
	}
	return record, true, nil
}

func insertIndicatorTx(ctx context.Context, tx pgx.Tx, record *IndicatorRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO indicators (
    record_id,
    incident_id,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13, $14, $14)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.IndicatorType, record.ValueKind, record.DisplayValue, record.NormalizedValue, record.DedupeKey, record.DefangedValue, record.HashAlgorithm, record.HashValue, record.STIXPattern, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateIndicatorTx(ctx context.Context, tx pgx.Tx, record IndicatorRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE indicators
   SET defanged_value = $2,
       hash_algorithm = $3,
       hash_value = $4,
       stix_pattern = $5,
       row_version = $6,
       updated_at = $7,
       updated_by_user_id = $8
 WHERE record_id = $1
`, record.RecordID, record.DefangedValue, record.HashAlgorithm, record.HashValue, record.STIXPattern, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update indicator: %w", err)
	}
	return nil
}

func refreshIndicatorProjectionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IndicatorProjectionRecord, error) {
	record, err := loadIndicatorRecordTx(ctx, tx, recordID)
	if err != nil {
		return IndicatorProjectionRecord{}, err
	}
	aggregate, err := loadIndicatorObservationAggregateTx(ctx, tx, record.RecordID)
	if err != nil {
		return IndicatorProjectionRecord{}, err
	}
	lifecycleSummary, err := loadIndicatorLifecycleSummaryTx(ctx, tx, record.RecordID)
	if err != nil {
		return IndicatorProjectionRecord{}, err
	}
	supportingLinkCount, err := loadIndicatorSupportingLinkCountTx(ctx, tx, record.RecordID)
	if err != nil {
		return IndicatorProjectionRecord{}, err
	}
	projected := IndicatorProjectionRecord{
		IndicatorRecord:   record,
		FirstObservedAt:   aggregate.FirstObservedAt,
		LastObservedAt:    aggregate.LastObservedAt,
		ObservationCount:  aggregate.ObservationCount,
		LifecycleSummary:  lifecycleSummary,
		SupportingLinkCnt: supportingLinkCount,
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO indicator_grid_projection (
    record_id,
    incident_id,
    row_version,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    first_observed_at,
    last_observed_at,
    observation_count,
    lifecycle_summary,
    supporting_link_count,
    edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    indicator_type = EXCLUDED.indicator_type,
    value_kind = EXCLUDED.value_kind,
    display_value = EXCLUDED.display_value,
    normalized_value = EXCLUDED.normalized_value,
    dedupe_key = EXCLUDED.dedupe_key,
    defanged_value = EXCLUDED.defanged_value,
    hash_algorithm = EXCLUDED.hash_algorithm,
    hash_value = EXCLUDED.hash_value,
    stix_pattern = EXCLUDED.stix_pattern,
    first_observed_at = EXCLUDED.first_observed_at,
    last_observed_at = EXCLUDED.last_observed_at,
    observation_count = EXCLUDED.observation_count,
    lifecycle_summary = EXCLUDED.lifecycle_summary,
    supporting_link_count = EXCLUDED.supporting_link_count,
    edited_at = EXCLUDED.edited_at
`, projected.RecordID, projected.IncidentID, projected.RowVersion, projected.IndicatorType, projected.ValueKind, projected.DisplayValue, projected.NormalizedValue, projected.DedupeKey, projected.DefangedValue, projected.HashAlgorithm, projected.HashValue, projected.STIXPattern, projected.FirstObservedAt, projected.LastObservedAt, projected.ObservationCount, projected.LifecycleSummary, projected.SupportingLinkCnt, projected.UpdatedAt.UTC()); err != nil {
		return IndicatorProjectionRecord{}, fmt.Errorf("upsert indicator projection: %w", err)
	}
	return projected, nil
}

type indicatorObservationAggregate struct {
	FirstObservedAt  *time.Time
	LastObservedAt   *time.Time
	ObservationCount int
}

func loadIndicatorObservationAggregateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (indicatorObservationAggregate, error) {
	var (
		firstObserved pgtype.Timestamptz
		lastObserved  pgtype.Timestamptz
		count         int
	)
	if err := tx.QueryRow(ctx, `
SELECT MIN(created_at), MAX(created_at), COUNT(*)
 FROM indicator_observations
 WHERE resolved_indicator_record_id = $1
   AND resolution_status = 'resolved'
   AND deleted_at IS NULL
`, recordID).Scan(&firstObserved, &lastObserved, &count); err != nil {
		return indicatorObservationAggregate{}, fmt.Errorf("load indicator observation aggregate: %w", err)
	}
	return indicatorObservationAggregate{
		FirstObservedAt:  timePointerFromPG(firstObserved),
		LastObservedAt:   timePointerFromPG(lastObserved),
		ObservationCount: count,
	}, nil
}

func loadIndicatorLifecycleSummaryTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (*string, error) {
	var summary pgtype.Text
	if err := tx.QueryRow(ctx, `
SELECT lifecycle_state
 FROM indicator_state_intervals
 WHERE indicator_record_id = $1
   AND deleted_at IS NULL
 ORDER BY CASE WHEN valid_to IS NULL THEN 0 ELSE 1 END ASC, valid_from DESC, indicator_state_interval_id DESC
 LIMIT 1
`, recordID).Scan(&summary); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("load indicator lifecycle summary: %w", err)
	}
	return textPointer(summary), nil
}

func loadIndicatorSupportingLinkCountTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int, error) {
	var count int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
  FROM active_record_links_v1
 WHERE dst_record_id = $1
`, recordID).Scan(&count); err != nil {
		return 0, fmt.Errorf("load indicator supporting link count: %w", err)
	}
	return count, nil
}

func loadIndicatorRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IndicatorRecord, error) {
	record, err := scanIndicatorRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id,
    r.deleted_at,
    r.deleted_by_user_id
  FROM indicators i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorRecord{}, ErrIndicatorNotFound
	}
	if err != nil {
		return IndicatorRecord{}, fmt.Errorf("load indicator record: %w", err)
	}
	return record, nil
}

func validateIndicatorRecordIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE record_id = $1
       AND incident_id = $2
       AND record_type = 'indicator'
       AND deleted_at IS NULL
)
`, recordID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate indicator record incident: %w", err)
	}
	if !exists {
		return ErrIndicatorNotFound
	}
	return nil
}

func insertIndicatorObservationTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, params IndicatorObservationCreateParams, createdAt time.Time) (IndicatorObservationRecord, error) {
	if params.IncidentID == uuid.Nil || params.SourceRecordID == uuid.Nil {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if strings.TrimSpace(params.SourceFieldKey) == "" || strings.TrimSpace(params.OriginKind) == "" || strings.TrimSpace(params.OriginLocator) == "" {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	observedText, ok := fieldnorm.NormalizeLine(params.ObservedText)
	if !ok {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if err := validateTimelineSourceIncidentTx(ctx, tx, params.IncidentID, params.SourceRecordID); err != nil {
		return IndicatorObservationRecord{}, err
	}
	parsedIndicatorType, normalizedCandidate, err := identity.NormalizeObservationCandidate(params.ParsedIndicatorType, params.NormalizedCandidate, observedText)
	if err != nil {
		return IndicatorObservationRecord{}, err
	}
	resolutionStatus := "unresolved"
	var resolvedByUserID *uuid.UUID
	var resolvedAt *time.Time
	resolutionMethod := params.ResolutionMethod
	if params.ResolvedIndicatorRecordID != nil {
		if err := validateIndicatorRecordIncidentTx(ctx, tx, params.IncidentID, *params.ResolvedIndicatorRecordID); err != nil {
			return IndicatorObservationRecord{}, err
		}
		resolutionStatus = "resolved"
		resolvedByUserID = &actorUserID
		value := createdAt.UTC()
		resolvedAt = &value
		if resolutionMethod == nil {
			value := observationCreateSource
			resolutionMethod = &value
		}
	}
	record := IndicatorObservationRecord{
		IncidentID:                params.IncidentID,
		SourceRecordID:            params.SourceRecordID,
		SourceFieldKey:            params.SourceFieldKey,
		OriginKind:                params.OriginKind,
		OriginLocator:             params.OriginLocator,
		ObservedText:              observedText,
		ParsedIndicatorType:       parsedIndicatorType,
		NormalizedCandidate:       normalizedCandidate,
		ResolutionStatus:          resolutionStatus,
		ResolvedIndicatorRecordID: params.ResolvedIndicatorRecordID,
		RowVersion:                1,
		CreatedByUserID:           actorUserID,
		CreatedAt:                 createdAt,
		ResolvedByUserID:          resolvedByUserID,
		ResolvedAt:                resolvedAt,
		ResolutionMethod:          resolutionMethod,
	}
	if err := tx.QueryRow(ctx, `
INSERT INTO indicator_observations (
    incident_id,
    source_record_id,
    source_field_key,
    origin_kind,
    origin_locator,
    observed_text,
    parsed_indicator_type,
    normalized_candidate,
    resolution_status,
    resolved_indicator_record_id,
    row_version,
    created_by_user_id,
    created_at,
    resolved_by_user_id,
    resolved_at,
    resolution_method
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, $11, $12, $13, $14, $15)
RETURNING indicator_observation_id
`, record.IncidentID, record.SourceRecordID, record.SourceFieldKey, record.OriginKind, record.OriginLocator, record.ObservedText, record.ParsedIndicatorType, record.NormalizedCandidate, record.ResolutionStatus, record.ResolvedIndicatorRecordID, record.CreatedByUserID, record.CreatedAt.UTC(), record.ResolvedByUserID, record.ResolvedAt, record.ResolutionMethod).Scan(&record.ObservationID); err != nil {
		return IndicatorObservationRecord{}, fmt.Errorf("insert indicator observation: %w", err)
	}
	return record, nil
}

func loadIndicatorObservationTx(ctx context.Context, tx pgx.Tx, observationID uuid.UUID, forUpdate bool) (IndicatorObservationRecord, error) {
	query := `
SELECT
    indicator_observation_id,
    incident_id,
    source_record_id,
    source_field_key,
    origin_kind,
    origin_locator,
    observed_text,
    parsed_indicator_type,
    normalized_candidate,
    resolution_status,
    resolved_indicator_record_id,
    row_version,
    created_by_user_id,
    created_at,
    resolved_by_user_id,
    resolved_at,
    resolution_method,
    deleted_at,
    deleted_by_user_id
  FROM indicator_observations
 WHERE indicator_observation_id = $1
   AND deleted_at IS NULL`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	record, err := scanIndicatorObservationRecord(tx.QueryRow(ctx, query, observationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorObservationRecord{}, ErrIndicatorObservationNotFound
	}
	if err != nil {
		return IndicatorObservationRecord{}, fmt.Errorf("load indicator observation: %w", err)
	}
	return record, nil
}

func insertIndicatorLifecycleIntervalTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, params IndicatorLifecycleAppendParams, createdAt time.Time) (IndicatorLifecycleIntervalRecord, error) {
	if strings.TrimSpace(params.LifecycleState) == "" {
		return IndicatorLifecycleIntervalRecord{}, ErrInvalidCreateRequest
	}
	record := IndicatorLifecycleIntervalRecord{
		IncidentID:        params.IncidentID,
		IndicatorRecordID: params.IndicatorRecordID,
		LifecycleState:    strings.TrimSpace(params.LifecycleState),
		ValidFrom:         params.ValidFrom.UTC(),
		ValidTo:           normalizeTimePointer(params.ValidTo),
		Confidence:        cloneIntPointer(params.Confidence),
		Rationale:         cloneStringPointer(params.Rationale),
		SupportRefs:       append([]string(nil), params.SupportRefs...),
		Assessor:          cloneStringPointer(params.Assessor),
		AssessedAt:        createdAt,
		RowVersion:        1,
		CreatedByUserID:   actorUserID,
		CreatedAt:         createdAt,
	}
	if record.ValidFrom.IsZero() {
		record.ValidFrom = createdAt
	}
	if err := tx.QueryRow(ctx, `
INSERT INTO indicator_state_intervals (
    incident_id,
    indicator_record_id,
    lifecycle_state,
    valid_from,
    valid_to,
    confidence,
    rationale,
    support_refs,
    assessor,
    assessed_at,
    row_version,
    created_by_user_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, 1, $11, $12)
RETURNING indicator_state_interval_id
`, record.IncidentID, record.IndicatorRecordID, record.LifecycleState, record.ValidFrom.UTC(), record.ValidTo, record.Confidence, record.Rationale, mustJSON(record.SupportRefs), record.Assessor, record.AssessedAt.UTC(), record.CreatedByUserID, record.CreatedAt.UTC()).Scan(&record.IntervalID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, fmt.Errorf("insert indicator lifecycle interval: %w", err)
	}
	return record, nil
}

func validateTimelineSourceIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceRecordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE record_id = $1
       AND incident_id = $2
)
`, sourceRecordID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate source record incident: %w", err)
	}
	if !exists {
		return revisions.ErrRecordDeletedUseRestore
	}
	return nil
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func normalizeIndicatorValue(indicatorType string, rawDisplay string, rawNormalized *string) (string, *string, error) {
	return identity.NormalizeValue(indicatorType, rawDisplay, rawNormalized)
}

func buildIndicatorDedupeKey(input indicatorUpsertInput) string {
	return identity.DedupeKey(input)
}

func buildIndicatorObservationValue(record IndicatorObservationRecord) map[string]any {
	return map[string]any{
		"indicator_observation_id":     record.ObservationID.String(),
		"incident_id":                  record.IncidentID.String(),
		"source_record_id":             record.SourceRecordID.String(),
		"source_field_key":             record.SourceFieldKey,
		"origin_kind":                  record.OriginKind,
		"origin_locator":               record.OriginLocator,
		"observed_text":                record.ObservedText,
		"parsed_indicator_type":        derefString(record.ParsedIndicatorType),
		"normalized_candidate":         derefString(record.NormalizedCandidate),
		"resolution_status":            record.ResolutionStatus,
		"resolved_indicator_record_id": formatUUIDPointer(record.ResolvedIndicatorRecordID),
		"row_version":                  record.RowVersion,
		"created_by_user_id":           record.CreatedByUserID.String(),
		"created_at":                   formatTimestamp(record.CreatedAt),
		"resolved_by_user_id":          formatUUIDPointer(record.ResolvedByUserID),
		"resolved_at":                  formatTimestampPointer(record.ResolvedAt),
		"resolution_method":            derefString(record.ResolutionMethod),
		"deleted_at":                   formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":           formatUUIDPointer(record.DeletedByUserID),
	}
}

func buildIndicatorLifecycleValue(record IndicatorLifecycleIntervalRecord) map[string]any {
	return map[string]any{
		"indicator_state_interval_id": record.IntervalID.String(),
		"incident_id":                 record.IncidentID.String(),
		"indicator_record_id":         record.IndicatorRecordID.String(),
		"lifecycle_state":             record.LifecycleState,
		"valid_from":                  formatTimestamp(record.ValidFrom),
		"valid_to":                    formatTimestampPointer(record.ValidTo),
		"confidence":                  derefInt(record.Confidence),
		"rationale":                   derefString(record.Rationale),
		"support_refs":                append([]string(nil), record.SupportRefs...),
		"assessor":                    derefString(record.Assessor),
		"assessed_at":                 formatTimestamp(record.AssessedAt),
		"row_version":                 record.RowVersion,
		"created_by_user_id":          record.CreatedByUserID.String(),
		"created_at":                  formatTimestamp(record.CreatedAt),
		"deleted_at":                  formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":          formatUUIDPointer(record.DeletedByUserID),
	}
}

func scanIndicatorRecord(scanner interface{ Scan(dest ...any) error }) (IndicatorRecord, error) {
	var (
		record           IndicatorRecord
		rawRecordID      pgtype.UUID
		rawIncidentID    pgtype.UUID
		rawNormalized    pgtype.Text
		rawDefanged      pgtype.Text
		rawHashAlgorithm pgtype.Text
		rawHashValue     pgtype.Text
		rawSTIXPattern   pgtype.Text
		rawCreatedBy     pgtype.UUID
		rawUpdatedBy     pgtype.UUID
		rawDeletedAt     pgtype.Timestamptz
		rawDeletedBy     pgtype.UUID
	)
	if err := scanner.Scan(
		&rawRecordID,
		&rawIncidentID,
		&record.IndicatorType,
		&record.ValueKind,
		&record.DisplayValue,
		&rawNormalized,
		&record.DedupeKey,
		&rawDefanged,
		&rawHashAlgorithm,
		&rawHashValue,
		&rawSTIXPattern,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&rawCreatedBy,
		&rawUpdatedBy,
		&rawDeletedAt,
		&rawDeletedBy,
	); err != nil {
		return IndicatorRecord{}, err
	}
	record.RecordID = uuid.Must(uuid.FromBytes(rawRecordID.Bytes[:]))
	record.IncidentID = uuid.Must(uuid.FromBytes(rawIncidentID.Bytes[:]))
	record.NormalizedValue = textPointer(rawNormalized)
	record.DefangedValue = textPointer(rawDefanged)
	record.HashAlgorithm = textPointer(rawHashAlgorithm)
	record.HashValue = textPointer(rawHashValue)
	record.STIXPattern = textPointer(rawSTIXPattern)
	record.CreatedByUser = uuid.Must(uuid.FromBytes(rawCreatedBy.Bytes[:]))
	record.UpdatedByUser = uuid.Must(uuid.FromBytes(rawUpdatedBy.Bytes[:]))
	record.DeletedAt = timePointerFromPG(rawDeletedAt)
	record.DeletedByUserID = uuidPointerFromPG(rawDeletedBy)
	return record, nil
}

func scanIndicatorObservationRecord(scanner interface{ Scan(dest ...any) error }) (IndicatorObservationRecord, error) {
	var (
		record              IndicatorObservationRecord
		rawObservationID    pgtype.UUID
		rawIncidentID       pgtype.UUID
		rawSourceRecordID   pgtype.UUID
		rawParsedType       pgtype.Text
		rawNormalized       pgtype.Text
		rawResolvedID       pgtype.UUID
		rawCreatedBy        pgtype.UUID
		rawResolvedBy       pgtype.UUID
		rawResolvedAt       pgtype.Timestamptz
		rawResolutionMethod pgtype.Text
		rawDeletedAt        pgtype.Timestamptz
		rawDeletedBy        pgtype.UUID
	)
	if err := scanner.Scan(
		&rawObservationID,
		&rawIncidentID,
		&rawSourceRecordID,
		&record.SourceFieldKey,
		&record.OriginKind,
		&record.OriginLocator,
		&record.ObservedText,
		&rawParsedType,
		&rawNormalized,
		&record.ResolutionStatus,
		&rawResolvedID,
		&record.RowVersion,
		&rawCreatedBy,
		&record.CreatedAt,
		&rawResolvedBy,
		&rawResolvedAt,
		&rawResolutionMethod,
		&rawDeletedAt,
		&rawDeletedBy,
	); err != nil {
		return IndicatorObservationRecord{}, err
	}
	record.ObservationID = uuid.Must(uuid.FromBytes(rawObservationID.Bytes[:]))
	record.IncidentID = uuid.Must(uuid.FromBytes(rawIncidentID.Bytes[:]))
	record.SourceRecordID = uuid.Must(uuid.FromBytes(rawSourceRecordID.Bytes[:]))
	record.ParsedIndicatorType = textPointer(rawParsedType)
	record.NormalizedCandidate = textPointer(rawNormalized)
	record.ResolvedIndicatorRecordID = uuidPointerFromPG(rawResolvedID)
	record.CreatedByUserID = uuid.Must(uuid.FromBytes(rawCreatedBy.Bytes[:]))
	record.ResolvedByUserID = uuidPointerFromPG(rawResolvedBy)
	record.ResolvedAt = timePointerFromPG(rawResolvedAt)
	record.ResolutionMethod = textPointer(rawResolutionMethod)
	record.DeletedAt = timePointerFromPG(rawDeletedAt)
	record.DeletedByUserID = uuidPointerFromPG(rawDeletedBy)
	return record, nil
}

func timePointerFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func derefInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func jsonEqual(left map[string]any, right map[string]any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return bytes.Equal(leftJSON, rightJSON)
}

func derefStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
