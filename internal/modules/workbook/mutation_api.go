package workbook

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	workbookCreateRouteKey          = "workbook.rows.create"
	workbookPatchRouteKey           = "workbook.records.patch"
	workbookSupersedeRouteKey       = "workbook.records.supersede"
	workbookConflictResolveRouteKey = "workbook.records.conflicts.resolve"
	workbookLinkedNoteRouteKey      = "workbook.records.linked_notes.create"
	maxPatchChanges                 = 32
	maxCollectionActions            = 64
)

type MutationResult struct {
	Payload                 map[string]any
	StatusCode              int
	Replayed                bool
	IncidentID              uuid.UUID
	RecordID                uuid.UUID
	ChangeSetID             uuid.UUID
	ClientTxnID             string
	RowVersion              int64
	ViewSchemaID            string
	ChangedFieldKeys        []string
	AdditionalRecordChanges []MutationResult
}

type CreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]ValueChange
	Collections  map[string]CollectionActionPayload
}

type LinkedNoteCreateRequest struct {
	ClientTxnID string
	Values      map[string]ValueChange
	Collections map[string]CollectionActionPayload
}

type PatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []PatchChange
}

type ConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	ResolvedChange *PatchChange
	CanonicalAny   any
}

type PatchChange struct {
	FieldKey     string
	Value        *ValueChange
	Collection   *CollectionActionPayload
	CanonicalAny any
}

type ValueChange struct {
	Kind      string
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	RiskRefText    string
	NormalizedText string
}

type MutationValidationError struct {
	Field      string
	ReasonCode string
}

func (e *MutationValidationError) Error() string {
	return "workbook: invalid mutation request"
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "workbook: row version conflict"
}

func (e *RowVersionConflictError) Details() map[string]any {
	return map[string]any{
		"record_id":           e.RecordID.String(),
		"base_row_version":    e.BaseRowVersion,
		"current_row_version": e.CurrentRowVersion,
	}
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "workbook: same field conflict"
}

type LifecycleValidationError struct {
	FromStatus     string
	ToStatus       string
	ViolatedGuards []string
	ReasonCode     string
}

func (e *LifecycleValidationError) Error() string {
	return "workbook: illegal transition"
}

