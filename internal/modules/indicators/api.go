package indicators

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	ViewSchemaID = "cartulary.view.indicators.v1"

	indicatorCreateRouteKey = "indicators.rows.create"

	IndicatorFindOrCreateParticipantV1 = "indicator_find_or_create_participant_v1"
)

const (
	httpStatusCreated = 201
	httpStatusOK      = 200
)

var (
	ErrInvalidCreateRequest  = errors.New("indicators: invalid create request")
	ErrSourceTextUnavailable = errors.New("indicators: source text unavailable")
)

type Store struct {
	pool               postgres.DB
	authStore          *authn.Store
	incidentAccess     incidentLifecycleAccess
	recordStore        *records.Store
	revisionsStore     revisionAppendPort
	projections        ProjectionPort
	sourceText         SourceTextPort
	now                func() time.Time
	sources            sourceRepository
	observations       observationRepository
	lifecycles         lifecycleRepository
	createService      indicatorCreateService
	observationService indicatorObservationService
	lifecycleService   indicatorLifecycleService
}

type StoreDependencies struct {
	Postgres    postgres.DB
	Revisions   *revisions.Appender
	Projections ProjectionPort
	SourceText  SourceTextPort
	Clock       func() time.Time
}

type ProjectionPort interface {
	RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error
	LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

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

func NewStore(dependencies StoreDependencies) (*Store, error) {
	if dependencies.Postgres == nil {
		return nil, fmt.Errorf("compose Indicators store: Postgres is required")
	}
	if dependencies.Revisions == nil {
		return nil, fmt.Errorf("compose Indicators store: Revisions is required")
	}
	if dependencies.Projections == nil {
		return nil, fmt.Errorf("compose Indicators store: Projections is required")
	}
	if dependencies.SourceText == nil {
		return nil, fmt.Errorf("compose Indicators store: SourceText is required")
	}
	now := dependencies.Clock
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	store := &Store{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: newIncidentLifecycleAccess(dependencies.Postgres),
		recordStore:    records.NewStore(),
		revisionsStore: newRevisionAppendAdapter(dependencies.Revisions),
		projections:    dependencies.Projections,
		sourceText:     dependencies.SourceText,
		now:            now,
		sources:        sourceRepository{},
		observations:   observationRepository{},
		lifecycles:     lifecycleRepository{},
	}
	store.createService = indicatorCreateService{owner: store}
	store.observationService = indicatorObservationService{owner: store}
	store.lifecycleService = indicatorLifecycleService{owner: store}
	return store, nil
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

func ValidateCreateCommand(command CreateCommand) error {
	_, err := indicatorInputFromCreateCommand(command)
	return err
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

type indicatorRecord struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
	RowVersion      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedByUser   uuid.UUID
	UpdatedByUser   uuid.UUID
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type CreateOutcome string

const (
	CreateOutcomeCreated  CreateOutcome = "created"
	CreateOutcomeReused   CreateOutcome = "reused"
	CreateOutcomeUpdated  CreateOutcome = "updated"
	CreateOutcomeReplayed CreateOutcome = "replayed"
)

type CreateResult struct {
	Outcome      CreateOutcome
	CanonicalRow map[string]any
	RecordID     uuid.UUID
	ChangeSetID  uuid.UUID
	RowVersion   int64
}

func buildStoredCreateResponse(changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": ViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}

func referenceFromRecord(record indicatorRecord) IndicatorReference {
	return IndicatorReference{
		RecordID: record.RecordID, IncidentID: record.IncidentID,
		IndicatorType: record.IndicatorType, ValueKind: record.ValueKind,
		DisplayValue:    record.DisplayValue,
		NormalizedValue: cloneStringPointer(record.NormalizedValue),
		DedupeKey:       record.DedupeKey,
	}
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func extractInt64FromPayload(payload map[string]any, path ...string) (int64, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	number, ok := current.(float64)
	if !ok || number < 1 || number != float64(int64(number)) {
		return 0, fmt.Errorf("decode payload integer path %q", strings.Join(path, "."))
	}
	return int64(number), nil
}

func entityVersionID(prefix string, recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("%s:%s:%d", prefix, recordID.String(), rowVersion)
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPointer(value string) *string {
	cloned := value
	return &cloned
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
}
