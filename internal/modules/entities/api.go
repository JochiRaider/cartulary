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
	IndicatorsViewSchemaID = "cartulary.view.indicators.v1"

	hostCreateRouteKey      = "entities.hosts.rows.create"
	identityCreateRouteKey  = "entities.identities.rows.create"
	indicatorCreateRouteKey = "entities.indicators.rows.create"
	maxCollectionActions    = 64
)

type CreateRequest struct {
	ClientTxnID string
	Values      map[string]string
	AliasAdds   map[string][]CollectionAction
}

type HostRecord struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	DisplayName           string
	AADDeviceID           *string
	FQDN                  *string
	Hostname              *string
	HostState             string
	LinkedEventCount      int
	EvidenceCount         int
	Location              *string
	OSPlatform            *string
	BusinessOwner         *string
	Criticality           *string
	ContainmentStatus     *string
	MergedIntoRecordID    *uuid.UUID
	EntityOrigin          string
	SeedMentionID         *uuid.UUID
	SuggestionOnlyAliases []string
	ReusableIdentifiers   []ReusableIdentifier
	RowVersion            int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
	CreatedByUser         uuid.UUID
	UpdatedByUser         uuid.UUID
}

type IdentityRecord struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	DisplayName           string
	AADObjectID           *string
	SID                   *string
	UPN                   *string
	Email                 *string
	SamAccountName        *string
	IdentityState         string
	LinkedEventCount      int
	EvidenceCount         int
	PrivilegeLevel        *string
	MFAState              *string
	ResetStatus           *string
	MergedIntoRecordID    *uuid.UUID
	EntityOrigin          string
	SeedMentionID         *uuid.UUID
	SuggestionOnlyAliases []string
	ReusableIdentifiers   []ReusableIdentifier
	RowVersion            int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
	CreatedByUser         uuid.UUID
	UpdatedByUser         uuid.UUID
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

