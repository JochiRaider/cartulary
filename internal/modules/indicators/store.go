package indicators

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
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
	ErrIndicatorSourceNotFound      = errors.New("indicators: source record not found")
	ErrResolvedIndicatorNotFound    = errors.New("indicators: resolved Indicator not found")
	ErrRowVersionConflict           = errors.New("indicators: row version conflict")
	ErrIllegalTransition            = errors.New("indicators: illegal transition")
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
	BaseRowVersion            int64
	SourceFieldKey            string
	SpanStartByte             int
	SpanEndByte               int
	ParsedIndicatorType       *string
	ResolvedIndicatorRecordID *uuid.UUID
	RequestID                 string
	ClientTxnID               string
	RequestHash               []byte
	originKind                indicatororigin.ObservationOrigin
	originLocator             string
	observedText              string
	normalizedCandidate       *string
}

type IndicatorObservationResolveParams struct {
	ObservationID             uuid.UUID
	ResolvedIndicatorRecordID uuid.UUID
	BaseRowVersion            int64
	RequestID                 string
	ClientTxnID               string
	RequestHash               []byte
}

type IndicatorObservationActionParams struct {
	ObservationID  uuid.UUID
	BaseRowVersion int64
	RequestID      string
	ClientTxnID    string
	RequestHash    []byte
}

type IndicatorLifecycleAppendParams struct {
	IncidentID        uuid.UUID
	IndicatorRecordID uuid.UUID
	BaseRowVersion    int64
	LifecycleState    string
	ValidFrom         time.Time
	ValidTo           *time.Time
	Confidence        *int
	Rationale         *string
	SupportRefs       []uuid.UUID
	Assessor          *string
	RequestID         string
	ClientTxnID       string
	RequestHash       []byte
}

type IndicatorObservationRecord struct {
	ObservationID             uuid.UUID  `json:"observation_id"`
	IncidentID                uuid.UUID  `json:"incident_id"`
	SourceRecordID            uuid.UUID  `json:"source_record_id"`
	SourceFieldKey            string     `json:"source_field_key"`
	OriginKind                string     `json:"origin_kind"`
	OriginLocator             string     `json:"origin_locator"`
	ObservedText              string     `json:"observed_text"`
	ParsedIndicatorType       *string    `json:"parsed_indicator_type"`
	NormalizedCandidate       *string    `json:"normalized_candidate"`
	ResolutionStatus          string     `json:"resolution_status"`
	ResolvedIndicatorRecordID *uuid.UUID `json:"resolved_indicator_record_id"`
	RowVersion                int64      `json:"row_version"`
	CreatedByUserID           uuid.UUID  `json:"created_by_user_id"`
	CreatedAt                 time.Time  `json:"created_at"`
	ResolvedByUserID          *uuid.UUID `json:"resolved_by_user_id"`
	ResolvedAt                *time.Time `json:"resolved_at"`
	ResolutionMethod          *string    `json:"resolution_method"`
	DeletedAt                 *time.Time `json:"-"`
	DeletedByUserID           *uuid.UUID `json:"-"`
}

type IndicatorLifecycleIntervalRecord struct {
	IntervalID        uuid.UUID   `json:"interval_id"`
	IncidentID        uuid.UUID   `json:"incident_id"`
	IndicatorRecordID uuid.UUID   `json:"indicator_record_id"`
	LifecycleState    string      `json:"lifecycle_state"`
	ValidFrom         time.Time   `json:"valid_from"`
	ValidTo           *time.Time  `json:"valid_to"`
	Confidence        *int        `json:"confidence"`
	Rationale         *string     `json:"rationale"`
	SupportRefs       []uuid.UUID `json:"support_refs"`
	Assessor          *string     `json:"assessor"`
	AssessedAt        time.Time   `json:"assessed_at"`
	RowVersion        int64       `json:"row_version"`
	CreatedByUserID   uuid.UUID   `json:"created_by_user_id"`
	CreatedAt         time.Time   `json:"created_at"`
	DeletedAt         *time.Time  `json:"-"`
	DeletedByUserID   *uuid.UUID  `json:"-"`
}

