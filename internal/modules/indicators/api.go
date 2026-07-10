package indicators

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
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
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: newIncidentLifecycleAccess(pool),
		recordStore:    records.NewStore(),
		revisionsStore: newRevisionAppendAdapter(),
	}
}

type CreateRequest struct {
	ClientTxnID string
	Values      map[string]string
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

func DecodeCreateRequest(reader io.Reader) (CreateRequest, *httpapi.APIError) {
	schema, ok := viewschema.Lookup(ViewSchemaID)
	if !ok {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request CreateRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	request.Values = make(map[string]string)
	for fieldKey, field := range schema.Fields() {
		value, ok := raw[fieldKey]
		if !ok {
			continue
		}
		if !field.Writable && !field.CreateWritable {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "readonly_field")
		}
		if string(value) == "null" {
			if field.Clearable {
				continue
			}
			return CreateRequest{}, invalidMutationPayload(fieldKey, "field_not_nullable")
		}

		var rawValue string
		if err := json.Unmarshal(value, &rawValue); err != nil {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		normalized, ok := fieldnorm.NormalizeLine(rawValue)
		if !ok {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		request.Values[fieldKey] = normalized
	}

	if len(schema.MinimumCreateFieldSets) > 0 && !createMinimumSatisfied(schema.MinimumCreateFieldSets, request.Values) {
		return CreateRequest{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}
	return request, nil
}

func createMinimumSatisfied(fieldSets [][]string, values map[string]string) bool {
	for _, fieldSet := range fieldSets {
		setSatisfied := true
		for _, fieldKey := range fieldSet {
			if strings.TrimSpace(values[fieldKey]) == "" {
				setSatisfied = false
				break
			}
		}
		if setSatisfied {
			return true
		}
	}
	return false
}

func CreateRequestHash(request CreateRequest) []byte {
	keys := make([]string, 0, len(request.Values))
	for key := range request.Values {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	payload := map[string]any{
		"view_schema_id": ViewSchemaID,
		"client_txn_id":  request.ClientTxnID,
	}
	for _, key := range keys {
		payload[key] = request.Values[key]
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
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

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
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

func optionalValue(values map[string]string, key string) *string {
	value, ok := values[key]
	if !ok {
		return nil
	}
	cloned := value
	return &cloned
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
