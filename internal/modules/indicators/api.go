package indicators

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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

var ErrInvalidCreateRequest = errors.New("indicators: invalid create request")

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidentLifecycleAccess
	recordStore    *records.Store
	revisionsStore revisionAppendPort
	sources        sourceRepository
	observations   observationRepository
	lifecycles     lifecycleRepository
}

type StoreDependencies struct {
	Postgres  postgres.DB
	Revisions *revisions.Appender
}

func NewStore(dependencies StoreDependencies) (*Store, error) {
	if dependencies.Postgres == nil {
		return nil, fmt.Errorf("compose Indicators store: Postgres is required")
	}
	if dependencies.Revisions == nil {
		return nil, fmt.Errorf("compose Indicators store: Revisions is required")
	}
	return &Store{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: newIncidentLifecycleAccess(dependencies.Postgres),
		recordStore:    records.NewStore(),
		revisionsStore: newRevisionAppendAdapter(dependencies.Revisions),
		sources:        sourceRepository{},
		observations:   observationRepository{},
		lifecycles:     lifecycleRepository{},
	}, nil
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
	Indicator IndicatorRecord
}

type IndicatorRecord struct {
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

type IndicatorProjectionRecord struct {
	IndicatorRecord
	FirstObservedAt   *time.Time
	LastObservedAt    *time.Time
	ObservationCount  int
	LifecycleSummary  *string
	SupportingLinkCnt int
}

type MutationResult struct {
	Payload     map[string]any
	StatusCode  int
	Replayed    bool
	RecordID    uuid.UUID
	ChangeSetID uuid.UUID
	RowVersion  int64
}

func BuildIndicatorRow(record IndicatorProjectionRecord) map[string]any {
	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells": map[string]any{
			"indicator.indicator_type":    map[string]any{"value": record.IndicatorType},
			"indicator.value_kind":        map[string]any{"value": record.ValueKind},
			"indicator.display_value":     map[string]any{"value": record.DisplayValue},
			"indicator.normalized_value":  map[string]any{"value": derefString(record.NormalizedValue)},
			"indicator.defanged_value":    map[string]any{"value": derefString(record.DefangedValue)},
			"indicator.hash_algorithm":    map[string]any{"value": derefString(record.HashAlgorithm)},
			"indicator.hash_value":        map[string]any{"value": derefString(record.HashValue)},
			"indicator.stix_pattern":      map[string]any{"value": derefString(record.STIXPattern)},
			"indicator.first_observed_at": map[string]any{"value": formatTimestampPointer(record.FirstObservedAt)},
			"indicator.last_observed_at":  map[string]any{"value": formatTimestampPointer(record.LastObservedAt)},
			"indicator.observation_count": map[string]any{"value": record.ObservationCount},
			"indicator.lifecycle_summary": map[string]any{"value": derefString(record.LifecycleSummary)},
			"indicator.supporting_link_count": map[string]any{
				"value": record.SupportingLinkCnt,
			},
		},
	}
	row["group_values"] = map[string]any{
		"indicator.indicator_type":    record.IndicatorType,
		"indicator.value_kind":        record.ValueKind,
		"indicator.lifecycle_summary": derefString(record.LifecycleSummary),
	}
	return row
}

func BuildMutationPayload(changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": ViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
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
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}