func DecodeCreateRequest(viewSchemaID string, reader io.Reader) (CreateRequest, *auth.APIError) {
	if !isWorkbookMutationSurface(viewSchemaID) {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}
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
	request := CreateRequest{
		ViewSchemaID: viewSchemaID,
		Values:       map[string]ValueChange{},
		Collections:  map[string]CollectionActionPayload{},
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	for fieldKey, field := range schema.Fields() {
		value, ok := raw[fieldKey]
		if !ok {
			continue
		}
		if field.ConflictResolutionClass == "collection_review" {
			payload, apiErr := decodeCollectionActionPayload(fieldKey, value)
			if apiErr != nil {
				return CreateRequest{}, apiErr
			}
			request.Collections[fieldKey] = payload
			continue
		}
		change, canonical, apiErr := decodeDirectValue(fieldKey, field, value, false)
		if apiErr != nil {
			return CreateRequest{}, apiErr
		}
		request.Values[fieldKey] = change
		_ = canonical
	}
	return request, nil
}

func DecodeLinkedNoteCreateRequest(reader io.Reader) (LinkedNoteCreateRequest, *auth.APIError) {
	create, apiErr := DecodeCreateRequest(NotesViewSchemaID, reader)
	if apiErr != nil {
		return LinkedNoteCreateRequest{}, apiErr
	}
	return LinkedNoteCreateRequest{
		ClientTxnID: create.ClientTxnID,
		Values:      create.Values,
		Collections: create.Collections,
	}, nil
}

func DecodePatchRequest(reader io.Reader) (PatchRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return PatchRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"view_schema_id":   {},
		"base_row_version": {},
		"client_txn_id":    {},
		"changes":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request PatchRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || !isWorkbookMutationSurface(request.ViewSchemaID) {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, ok := raw["base_row_version"]; !ok {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	value, ok := raw["changes"]
	if !ok {
		return PatchRequest{}, invalidMutationPayload("changes", "missing_required_field")
	}
	var rawChanges []json.RawMessage
	if err := json.Unmarshal(value, &rawChanges); err != nil {
		return PatchRequest{}, invalidMutationPayload("changes", "invalid_value")
	}
	if len(rawChanges) == 0 {
		return PatchRequest{}, invalidMutationPayload("changes", "empty_changes")
	}
	if len(rawChanges) > maxPatchChanges {
		return PatchRequest{}, invalidMutationPayloadWithDetails("changes", "change_count_exceeded", map[string]any{
			"requested_count": len(rawChanges),
			"max_count":       maxPatchChanges,
		})
	}
	seen := map[string]struct{}{}
	for _, rawChange := range rawChanges {
		change, apiErr := decodePatchChange(request.ViewSchemaID, rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, ok := seen[change.FieldKey]; ok {
			return PatchRequest{}, invalidMutationPayload("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.Changes = append(request.Changes, change)
	}
	slices.SortFunc(request.Changes, func(left PatchChange, right PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func DecodeConflictResolveRequest(reader io.Reader, token string, claims workbookConflictTokenClaims) (ConflictResolveRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return ConflictResolveRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"conflict_token":  {},
		"resolution_kind": {},
		"client_txn_id":   {},
		"resolved_value":  {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ConflictResolveRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	request := ConflictResolveRequest{ConflictToken: token}
	if value, ok := raw["conflict_token"]; !ok {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ConflictToken); err != nil || request.ConflictToken != token {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, ok := raw["resolution_kind"]; !ok {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ResolutionKind); err != nil {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	resolvedValue, hasResolvedValue := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if hasResolvedValue {
			return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !hasResolvedValue {
		return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "missing_required_field")
	}
	field, ok := viewschema.LookupField(claims.ViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable || isReadOnlySystemField(claims.FieldKey) {
		return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeCollectionActionPayload(claims.FieldKey, resolvedValue)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Collection = &payload
		change.CanonicalAny = canonicalCollectionActionPayload(payload)
	} else {
		value, canonical, apiErr := decodeDirectValue(claims.FieldKey, field, resolvedValue, true)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Value = &value
		change.CanonicalAny = canonical
	}
	request.ResolvedChange = &change
	request.CanonicalAny = change.CanonicalAny
	return request, nil
}

func decodePatchChange(viewSchemaID string, raw json.RawMessage) (PatchChange, *auth.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return PatchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}
	fieldValue, ok := object["field_key"]
	if !ok {
		return PatchChange{}, invalidMutationPayload("changes", "missing_field_key")
	}
	var fieldKey string
	if err := json.Unmarshal(fieldValue, &fieldKey); err != nil {
		return PatchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	if isReadOnlySystemField(fieldKey) {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	change := PatchChange{FieldKey: fieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return PatchChange{}, invalidMutationPayload("action_payload", "missing_required_field")
		}
		payload, apiErr := decodeCollectionActionPayload(fieldKey, actionPayload)
		if apiErr != nil {
			return PatchChange{}, apiErr
		}
		change.Collection = &payload
		change.CanonicalAny = canonicalCollectionActionPayload(payload)
		return change, nil
	}
	if !hasValue {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	direct, canonical, apiErr := decodeDirectValue(fieldKey, field, value, true)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	change.Value = &direct
	change.CanonicalAny = canonical
	return change, nil
}

func decodeDirectValue(fieldKey string, field viewschema.Field, value json.RawMessage, patch bool) (ValueChange, any, *auth.APIError) {
	if string(value) == "null" {
		if field.Clearable {
			return ValueChange{Kind: "null"}, nil, nil
		}
		return ValueChange{}, nil, invalidMutationPayload(fieldKey, "field_not_nullable")
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		var raw string
		if err := json.Unmarshal(value, &raw); err != nil {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		utc, ok := fieldnorm.NormalizeTimestampInstant(raw)
		if !ok {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return ValueChange{Kind: "timestamp", Timestamp: &utc}, utc.Format(time.RFC3339Nano), nil
	}
	if field.DirectReferenceContractID != nil {
		var raw string
		if err := json.Unmarshal(value, &raw); err != nil {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		parsed, err := uuid.Parse(raw)
		if err != nil || parsed.String() != raw {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return ValueChange{Kind: "uuid", UUID: &parsed}, parsed.String(), nil
	}
	if isUUIDField(fieldKey, field) {
		var raw string
		if err := json.Unmarshal(value, &raw); err != nil {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		parsed, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return ValueChange{Kind: "uuid", UUID: &parsed}, parsed.String(), nil
	}
	if field.ReadKind == "number" {
		parsed, ok := decodeIntegerValue(value)
		if !ok {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return ValueChange{Kind: "number", Number: &parsed}, parsed, nil
	}
	if field.ReadKind == "boolean" {
		parsed, ok := decodeBooleanValue(value)
		if !ok {
			return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return ValueChange{Kind: "bool", Bool: &parsed}, parsed, nil
	}
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	normalized, ok := normalizeStringContract(field, raw)
	if !ok {
		return ValueChange{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	if fieldKey == "evidence.storage_ref" && blobref.IsServerManagedStorageRef(normalized) {
		return ValueChange{}, nil, invalidMutationPayload(fieldKey, "reserved_server_managed_ref")
	}
	return ValueChange{Kind: "text", Text: &normalized}, normalized, nil
}

func decodeIntegerValue(value json.RawMessage) (int64, bool) {
	var rawNumber int64
	if err := json.Unmarshal(value, &rawNumber); err == nil {
		return rawNumber, true
	}
	var rawText string
	if err := json.Unmarshal(value, &rawText); err != nil {
		return 0, false
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(rawText), 10, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func decodeBooleanValue(value json.RawMessage) (bool, bool) {
	var rawBool bool
	if err := json.Unmarshal(value, &rawBool); err == nil {
		return rawBool, true
	}
	var rawText string
	if err := json.Unmarshal(value, &rawText); err != nil {
		return false, false
	}
	switch strings.TrimSpace(rawText) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

func decodeCollectionActionPayload(fieldKey string, raw json.RawMessage) (CollectionActionPayload, *auth.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	if !objectHasOnlyFields(object, "kind", "actions") {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var kind string
	if err := json.Unmarshal(object["kind"], &kind); err != nil || kind != "collection_actions_v1" {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var rawActions []json.RawMessage
	if err := json.Unmarshal(object["actions"], &rawActions); err != nil {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	if len(rawActions) == 0 {
		return CollectionActionPayload{}, invalidMutationPayloadWithDetails(fieldKey+".actions", "empty_collection_actions", map[string]any{"field_key": fieldKey})
	}
	if len(rawActions) > maxCollectionActions {
		return CollectionActionPayload{}, invalidMutationPayloadWithDetails(fieldKey+".actions", "collection_action_count_exceeded", map[string]any{
			"field_key":       fieldKey,
			"requested_count": len(rawActions),
			"max_count":       maxCollectionActions,
		})
	}
	payload := CollectionActionPayload{Actions: make([]CollectionAction, 0, len(rawActions))}
	for _, rawActionData := range rawActions {
		action, apiErr := decodeCollectionAction(fieldKey, rawActionData)
		if apiErr != nil {
			return CollectionActionPayload{}, apiErr
		}
		payload.Actions = append(payload.Actions, action)
	}
	return payload, nil
}

func decodeCollectionAction(fieldKey string, raw json.RawMessage) (CollectionAction, *auth.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var op string
	if err := json.Unmarshal(object["op"], &op); err != nil {
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	action := CollectionAction{Op: op}
	switch op {
	case "add_token":
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	case "add_tag":
		if !isTagCollection(fieldKey) || !objectHasOnlyFields(object, "op", "tag_name") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		rawText, ok := decodeStringActionField(object, "tag_name")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		label, normalized, ok := fieldnorm.NormalizeTagLabel(rawText)
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.RawText = label
		action.NormalizedText = normalized
	case "remove_tag":
		if !isTagCollection(fieldKey) || !objectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := decodeStringActionField(object, "item_ref")
		if !ok || !isExactRecordTagItemRef(itemRef) {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	case "add_record_ref":
		if !isRecordRefCollection(fieldKey) || !objectHasOnlyFields(object, "op", "linked_record_id") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		parsed, ok := decodeUUIDActionField(object, "linked_record_id")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.LinkedRecordID = &parsed
	case "remove_record_ref":
		if !isRecordRefCollection(fieldKey) || !objectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := decodeStringActionField(object, "item_ref")
		if !ok || !isExactUUIDItemRef(itemRef, "record_ref:") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	case "add_party_ref":
		if !isPartyRefCollection(fieldKey) || !objectHasOnlyFields(object, "op", "party_id") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		parsed, ok := decodeUUIDActionField(object, "party_id")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.PartyID = &parsed
	case "remove_party_ref":
		if !isPartyRefCollection(fieldKey) || !objectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := decodeStringActionField(object, "item_ref")
		if !ok || !isExactUUIDItemRef(itemRef, "party_ref:") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	case "add_risk_ref":
		if fieldKey != "handoff.open_risk_refs" || !objectHasOnlyFields(object, "op", "risk_ref_text") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		rawText, ok := decodeStringActionField(object, "risk_ref_text")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		normalized, ok := fieldnorm.NormalizeLine(rawText)
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.RiskRefText = normalized
		action.NormalizedText = normalized
	case "remove_risk_ref":
		if fieldKey != "handoff.open_risk_refs" || !objectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := decodeStringActionField(object, "item_ref")
		if !ok || !isExactUUIDItemRef(itemRef, "risk_ref:") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	default:
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	return action, nil
}

func CreateRequestHash(request CreateRequest) []byte {
	payload := map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"values":         canonicalValues(request.Values),
		"collection_ops": canonicalCollections(request.Collections),
	}
	return hashRequestPayload(payload)
}

func PatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{
			"field_key": change.FieldKey,
			"value":     change.CanonicalAny,
		})
	}
	return hashRequestPayload(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"changes":          changes,
	})
}

func ConflictResolveRequestHash(claims workbookConflictTokenClaims, request ConflictResolveRequest) []byte {
	return hashRequestPayload(map[string]any{
		"conflict_token":      request.ConflictToken,
		"resolution_kind":     request.ResolutionKind,
		"client_txn_id":       request.ClientTxnID,
		"record_id":           claims.RecordID,
		"view_schema_id":      claims.ViewSchemaID,
		"field_key":           claims.FieldKey,
		"current_row_version": claims.CurrentRowVersion,
		"resolved_value":      request.CanonicalAny,
	})
}

func LinkedNoteCreateRequestHash(sourceRecordID uuid.UUID, request LinkedNoteCreateRequest) []byte {
	payload := map[string]any{
		"source_record_id": sourceRecordID.String(),
		"view_schema_id":   NotesViewSchemaID,
		"values":           canonicalValues(request.Values),
		"collection_ops":   canonicalCollections(request.Collections),
	}
	return hashRequestPayload(payload)
}

func BuildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": viewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	for key, value := range extra {
		details[key] = value
	}
	return &auth.APIError{Status: http.StatusBadRequest, Code: "invalid_mutation_payload", Message: "invalid mutation payload", Details: details}
}

func rowVersionConflictError(details map[string]any) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: details}
}

func sameFieldConflictError(err *SameFieldConflictError) *auth.APIError {
	conflict := any(nil)
	if err != nil {
		conflict = err.Conflict
	}
	return &auth.APIError{Status: http.StatusConflict, Code: "same_field_conflict", Message: "same field conflict", Details: map[string]any{}, Conflict: conflict}
}

func mutationValidationError(field string, reasonCode string) error {
	return &MutationValidationError{Field: field, ReasonCode: reasonCode}
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func normalizeStringContract(field viewschema.Field, raw string) (string, bool) {
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		return fieldnorm.NormalizeNote(raw)
	}
	return fieldnorm.NormalizeLine(raw)
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

func decodeUUIDActionField(object map[string]json.RawMessage, field string) (uuid.UUID, bool) {
	text, ok := decodeStringActionField(object, field)
	if !ok {
		return uuid.UUID{}, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func decodeStringActionField(object map[string]json.RawMessage, field string) (string, bool) {
	value, ok := object[field]
	if !ok {
		return "", false
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil || strings.TrimSpace(text) == "" {
		return "", false
	}
	return text, true
}

func isExactUUIDItemRef(itemRef string, prefix string) bool {
	if !strings.HasPrefix(itemRef, prefix) {
		return false
	}
	suffix := strings.TrimPrefix(itemRef, prefix)
	parsed, err := uuid.Parse(suffix)
	return err == nil && parsed.String() == suffix
}

func isExactRecordTagItemRef(itemRef string) bool {
	parts := strings.Split(itemRef, ":")
	if len(parts) != 3 || parts[0] != "record_tag" {
		return false
	}
	recordID, err := uuid.Parse(parts[1])
	if err != nil || recordID.String() != parts[1] {
		return false
	}
	tagID, err := uuid.Parse(parts[2])
	return err == nil && tagID.String() == parts[2]
}

func canonicalValues(values map[string]ValueChange) map[string]any {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		value := canonicalValue(values[key])
		if value == nil {
			continue
		}
		result[key] = value
	}
	return result
}

func canonicalValue(value ValueChange) any {
	switch value.Kind {
	case "timestamp":
		if value.Timestamp == nil {
			return nil
		}
		return value.Timestamp.UTC().Format(time.RFC3339Nano)
	case "uuid":
		if value.UUID == nil {
			return nil
		}
		return value.UUID.String()
	case "text":
		if value.Text == nil {
			return nil
		}
		return *value.Text
	case "number":
		if value.Number == nil {
			return nil
		}
		return *value.Number
	case "bool":
		if value.Bool == nil {
			return nil
		}
		return *value.Bool
	default:
		return nil
	}
}

func canonicalCollections(collections map[string]CollectionActionPayload) map[string]any {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		result[key] = canonicalCollectionActionPayload(collections[key])
	}
	return result
}

func canonicalCollectionActionPayload(payload CollectionActionPayload) map[string]any {
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.LinkedRecordID != nil {
			entry["linked_record_id"] = action.LinkedRecordID.String()
		}
		if action.PartyID != nil {
			entry["party_id"] = action.PartyID.String()
		}
		if action.ItemRef != "" {
			entry["item_ref"] = action.ItemRef
		}
		if action.Op == "add_tag" && action.RawText != "" {
			entry["tag_name"] = action.RawText
		}
		if action.Op == "add_token" && action.NormalizedText != "" {
			entry["raw_text"] = action.NormalizedText
		}
		if action.Op == "add_risk_ref" && action.NormalizedText != "" {
			entry["risk_ref_text"] = action.NormalizedText
		}
		actions = append(actions, entry)
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func isWorkbookMutationSurface(viewSchemaID string) bool {
	switch viewSchemaID {
	case EvidenceViewSchemaID, PartiesViewSchemaID, NotesViewSchemaID, TaskRequestsViewSchemaID, DecisionsViewSchemaID,
		CommLogViewSchemaID, HandoffViewSchemaID, StatusReviewViewSchemaID, LessonViewSchemaID,
		FindingsViewSchemaID, InvestigativeQueriesViewSchemaID, ForensicKeywordsViewSchemaID:
		return true
	default:
		return false
	}
}

func isReadOnlySystemField(fieldKey string) bool {
	switch fieldKey {
	case "comm_log.comm_id", "handoff.handoff_id", "status_review.status_review_id", "lesson.lesson_id",
		"comm_log.timestamp_day", "comm_log.next_report_day", "comm_log.updated_at",
		"handoff.timestamp_day", "handoff.ack_state", "handoff.updated_at",
		"status_review.timestamp_day", "status_review.next_report_day", "status_review.updated_at",
		"lesson.timestamp_day", "lesson.updated_at", "party.updated_at",
		"evidence.blob_hash", "evidence.upload_state", "evidence.linked_record_count", "evidence.edited_at",
		"note.linked_record_count", "note.updated_at", "note.created_by_user_id",
		"task.linked_record_count", "task.updated_at", "task.no_owner",
		"decision.affected_record_count", "decision.supersedes_record_id", "decision.updated_at", "decision.is_superseded",
		"finding.closed_at", "finding.confidence_band", "finding.updated_at",
		"investigative_query.query_id", "investigative_query.created_by_user_id", "investigative_query.created_at", "investigative_query.created_day",
		"forensic_keyword.keyword_id", "forensic_keyword.created_at", "forensic_keyword.created_day":
		return true
	default:
		return false
	}
}

func isUUIDField(fieldKey string, field viewschema.Field) bool {
	return strings.HasSuffix(fieldKey, "_user_id") || field.DirectReferenceContractID != nil
}

func isRecordRefCollection(fieldKey string) bool {
	switch fieldKey {
	case "comm_log.decision_ids", "comm_log.action_task_ids",
		"handoff.open_task_ids", "handoff.open_decision_ids",
		"status_review.blocked_task_ids", "status_review.pending_evidence_ids", "status_review.open_decision_ids",
		"lesson.follow_up_task_ids", "lesson.evidence_refs",
		"task.linked_record_ids", "decision.support_refs", "decision.affected_record_ids",
		"finding.supporting_refs", "finding.contradictory_refs":
		return true
	default:
		return false
	}
}

func isTagCollection(fieldKey string) bool {
	return fieldKey == "note.tags"
}

func isPartyRefCollection(fieldKey string) bool {
	return fieldKey == "comm_log.audience_party_ids" || fieldKey == "comm_log.attendee_party_ids"
}
