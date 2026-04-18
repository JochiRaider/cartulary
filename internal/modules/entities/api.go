package entities

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	HostsViewSchemaID      = "cartulary.view.hosts.v1"
	IdentitiesViewSchemaID = "cartulary.view.identities.v1"

	hostCreateRouteKey     = "entities.hosts.rows.create"
	identityCreateRouteKey = "entities.identities.rows.create"
)

type CreateRequest struct {
	ClientTxnID string
	Values      map[string]string
}

type HostRecord struct {
	RecordID      uuid.UUID
	IncidentID    uuid.UUID
	DisplayName   string
	Hostname      *string
	HostState     string
	RowVersion    int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedByUser uuid.UUID
	UpdatedByUser uuid.UUID
}

type IdentityRecord struct {
	RecordID       uuid.UUID
	IncidentID     uuid.UUID
	DisplayName    string
	UPN            *string
	Email          *string
	SamAccountName *string
	IdentityState  string
	RowVersion     int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedByUser  uuid.UUID
	UpdatedByUser  uuid.UUID
}

type MutationResult struct {
	Payload     map[string]any
	StatusCode  int
	Replayed    bool
	RecordID    uuid.UUID
	ChangeSetID uuid.UUID
	RowVersion  int64
}

func DecodeCreateRequest(viewSchemaID string, reader io.Reader) (CreateRequest, *auth.APIError) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable {
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
		if !field.Writable {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "readonly_field")
		}
		if field.EntityBindingMode == nil || *field.EntityBindingMode != "entity_origin" {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "unsupported_field_key")
		}
		if field.ConflictResolutionClass == "collection_review" {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "unsupported_field_key")
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

	if len(request.Values) == 0 && !schema.PermitsZeroFieldCreate {
		return CreateRequest{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}

	return request, nil
}

func CreateRequestHash(viewSchemaID string, request CreateRequest) []byte {
	keys := make([]string, 0, len(request.Values))
	for key := range request.Values {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	payload := map[string]any{
		"view_schema_id": viewSchemaID,
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

func BuildHostRow(record HostRecord) map[string]any {
	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells": map[string]any{
			"host.display_name":       map[string]any{"value": record.DisplayName},
			"host.hostname":           map[string]any{"value": derefString(record.Hostname)},
			"host.aliases":            map[string]any{"value": collectionValue(false, nil)},
			"host.host_state":         map[string]any{"value": record.HostState},
			"host.linked_event_count": map[string]any{"value": 0},
			"host.evidence_count":     map[string]any{"value": 0},
			"host.location":           map[string]any{"value": nil},
			"host.os_platform":        map[string]any{"value": nil},
			"host.business_owner":     map[string]any{"value": nil},
			"host.criticality":        map[string]any{"value": nil},
			"host.containment_status": map[string]any{"value": nil},
			"host.edited_at":          map[string]any{"value": formatTimestamp(record.UpdatedAt)},
		},
	}
	row["group_values"] = map[string]any{
		"host.host_state":         record.HostState,
		"host.criticality":        nil,
		"host.containment_status": nil,
	}
	return row
}

func BuildIdentityRow(record IdentityRecord) map[string]any {
	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells": map[string]any{
			"identity.display_name":       map[string]any{"value": record.DisplayName},
			"identity.upn":                map[string]any{"value": derefString(record.UPN)},
			"identity.email":              map[string]any{"value": derefString(record.Email)},
			"identity.sam_account_name":   map[string]any{"value": derefString(record.SamAccountName)},
			"identity.aliases":            map[string]any{"value": collectionValue(false, nil)},
			"identity.identity_state":     map[string]any{"value": record.IdentityState},
			"identity.linked_event_count": map[string]any{"value": 0},
			"identity.evidence_count":     map[string]any{"value": 0},
			"identity.privilege_level":    map[string]any{"value": nil},
			"identity.mfa_state":          map[string]any{"value": nil},
			"identity.reset_status":       map[string]any{"value": nil},
			"identity.edited_at":          map[string]any{"value": formatTimestamp(record.UpdatedAt)},
		},
	}
	row["group_values"] = map[string]any{
		"identity.identity_state":  record.IdentityState,
		"identity.privilege_level": nil,
		"identity.mfa_state":       nil,
		"identity.reset_status":    nil,
	}
	return row
}

func BuildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": viewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func incidentNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *auth.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &auth.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: details,
	}
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func requiredRoleDescription(roles ...string) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return roles[0]
	}
	return strings.Join(roles, "|")
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func collectionValue(ordered bool, items []map[string]any) map[string]any {
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   items,
	}
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
