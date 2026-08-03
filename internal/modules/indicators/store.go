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
	afterRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, record.RecordID)
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
	record, err := s.observations.insertTx(ctx, tx, actor.ID, params, createdAt)
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
		if err := s.refreshProjectionRowTx(ctx, tx, *record.ResolvedIndicatorRecordID); err != nil {
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

	current, err := s.observations.loadTx(ctx, tx, params.ObservationID, true)
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, current.IncidentID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.sources.validateIncidentTx(ctx, tx, current.IncidentID, params.ResolvedIndicatorRecordID); err != nil {
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
	if err := s.observations.resolveTx(ctx, tx, next); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
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
		if err := s.refreshProjectionRowTx(ctx, tx, recordID); err != nil {
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

	if err := s.sources.validateIncidentTx(ctx, tx, params.IncidentID, params.IndicatorRecordID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	createdAt := params.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	record, err := s.lifecycles.insertTx(ctx, tx, actor.ID, params, createdAt)
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
	if err := s.refreshProjectionRowTx(ctx, tx, params.IndicatorRecordID); err != nil {
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
	current, matched, err := s.sources.loadByDedupeTx(ctx, tx, incidentID, input.IndicatorType, input.DedupeKey)
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
		if err := s.sources.insertTx(ctx, tx, &record); err != nil {
			return IndicatorRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	beforeRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, current.RecordID)
	if err != nil {
		return IndicatorRecord{}, nil, "", 0, err
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
			return IndicatorRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := s.sources.updateTx(ctx, tx, next); err != nil {
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

func (s *Store) refreshProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.projections.RefreshRowTx(ctx, tx, ViewSchemaID, recordID)
}

func (s *Store) loadProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return s.projections.LoadRowTx(ctx, tx, ViewSchemaID, recordID)
}

func (s *Store) refreshAndLoadProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	if err := s.refreshProjectionRowTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return s.loadProjectionRowTx(ctx, tx, recordID)
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