type AffectedRecordVersion struct {
	RecordID   uuid.UUID `json:"record_id"`
	RowVersion int64     `json:"row_version"`
}

type IndicatorObservationMutationResult struct {
	Observation     IndicatorObservationRecord `json:"observation"`
	ChangeSetID     uuid.UUID                  `json:"change_set_id"`
	Replayed        bool                       `json:"replayed"`
	AffectedRecords []AffectedRecordVersion    `json:"affected_records"`
}

type IndicatorLifecycleMutationResult struct {
	Interval        IndicatorLifecycleIntervalRecord `json:"interval"`
	ChangeSetID     uuid.UUID                        `json:"change_set_id"`
	Replayed        bool                             `json:"replayed"`
	AffectedRecords []AffectedRecordVersion          `json:"affected_records"`
}

type indicatorUpsertInput = identity.Canonical

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
		"support_refs":                indicatorUUIDStrings(record.SupportRefs),
		"assessor":                    derefString(record.Assessor),
		"assessed_at":                 formatTimestamp(record.AssessedAt),
		"row_version":                 record.RowVersion,
		"created_by_user_id":          record.CreatedByUserID.String(),
		"created_at":                  formatTimestamp(record.CreatedAt),
		"deleted_at":                  formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":          formatUUIDPointer(record.DeletedByUserID),
	}
}

func indicatorUUIDStrings(values []uuid.UUID) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = value.String()
	}
	return result
}

func scanIndicatorRecord(scanner interface{ Scan(dest ...any) error }) (indicatorRecord, error) {
	var (
		record           indicatorRecord
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
		return indicatorRecord{}, err
	}
	if !rawRecordID.Valid || !rawIncidentID.Valid || !rawCreatedBy.Valid || !rawUpdatedBy.Valid {
		return indicatorRecord{}, fmt.Errorf("scan indicator: required UUID is NULL")
	}
	record.RecordID = uuid.UUID(rawRecordID.Bytes)
	record.IncidentID = uuid.UUID(rawIncidentID.Bytes)
	record.NormalizedValue = textPointer(rawNormalized)
	record.DefangedValue = textPointer(rawDefanged)
	record.HashAlgorithm = textPointer(rawHashAlgorithm)
	record.HashValue = textPointer(rawHashValue)
	record.STIXPattern = textPointer(rawSTIXPattern)
	record.CreatedByUser = uuid.UUID(rawCreatedBy.Bytes)
	record.UpdatedByUser = uuid.UUID(rawUpdatedBy.Bytes)
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
		rawOriginKind       string
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
		&rawOriginKind,
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
	originKind, err := indicatororigin.Parse(rawOriginKind)
	if err != nil {
		return IndicatorObservationRecord{}, err
	}
	record.OriginKind = originKind.String()
	if !rawObservationID.Valid || !rawIncidentID.Valid || !rawSourceRecordID.Valid || !rawCreatedBy.Valid {
		return IndicatorObservationRecord{}, fmt.Errorf("scan indicator observation: required UUID is NULL")
	}
	record.ObservationID = uuid.UUID(rawObservationID.Bytes)
	record.IncidentID = uuid.UUID(rawIncidentID.Bytes)
	record.SourceRecordID = uuid.UUID(rawSourceRecordID.Bytes)
	record.ParsedIndicatorType = textPointer(rawParsedType)
	record.NormalizedCandidate = textPointer(rawNormalized)
	record.ResolvedIndicatorRecordID = uuidPointerFromPG(rawResolvedID)
	record.CreatedByUserID = uuid.UUID(rawCreatedBy.Bytes)
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

func jsonEqual(left map[string]any, right map[string]any) bool {
	return reflect.DeepEqual(left, right)
}
