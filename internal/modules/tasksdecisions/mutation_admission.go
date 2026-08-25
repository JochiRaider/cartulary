package tasksdecisions

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	maxMutationPatchChanges      = 32
	maxMutationCollectionActions = 64
)

type ConflictClaims struct {
	RecordID          uuid.UUID
	ViewSchemaID      string
	FieldKey          string
	CurrentRowVersion int64
}

type ConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	Patch          *PatchRequest
	CanonicalValue any
}

func AdmitCreateJSON(viewSchemaID string, reader io.Reader) (CreateRequest, *AdmissionFailure) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok || !schema.CreateCapable || !isMutationView(viewSchemaID) {
		return CreateRequest{}, invalidAdmission("view_schema_id", "unknown_view_schema")
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return CreateRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	request := CreateRequest{
		ViewSchemaID: viewSchemaID,
		Values:       map[string]FieldValue{},
		Collections:  map[string]CollectionActionPayload{},
	}
	if value, present := raw["client_txn_id"]; !present {
		return CreateRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		if field.ConflictResolutionClass == "collection_review" {
			payload, apiErr := decodeMutationCollectionPayload(fieldKey, value)
			if apiErr != nil {
				return CreateRequest{}, apiErr
			}
			request.Collections[fieldKey] = payload
			continue
		}
		admitted, _, apiErr := decodeMutationValue(fieldKey, field, value, false)
		if apiErr != nil {
			return CreateRequest{}, apiErr
		}
		request.Values[fieldKey] = admitted
	}
	return request, nil
}

