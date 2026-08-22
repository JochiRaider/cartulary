package hostidentity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	HostsViewSchemaID      = entitycontract.HostsViewSchemaID
	IdentitiesViewSchemaID = entitycontract.IdentitiesViewSchemaID

	hostCreateRouteKey     = "entities.hosts.rows.create"
	identityCreateRouteKey = "entities.identities.rows.create"
	maxCollectionActions   = 64
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
	SuggestionOnlyAliases []AliasValue
	AliasMutations        []AliasMutationValue
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
	SuggestionOnlyAliases []AliasValue
	AliasMutations        []AliasMutationValue
	ReusableIdentifiers   []ReusableIdentifier
	RowVersion            int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
	CreatedByUser         uuid.UUID
	UpdatedByUser         uuid.UUID
}

type ReusableIdentifier struct {
	EntityPreservedIdentifierID uuid.UUID
	IdentifierClass             string
	RawValue                    string
	NormalizedValue             string
}

type AliasValue struct {
	EntityAliasID uuid.UUID
	AliasText     string
}

type AliasMutationValue struct {
	EntityAliasID uuid.UUID
	IncidentID    uuid.UUID
	RecordID      uuid.UUID
	EntityType    string
	AliasText     string
	CreatedByUser uuid.UUID
	CreatedAt     time.Time
	DeletedAt     *time.Time
}

func (value AliasMutationValue) MutationValue() map[string]any {
	var deletedAt any
	if value.DeletedAt != nil {
		deletedAt = value.DeletedAt.UTC()
	}
	return map[string]any{
		"entity_alias_id":    value.EntityAliasID.String(),
		"incident_id":        value.IncidentID.String(),
		"record_id":          value.RecordID.String(),
		"entity_type":        value.EntityType,
		"raw_text":           value.AliasText,
		"normalized_text":    value.AliasText,
		"classification":     "suggestion_only",
		"created_by_user_id": value.CreatedByUser.String(),
		"created_at":         value.CreatedAt.UTC(),
		"deleted_at":         deletedAt,
	}
}

type AliasSyncResult struct {
	Added              []AliasMutationValue
	DuplicateNoopCount int
}

func (result AliasSyncResult) Changed() bool { return len(result.Added) > 0 }

type AliasAppliedMutation struct {
	OperationKind string
	TargetID      string
	BeforeValue   map[string]any
	AfterValue    map[string]any
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

func ParseEntityAliasItemRef(itemRef string) (uuid.UUID, error) {
	const prefix = "entity_alias:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.Nil, fmt.Errorf("invalid entity alias item ref")
	}
	rawID := strings.TrimPrefix(itemRef, prefix)
	aliasID, err := uuid.Parse(rawID)
	if err != nil || aliasID.String() != rawID {
		return uuid.Nil, fmt.Errorf("invalid entity alias item ref")
	}
	return aliasID, nil
}

func DecodeCreateRequest(viewSchemaID string, reader io.Reader) (CreateRequest, *httpapi.APIError) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for _, descriptor := range entityFields.descriptors(viewSchemaID) {
		if descriptor.participatesInCreate() {
			allowed[descriptor.fieldKey] = struct{}{}
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
	for _, descriptor := range entityFields.descriptors(viewSchemaID) {
		fieldKey := descriptor.fieldKey
		field := descriptor.owner
		value, ok := raw[fieldKey]
		if !ok {
			continue
		}
		if !field.Writable && !field.CreateWritable {
			return CreateRequest{}, invalidMutationPayload(fieldKey, "readonly_field")
		}
		if descriptor.patch == entityFieldPatchCollection {
			actions, ok := decodeAliasActionPayload(viewSchemaID, fieldKey, value)
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
	return entityFields.buildHostRow(record)
}

func BuildIdentityRow(record IdentityRecord) map[string]any {
	return entityFields.buildIdentityRow(record)
}

func BuildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": viewSchemaID,
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

func decodeAliasActionPayload(viewSchemaID string, fieldKey string, value json.RawMessage) ([]CollectionAction, bool) {
	descriptor, ok := entityFields.lookup(viewSchemaID, fieldKey)
	if !ok || descriptor.patch != entityFieldPatchCollection {
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
			normalized, ok := fieldnorm.NormalizeAliasText(rawText)
			if !ok {
				return nil, false
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        normalized,
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

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func aliasCollectionItems(values []AliasValue) []map[string]any {
	if len(values) == 0 {
		return nil
	}
	items := make([]map[string]any, 0, len(values))
	for _, value := range values {
		items = append(items, map[string]any{
			"item_ref":     "entity_alias:" + value.EntityAliasID.String(),
			"item_kind":    "alias",
			"display_text": value.AliasText,
			"alias_text":   value.AliasText,
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
