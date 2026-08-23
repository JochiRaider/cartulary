package indicators

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	ViewSchemaID = "cartulary.view.indicators.v1"

	IndicatorFindOrCreateParticipantV1 = "indicator_find_or_create_participant_v1"

	observationCreateSource  = "indicators.observations.capture"
	observationResolveSource = "indicators.observations.resolve"
	lifecycleAppendSource    = "indicators.lifecycle.append"
)

var (
	ErrInvalidCreateRequest         = errors.New("indicators: invalid create request")
	ErrSourceTextUnavailable        = errors.New("indicators: source text unavailable")
	ErrIndicatorNotFound            = errors.New("indicators: indicator not found")
	ErrIndicatorObservationNotFound = errors.New("indicators: indicator observation not found")
	ErrIndicatorSourceNotFound      = errors.New("indicators: source record not found")
	ErrResolvedIndicatorNotFound    = errors.New("indicators: resolved Indicator not found")
	ErrRowVersionConflict           = errors.New("indicators: row version conflict")
	ErrIllegalTransition            = errors.New("indicators: illegal transition")
)

// SourceTextPort is the narrow transaction-visible read boundary used by
// manual observation admission. Its implementation resolves the owning view
// contract and returns the canonical projected row with the exact text value.
type SourceTextPort interface {
	LoadTextTx(context.Context, pgx.Tx, uuid.UUID, string, string) (SourceTextValue, error)
	LoadRowTx(context.Context, pgx.Tx, uuid.UUID, string, string) (map[string]any, error)
	RefreshAndLoadRowTx(context.Context, pgx.Tx, uuid.UUID, string, string) (map[string]any, error)
}

type SourceTextValue struct {
	ViewSchemaID string
	Text         string
	Row          map[string]any
}

type CreateCommand struct {
	ClientTxnID     string
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
}

type CreateResult struct {
	Created      bool
	Replayed     bool
	CanonicalRow map[string]any
	RecordID     uuid.UUID
	ChangeSetID  uuid.UUID
	RowVersion   int64
}

type IndicatorFindOrCreateParticipantCommand struct {
	IncidentID        uuid.UUID
	Actor             authn.UserRecord
	IndicatorType     string
	ValueKind         string
	DisplayValue      string
	NormalizedValue   *string
	OperationContext  string
	OperationOccurred time.Time
}

type IndicatorFindOrCreateParticipantResult struct {
	SchemaID  string
	Status    string
	Indicator IndicatorReference
}

// IndicatorReference is the immutable Indicator identity contract exposed to
// transaction participants. Envelope state and optional representations remain
// owner-internal so consumers cannot couple to unrelated persistence details.
type IndicatorReference struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
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

func ValidateCreateCommand(command CreateCommand) error {
	_, err := indicatorInputFromCreateCommand(command)
	return err
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