func AdmitPatchJSON(reader io.Reader) (PatchRequest, *AdmissionFailure) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return PatchRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	var request PatchRequest
	if value, present := raw["view_schema_id"]; !present {
		return PatchRequest{}, invalidAdmission("view_schema_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ViewSchemaID) != nil || !isMutationView(request.ViewSchemaID) {
		return PatchRequest{}, invalidAdmission("view_schema_id", "invalid_view_schema_id")
	}
	if value, present := raw["base_row_version"]; !present {
		return PatchRequest{}, invalidAdmission("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidAdmission("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	var rawChanges []json.RawMessage
	if value, present := raw["changes"]; !present {
		return PatchRequest{}, invalidAdmission("changes", "missing_required_field")
	} else if json.Unmarshal(value, &rawChanges) != nil {
		return PatchRequest{}, invalidAdmission("changes", "invalid_value")
	}
	if len(rawChanges) == 0 {
		return PatchRequest{}, invalidAdmission("changes", "empty_changes")
	}
	if len(rawChanges) > maxMutationPatchChanges {
		return PatchRequest{}, invalidCountAdmission(
			"changes", "change_count_exceeded", len(rawChanges), maxMutationPatchChanges, "",
		)
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, apiErr := decodeMutationPatchChange(request.ViewSchemaID, rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, duplicate := seen[change.FieldKey]; duplicate {
			return PatchRequest{}, invalidAdmission("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.Changes = append(request.Changes, change)
	}
	slices.SortFunc(request.Changes, func(left, right PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func AdmitConflictResolveJSON(reader io.Reader, token string, claims ConflictClaims) (ConflictResolveRequest, *AdmissionFailure) {
	if claims.RecordID == uuid.Nil || !isMutationView(claims.ViewSchemaID) || claims.CurrentRowVersion < 1 {
		return ConflictResolveRequest{}, invalidAdmission("conflict_token", "invalid_value")
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return ConflictResolveRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	request := ConflictResolveRequest{ConflictToken: token}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveRequest{}, invalidAdmission("conflict_token", "missing_required_field")
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return ConflictResolveRequest{}, invalidAdmission("conflict_token", "invalid_value")
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveRequest{}, invalidAdmission("resolution_kind", "missing_required_field")
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return ConflictResolveRequest{}, invalidAdmission("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveRequest{}, invalidAdmission("resolution_kind", "invalid_value")
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	resolved, present := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if present {
			return ConflictResolveRequest{}, invalidAdmission("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !present {
		return ConflictResolveRequest{}, invalidAdmission("resolved_value", "missing_required_field")
	}
	field, ok := viewschema.LookupField(claims.ViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return ConflictResolveRequest{}, invalidAdmission("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeMutationCollectionPayload(claims.FieldKey, resolved)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Collection = &payload
		change.CanonicalValue = canonicalMutationCollection(payload)
	} else {
		value, canonical, apiErr := decodeMutationValue(claims.FieldKey, field, resolved, true)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Value, change.CanonicalValue = &value, canonical
	}
	request.Patch = &PatchRequest{
		ViewSchemaID: claims.ViewSchemaID, BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID: request.ClientTxnID, Changes: []PatchChange{change},
	}
	request.CanonicalValue = change.CanonicalValue
	return request, nil
}

func AdmitSupersedeJSON(reader io.Reader) (SupersedeRequest, *AdmissionFailure) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return SupersedeRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"base_row_version": {}, "client_txn_id": {}, "reason": {}, "replacement_record_id": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return SupersedeRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	var request SupersedeRequest
	if value, present := raw["base_row_version"]; !present {
		return SupersedeRequest{}, invalidAdmission("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return SupersedeRequest{}, invalidAdmission("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return SupersedeRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return SupersedeRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	if value, present := raw["reason"]; !present {
		return SupersedeRequest{}, invalidAdmission("reason", "missing_required_field")
	} else {
		var rawReason string
		if json.Unmarshal(value, &rawReason) != nil {
			return SupersedeRequest{}, invalidAdmission("reason", "invalid_value")
		}
		reason, ok := fieldnorm.NormalizeNote(rawReason)
		if !ok {
			return SupersedeRequest{}, invalidAdmission("reason", "invalid_value")
		}
		request.Reason = reason
	}
	if value, present := raw["replacement_record_id"]; present {
		if string(value) == "null" {
			return SupersedeRequest{}, invalidAdmission("replacement_record_id", "field_not_nullable")
		}
		var text string
		if json.Unmarshal(value, &text) != nil {
			return SupersedeRequest{}, invalidAdmission("replacement_record_id", "invalid_value")
		}
		parsed, parseErr := uuid.Parse(text)
		if parseErr != nil {
			return SupersedeRequest{}, invalidAdmission("replacement_record_id", "invalid_value")
		}
		request.ReplacementRecordID = &parsed
	}
	return request, nil
}

func CreateRequestHash(request CreateRequest) []byte {
	return hashMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"values":         canonicalMutationValues(request.Values),
		"collection_ops": canonicalMutationCollections(request.Collections),
		"create_inputs":  map[string]any{},
	})
}

func PatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": change.CanonicalValue})
	}
	return hashMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID, "base_row_version": request.BaseRowVersion, "changes": changes,
	})
}

func ConflictResolveRequestHash(claims ConflictClaims, request ConflictResolveRequest) []byte {
	return hashMutationPayload(map[string]any{
		"conflict_token": request.ConflictToken, "resolution_kind": request.ResolutionKind,
		"record_id": claims.RecordID, "view_schema_id": claims.ViewSchemaID,
		"field_key": claims.FieldKey, "current_row_version": claims.CurrentRowVersion,
		"resolved_value": request.CanonicalValue,
	})
}

func SupersedeRequestHash(request SupersedeRequest) []byte {
	return hashMutationPayload(map[string]any{
		"base_row_version": request.BaseRowVersion, "reason": request.Reason,
		"replacement_record_id": request.ReplacementRecordID,
	})
}

func decodeMutationPatchChange(viewSchemaID string, raw json.RawMessage) (PatchChange, *AdmissionFailure) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return PatchChange{}, invalidAdmission("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return PatchChange{}, invalidAdmission("changes", "unknown_field")
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return PatchChange{}, invalidAdmission("changes", "missing_field_key")
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return PatchChange{}, invalidAdmission("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidAdmission(fieldKey, "unsupported_field_key")
	}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidAdmission("changes", "invalid_change")
	}
	change := PatchChange{FieldKey: fieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return PatchChange{}, invalidAdmission("action_payload", "missing_required_field")
		}
		payload, apiErr := decodeMutationCollectionPayload(fieldKey, actionPayload)
		if apiErr != nil {
			return PatchChange{}, apiErr
		}
		change.Collection, change.CanonicalValue = &payload, canonicalMutationCollection(payload)
		return change, nil
	}
	if !hasValue {
		return PatchChange{}, invalidAdmission("value", "missing_required_field")
	}
	direct, canonical, apiErr := decodeMutationValue(fieldKey, field, value, true)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	change.Value, change.CanonicalValue = &direct, canonical
	return change, nil
}

func decodeMutationValue(fieldKey string, field viewschema.Field, raw json.RawMessage, patch bool) (FieldValue, any, *AdmissionFailure) {
	if string(raw) == "null" {
		if !patch || field.Clearable {
			return FieldValue{}, nil, nil
		}
		return FieldValue{}, nil, invalidAdmission(fieldKey, "field_not_nullable")
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
		}
		normalized, ok := fieldnorm.NormalizeTimestampInstant(text)
		if !ok {
			return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
		}
		return FieldValue{Timestamp: &normalized}, normalized.Format(time.RFC3339Nano), nil
	}
	if field.DirectReferenceContractID != nil || strings.HasSuffix(fieldKey, "_user_id") {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
		}
		parsed, err := uuid.Parse(strings.TrimSpace(text))
		if err != nil || (field.DirectReferenceContractID != nil && parsed.String() != text) {
			return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
		}
		return FieldValue{UUID: &parsed}, parsed.String(), nil
	}
	if field.ReadKind == "number" {
		value, ok := decodeMutationInteger(raw)
		if !ok {
			return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
		}
		return FieldValue{Number: &value}, value, nil
	}
	if field.ReadKind == "boolean" {
		var value bool
		if json.Unmarshal(raw, &value) != nil {
			return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
		}
		return FieldValue{Bool: &value}, value, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
	}
	var normalized string
	var ok bool
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		normalized, ok = fieldnorm.NormalizeNote(text)
	} else {
		normalized, ok = fieldnorm.NormalizeLine(text)
	}
	if !ok {
		return FieldValue{}, nil, invalidAdmission(fieldKey, "invalid_value")
	}
	return FieldValue{Text: &normalized}, normalized, nil
}

func decodeMutationCollectionPayload(fieldKey string, raw json.RawMessage) (CollectionActionPayload, *AdmissionFailure) {
	if !IsRecordRefCollectionField(fieldKey) {
		return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !mutationObjectHasOnlyFields(object, "kind", "actions") {
		return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
	}
	var kind string
	if json.Unmarshal(object["kind"], &kind) != nil || kind != "collection_actions_v1" {
		return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
	}
	var rawActions []json.RawMessage
	if json.Unmarshal(object["actions"], &rawActions) != nil {
		return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
	}
	if len(rawActions) == 0 {
		return CollectionActionPayload{}, invalidCollectionAdmission(
			fieldKey+".actions", "empty_collection_actions", fieldKey,
		)
	}
	if len(rawActions) > maxMutationCollectionActions {
		return CollectionActionPayload{}, invalidCountAdmission(
			fieldKey+".actions", "collection_action_count_exceeded",
			len(rawActions), maxMutationCollectionActions, fieldKey,
		)
	}
	payload := CollectionActionPayload{Actions: make([]CollectionAction, 0, len(rawActions))}
	for _, rawAction := range rawActions {
		var object map[string]json.RawMessage
		if json.Unmarshal(rawAction, &object) != nil {
			return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
		}
		var op string
		if json.Unmarshal(object["op"], &op) != nil || !AllowsCollectionOp(fieldKey, op) {
			return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
		}
		action := CollectionAction{Op: op}
		switch op {
		case "add_record_ref":
			if !mutationObjectHasOnlyFields(object, "op", "linked_record_id") {
				return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
			}
			var text string
			if json.Unmarshal(object["linked_record_id"], &text) != nil {
				return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
			}
			id, err := uuid.Parse(text)
			if err != nil || id.String() != text {
				return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
			}
			action.LinkedRecordID = &id
		case "remove_record_ref":
			if !mutationObjectHasOnlyFields(object, "op", "item_ref") {
				return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
			}
			if json.Unmarshal(object["item_ref"], &action.ItemRef) != nil {
				return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
			}
			if _, err := links.ParseRecordRefItemRef(action.ItemRef); err != nil {
				return CollectionActionPayload{}, invalidAdmission(fieldKey, "invalid_value")
			}
		}
		payload.Actions = append(payload.Actions, action)
	}
	return payload, nil
}

func isMutationView(viewSchemaID string) bool {
	return viewSchemaID == TaskRequestsViewSchemaID || viewSchemaID == DecisionsViewSchemaID
}

func mutationObjectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[field] = struct{}{}
		if _, present := object[field]; !present {
			return false
		}
	}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return false
		}
	}
	return true
}

func decodeMutationInteger(raw json.RawMessage) (int64, bool) {
	var value int64
	if json.Unmarshal(raw, &value) == nil {
		return value, true
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return 0, false
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	return parsed, err == nil
}

func canonicalMutationValues(values map[string]FieldValue) map[string]any {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		if value := canonicalMutationValue(values[key]); value != nil {
			result[key] = value
		}
	}
	return result
}

func canonicalMutationValue(value FieldValue) any {
	switch {
	case value.Timestamp != nil:
		return value.Timestamp.UTC().Format(time.RFC3339Nano)
	case value.UUID != nil:
		return value.UUID.String()
	case value.Text != nil:
		return *value.Text
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func canonicalMutationCollections(collections map[string]CollectionActionPayload) map[string]any {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		result[key] = canonicalMutationCollection(collections[key])
	}
	return result
}

func canonicalMutationCollection(payload CollectionActionPayload) map[string]any {
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.LinkedRecordID != nil {
			entry["linked_record_id"] = action.LinkedRecordID.String()
		}
		if action.ItemRef != "" {
			entry["item_ref"] = action.ItemRef
		}
		actions = append(actions, entry)
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func hashMutationPayload(value any) []byte {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}