type ReusableIdentifier struct {
	EntityPreservedIdentifierID uuid.UUID
	IdentifierClass             string
	RawValue                    string
	NormalizedValue             string
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

type CollectionAction struct {
	Op             string
	RawText        string
	NormalizedText string
	ItemRef        string
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
	request.AliasAdds = make(map[string][]CollectionAction)
	for fieldKey, field := range schema.Fields() {
		value, ok := raw[fieldKey]
		if !ok {
			continue
		}
		if !field.Writable && !field.CreateWritable {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "readonly_field")
		}
		if field.ConflictResolutionClass == "collection_review" {
			actions, ok := decodeAliasActionPayload(fieldKey, value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.AliasAdds[fieldKey] = actions
			continue
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

	if len(schema.MinimumCreateFieldSets) > 0 {
		if !createMinimumSatisfied(schema.MinimumCreateFieldSets, request.Values) {
			return CreateRequest{}, invalidMutationPayload("payload", "at_least_one_value_required")
		}
	} else if len(request.Values) == 0 && len(request.AliasAdds) == 0 && !schema.PermitsZeroFieldCreate {
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
	aliasKeys := make([]string, 0, len(request.AliasAdds))
	for key := range request.AliasAdds {
		aliasKeys = append(aliasKeys, key)
	}
	slices.Sort(aliasKeys)
	for _, key := range aliasKeys {
		values := make([]string, 0, len(request.AliasAdds[key]))
		for _, action := range request.AliasAdds[key] {
			values = append(values, action.NormalizedText)
		}
		slices.Sort(values)
		payload[key] = values
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
			"host.display_name":  map[string]any{"value": record.DisplayName},
			"host.aad_device_id": map[string]any{"value": derefString(record.AADDeviceID)},
			"host.fqdn":          map[string]any{"value": derefString(record.FQDN)},
			"host.hostname":      map[string]any{"value": derefString(record.Hostname)},
			"host.aliases":       map[string]any{"value": collectionValue(false, aliasCollectionItems(record.SuggestionOnlyAliases))},
			"host.reusable_identifiers": map[string]any{
				"value": collectionValue(false, hostReusableIdentifierCollectionItems(record)),
			},
			"host.host_state":         map[string]any{"value": record.HostState},
			"host.linked_event_count": map[string]any{"value": record.LinkedEventCount},
			"host.evidence_count":     map[string]any{"value": record.EvidenceCount},
			"host.location":           map[string]any{"value": derefString(record.Location)},
			"host.os_platform":        map[string]any{"value": derefString(record.OSPlatform)},
			"host.business_owner":     map[string]any{"value": derefString(record.BusinessOwner)},
			"host.criticality":        map[string]any{"value": derefString(record.Criticality)},
			"host.containment_status": map[string]any{"value": derefString(record.ContainmentStatus)},
			"host.edited_at":          map[string]any{"value": formatTimestamp(record.UpdatedAt)},
		},
	}
	row["group_values"] = map[string]any{
		"host.host_state":         record.HostState,
		"host.criticality":        derefString(record.Criticality),
		"host.containment_status": derefString(record.ContainmentStatus),
	}
	return row
}

func BuildIdentityRow(record IdentityRecord) map[string]any {
	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells": map[string]any{
			"identity.display_name":     map[string]any{"value": record.DisplayName},
			"identity.aad_object_id":    map[string]any{"value": derefString(record.AADObjectID)},
			"identity.sid":              map[string]any{"value": derefString(record.SID)},
			"identity.upn":              map[string]any{"value": derefString(record.UPN)},
			"identity.email":            map[string]any{"value": derefString(record.Email)},
			"identity.sam_account_name": map[string]any{"value": derefString(record.SamAccountName)},
			"identity.aliases":          map[string]any{"value": collectionValue(false, aliasCollectionItems(record.SuggestionOnlyAliases))},
			"identity.reusable_identifiers": map[string]any{
				"value": collectionValue(false, identityReusableIdentifierCollectionItems(record)),
			},
			"identity.identity_state":     map[string]any{"value": record.IdentityState},
			"identity.linked_event_count": map[string]any{"value": record.LinkedEventCount},
			"identity.evidence_count":     map[string]any{"value": record.EvidenceCount},
			"identity.privilege_level":    map[string]any{"value": derefString(record.PrivilegeLevel)},
			"identity.mfa_state":          map[string]any{"value": derefString(record.MFAState)},
			"identity.reset_status":       map[string]any{"value": derefString(record.ResetStatus)},
			"identity.edited_at":          map[string]any{"value": formatTimestamp(record.UpdatedAt)},
		},
	}
	row["group_values"] = map[string]any{
		"identity.identity_state":  record.IdentityState,
		"identity.privilege_level": derefString(record.PrivilegeLevel),
		"identity.mfa_state":       derefString(record.MFAState),
		"identity.reset_status":    derefString(record.ResetStatus),
	}
	return row
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

func exactMatchConflictError(entityType string, identifierClass string, candidateRecordIDs []uuid.UUID) *auth.APIError {
	details := map[string]any{
		"reason_code":      "merge_required",
		"entity_type":      entityType,
		"identifier_class": identifierClass,
	}
	if len(candidateRecordIDs) > 0 {
		ids := make([]string, 0, len(candidateRecordIDs))
		for _, recordID := range candidateRecordIDs {
			ids = append(ids, recordID.String())
		}
		details["candidate_record_ids"] = ids
	}
	return &auth.APIError{
		Status:  http.StatusConflict,
		Code:    "entity_match_conflict",
		Message: "entity match conflict",
		Details: details,
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

func objectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[field] = struct{}{}
		if _, ok := object[field]; !ok {
			return false
		}
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}

func decodeAliasActionPayload(fieldKey string, value json.RawMessage) ([]CollectionAction, bool) {
	if fieldKey != "host.aliases" && fieldKey != "identity.aliases" {
		return nil, false
	}

	var payload struct {
		Kind    string                       `json:"kind"`
		Actions []map[string]json.RawMessage `json:"actions"`
	}
	if err := json.Unmarshal(value, &payload); err != nil {
		return nil, false
	}
	if payload.Kind != "collection_actions_v1" || len(payload.Actions) == 0 || len(payload.Actions) > maxCollectionActions {
		return nil, false
	}

	actions := make([]CollectionAction, 0, len(payload.Actions))
	for _, rawAction := range payload.Actions {
		var op string
		if err := json.Unmarshal(rawAction["op"], &op); err != nil {
			return nil, false
		}
		switch op {
		case "add_alias":
			if len(rawAction) != 2 {
				return nil, false
			}
			var rawText string
			if err := json.Unmarshal(rawAction["alias_text"], &rawText); err != nil {
				return nil, false
			}
			normalized, ok := fieldnorm.NormalizeLine(rawText)
			if !ok {
				return nil, false
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        rawText,
				NormalizedText: normalized,
			})
		default:
			return nil, false
		}
	}
	return actions, true
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

func aliasCollectionItems(values []string) []map[string]any {
	if len(values) == 0 {
		return nil
	}
	items := make([]map[string]any, 0, len(values))
	for _, value := range values {
		items = append(items, map[string]any{
			"item_kind":    "suggestion_only_alias",
			"display_text": value,
			"raw_text":     value,
		})
	}
	return items
}

func hostReusableIdentifierCollectionItems(record HostRecord) []map[string]any {
	return reusableIdentifierCollectionItems(record.ReusableIdentifiers, func(identifierClass string) string {
		return hostCanonicalNormalized(record, identifierClass)
	})
}

func identityReusableIdentifierCollectionItems(record IdentityRecord) []map[string]any {
	return reusableIdentifierCollectionItems(record.ReusableIdentifiers, func(identifierClass string) string {
		return identityCanonicalNormalized(record, identifierClass)
	})
}

func reusableIdentifierCollectionItems(values []ReusableIdentifier, canonicalNormalized func(string) string) []map[string]any {
	if len(values) == 0 {
		return nil
	}
	orderedValues := append([]ReusableIdentifier(nil), values...)
	slices.SortFunc(orderedValues, func(left ReusableIdentifier, right ReusableIdentifier) int {
		if cmp := strings.Compare(left.IdentifierClass, right.IdentifierClass); cmp != 0 {
			return cmp
		}
		if cmp := strings.Compare(left.NormalizedValue, right.NormalizedValue); cmp != 0 {
			return cmp
		}
		if cmp := strings.Compare(left.RawValue, right.RawValue); cmp != 0 {
			return cmp
		}
		return strings.Compare(left.EntityPreservedIdentifierID.String(), right.EntityPreservedIdentifierID.String())
	})

	items := make([]map[string]any, 0, len(orderedValues))
	for _, value := range orderedValues {
		if canonicalNormalized(value.IdentifierClass) == value.NormalizedValue {
			continue
		}
		items = append(items, map[string]any{
			"item_ref":         "entity_preserved_identifier:" + value.EntityPreservedIdentifierID.String(),
			"item_kind":        "reusable_identifier",
			"identifier_class": value.IdentifierClass,
			"raw_value":        value.RawValue,
			"normalized_value": value.NormalizedValue,
			"display_text":     reusableIdentifierDisplayText(value.IdentifierClass, value.RawValue),
		})
	}
	return items
}

func reusableIdentifierDisplayText(identifierClass string, rawValue string) string {
	label := identifierClass
	switch identifierClass {
	case "aad_device_id":
		label = "AAD Device ID"
	case "fqdn":
		label = "FQDN"
	case "hostname":
		label = "Hostname"
	case "aad_object_id":
		label = "AAD Object ID"
	case "sid":
		label = "SID"
	case "upn":
		label = "UPN"
	case "email":
		label = "Email"
	case "sam_account_name":
		label = "SAM Account Name"
	}
	return label + ": " + rawValue
}
