package artifacts

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/cases"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	maxMutationPatchChanges      = 32
	maxMutationCollectionActions = 64
)

var riskReferenceCaseFolder = cases.Fold()

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

// DecodeCreateRequest admits one of the fixed Artifact create surfaces. All
// view-specific normalization and minimum-create rules remain owner-local.
func DecodeCreateRequest(viewSchemaID string, reader io.Reader) (CreateRequest, *httpapi.APIError) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok || !schema.CreateCapable || !isArtifactBackedView(viewSchemaID) {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return CreateRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	request := CreateRequest{
		ViewSchemaID: viewSchemaID,
		Values:       map[string]FieldValue{},
		Collections:  map[string]CollectionActionPayload{},
	}
	if value, present := raw["client_txn_id"]; !present {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		if field.ConflictResolutionClass == "collection_review" {
			payload, apiErr := decodeArtifactCollectionActionPayload(fieldKey, value)
			if apiErr != nil {
				return CreateRequest{}, apiErr
			}
			request.Collections[fieldKey] = payload
			continue
		}
		admitted, _, apiErr := decodeArtifactValue(fieldKey, field, value, false)
		if apiErr != nil {
			return CreateRequest{}, apiErr
		}
		request.Values[fieldKey] = admitted
	}
	if err := validateCreateParams(createParams{ViewSchemaID: viewSchemaID, Values: request.Values}); err != nil {
		var validation *ValidationError
		if errors.As(err, &validation) {
			return CreateRequest{}, invalidMutationPayload(validation.Field, validation.ReasonCode)
		}
		return CreateRequest{}, invalidMutationPayload("payload", "invalid_value")
	}
	return request, nil
}

func DecodeContextualNoteCreateRequest(reader io.Reader) (ContextualNoteCreateRequest, *httpapi.APIError) {
	request, apiErr := DecodeCreateRequest(NotesViewSchemaID, reader)
	if apiErr != nil {
		return ContextualNoteCreateRequest{}, apiErr
	}
	return ContextualNoteCreateRequest{
		ClientTxnID: request.ClientTxnID,
		Values:      request.Values,
		Collections: request.Collections,
	}, nil
}

func DecodePatchRequest(reader io.Reader) (PatchRequest, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return PatchRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request PatchRequest
	if value, present := raw["view_schema_id"]; !present {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ViewSchemaID) != nil || !isArtifactBackedView(request.ViewSchemaID) {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, present := raw["base_row_version"]; !present {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	rawChanges, apiErr := decodeArtifactRawChanges(raw["changes"])
	if apiErr != nil {
		return PatchRequest{}, apiErr
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, apiErr := decodeArtifactPatchChange(request.ViewSchemaID, rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, duplicate := seen[change.FieldKey]; duplicate {
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

func DecodeConflictResolveRequest(
	reader io.Reader,
	token string,
	claims ConflictClaims,
) (ConflictResolveRequest, *httpapi.APIError) {
	if claims.RecordID == uuid.Nil || !isArtifactBackedView(claims.ViewSchemaID) || claims.CurrentRowVersion < 1 {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return ConflictResolveRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	request := ConflictResolveRequest{ConflictToken: token}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	resolvedValue, present := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if present {
			return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !present {
		return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "missing_required_field")
	}
	field, ok := viewschema.LookupField(claims.ViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeArtifactCollectionActionPayload(claims.FieldKey, resolvedValue)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Collection = &payload
		change.CanonicalValue = canonicalArtifactCollectionPayload(payload)
	} else {
		value, canonical, apiErr := decodeArtifactValue(claims.FieldKey, field, resolvedValue, true)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Value = &value
		change.CanonicalValue = canonical
	}
	request.Patch = &PatchRequest{
		ViewSchemaID: claims.ViewSchemaID, BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID: request.ClientTxnID, Changes: []PatchChange{change},
	}
	request.CanonicalValue = change.CanonicalValue
	return request, nil
}

func CreateRequestHash(request CreateRequest) []byte {
	return hashArtifactMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"values":         canonicalArtifactValues(request.Values),
		"collection_ops": canonicalArtifactCollections(request.Collections),
		"create_inputs":  map[string]any{},
	})
}

func PatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": change.CanonicalValue})
	}
	return hashArtifactMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID, "base_row_version": request.BaseRowVersion, "changes": changes,
	})
}

func ConflictResolveRequestHash(claims ConflictClaims, request ConflictResolveRequest) []byte {
	return hashArtifactMutationPayload(map[string]any{
		"conflict_token": request.ConflictToken, "resolution_kind": request.ResolutionKind,
		"record_id": claims.RecordID, "view_schema_id": claims.ViewSchemaID,
		"field_key": claims.FieldKey, "current_row_version": claims.CurrentRowVersion,
		"resolved_value": request.CanonicalValue,
	})
}

func ContextualNoteCreateRequestHash(sourceRecordID uuid.UUID, request ContextualNoteCreateRequest) []byte {
	return hashArtifactMutationPayload(map[string]any{
		"source_record_id": sourceRecordID.String(), "view_schema_id": NotesViewSchemaID,
		"values": canonicalArtifactValues(request.Values), "collection_ops": canonicalArtifactCollections(request.Collections),
	})
}

func decodeArtifactRawChanges(raw json.RawMessage) ([]json.RawMessage, *httpapi.APIError) {
	if raw == nil {
		return nil, invalidMutationPayload("changes", "missing_required_field")
	}
	var changes []json.RawMessage
	if json.Unmarshal(raw, &changes) != nil {
		return nil, invalidMutationPayload("changes", "invalid_value")
	}
	if len(changes) == 0 {
		return nil, invalidMutationPayload("changes", "empty_changes")
	}
	if len(changes) > maxMutationPatchChanges {
		return nil, invalidMutationPayloadWithDetails("changes", "change_count_exceeded", map[string]any{
			"requested_count": len(changes), "max_count": maxMutationPatchChanges,
		})
	}
	return changes, nil
}

func decodeArtifactPatchChange(viewSchemaID string, raw json.RawMessage) (PatchChange, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return PatchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return PatchChange{}, invalidMutationPayload("changes", "missing_field_key")
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return PatchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidMutationPayload(fieldKey, "unsupported_field_key")
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
		payload, apiErr := decodeArtifactCollectionActionPayload(fieldKey, actionPayload)
		if apiErr != nil {
			return PatchChange{}, apiErr
		}
		change.Collection = &payload
		change.CanonicalValue = canonicalArtifactCollectionPayload(payload)
		return change, nil
	}
	if !hasValue {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	direct, canonical, apiErr := decodeArtifactValue(fieldKey, field, value, true)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	change.Value = &direct
	change.CanonicalValue = canonical
	return change, nil
}

func decodeArtifactValue(fieldKey string, field viewschema.Field, raw json.RawMessage, patch bool) (FieldValue, any, *httpapi.APIError) {
	if string(raw) == "null" {
		if !patch || field.Clearable {
			return FieldValue{}, nil, nil
		}
		return FieldValue{}, nil, invalidMutationPayload(fieldKey, "field_not_nullable")
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		normalized, ok := fieldnorm.NormalizeTimestampInstant(text)
		if !ok {
			return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return FieldValue{Timestamp: &normalized}, normalized.Format(time.RFC3339Nano), nil
	}
	if field.DirectReferenceContractID != nil || strings.HasSuffix(fieldKey, "_user_id") {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		parsed, err := uuid.Parse(strings.TrimSpace(text))
		if err != nil || (field.DirectReferenceContractID != nil && parsed.String() != text) {
			return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return FieldValue{UUID: &parsed}, parsed.String(), nil
	}
	if field.ReadKind == "number" {
		parsed, ok := decodeArtifactInteger(raw)
		if !ok {
			return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return FieldValue{Number: &parsed}, parsed, nil
	}
	if field.ReadKind == "boolean" {
		parsed, ok := decodeArtifactBoolean(raw)
		if !ok {
			return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return FieldValue{Bool: &parsed}, parsed, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var normalized string
	var ok bool
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		normalized, ok = fieldnorm.NormalizeNote(text)
	} else {
		normalized, ok = fieldnorm.NormalizeLine(text)
	}
	if !ok {
		return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	return FieldValue{Text: &normalized}, normalized, nil
}

func decodeArtifactCollectionActionPayload(fieldKey string, raw json.RawMessage) (CollectionActionPayload, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !artifactObjectHasOnlyFields(object, "kind", "actions") {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var kind string
	if json.Unmarshal(object["kind"], &kind) != nil || kind != "collection_actions_v1" {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var rawActions []json.RawMessage
	if json.Unmarshal(object["actions"], &rawActions) != nil {
		return CollectionActionPayload{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	if len(rawActions) == 0 {
		return CollectionActionPayload{}, invalidMutationPayloadWithDetails(fieldKey+".actions", "empty_collection_actions", map[string]any{"field_key": fieldKey})
	}
	if len(rawActions) > maxMutationCollectionActions {
		return CollectionActionPayload{}, invalidMutationPayloadWithDetails(fieldKey+".actions", "collection_action_count_exceeded", map[string]any{
			"field_key": fieldKey, "requested_count": len(rawActions), "max_count": maxMutationCollectionActions,
		})
	}
	payload := CollectionActionPayload{Actions: make([]CollectionAction, 0, len(rawActions))}
	for _, rawAction := range rawActions {
		action, apiErr := decodeArtifactCollectionAction(fieldKey, rawAction)
		if apiErr != nil {
			return CollectionActionPayload{}, apiErr
		}
		payload.Actions = append(payload.Actions, action)
	}
	return payload, nil
}

func decodeArtifactCollectionAction(fieldKey string, raw json.RawMessage) (CollectionAction, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var op string
	if json.Unmarshal(object["op"], &op) != nil {
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	action := CollectionAction{Op: op}
	switch op {
	case "add_token":
		if !artifactObjectHasOnlyFields(object, "op", "raw_text") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		text, ok := artifactStringActionField(object, "raw_text")
		normalized, valid := fieldnorm.NormalizeLine(text)
		if !ok || !valid {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.RawText, action.NormalizedText = text, normalized
	case "add_tag":
		if !artifactObjectHasOnlyFields(object, "op", "tag_name") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		text, ok := artifactStringActionField(object, "tag_name")
		label, normalized, valid := fieldnorm.NormalizeTagLabel(text)
		if !ok || !valid {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.RawText, action.NormalizedText = label, normalized
	case "remove_tag":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		if _, _, err := links.ParseRecordTagItemRef(itemRef); err != nil {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	case "add_record_ref":
		if !artifactObjectHasOnlyFields(object, "op", "linked_record_id") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		id, ok := artifactUUIDActionField(object, "linked_record_id")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.LinkedRecordID = &id
	case "remove_record_ref":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		if _, err := links.ParseRecordRefItemRef(itemRef); err != nil {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	case "add_party_ref":
		if !artifactObjectHasOnlyFields(object, "op", "party_id") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		id, ok := artifactUUIDActionField(object, "party_id")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.PartyID = &id
	case "remove_party_ref":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		if _, err := links.ParsePartyRefItemRef(itemRef); err != nil {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	case "add_risk_ref":
		if !artifactObjectHasOnlyFields(object, "op", "risk_ref_text") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		text, ok := artifactStringActionField(object, "risk_ref_text")
		normalized, valid := fieldnorm.NormalizeLine(text)
		if !ok || !valid {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.RiskRefText = normalized
		action.NormalizedText = riskReferenceCaseFolder.String(normalized)
	case "remove_risk_ref":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		if _, err := ParseRiskRefItemRef(itemRef); err != nil {
			return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		action.ItemRef = itemRef
	default:
		return CollectionAction{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	return action, nil
}

func decodeArtifactInteger(raw json.RawMessage) (int64, bool) {
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

func decodeArtifactBoolean(raw json.RawMessage) (bool, bool) {
	var value bool
	if json.Unmarshal(raw, &value) == nil {
		return value, true
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return false, false
	}
	switch strings.TrimSpace(text) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

func artifactObjectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
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

func artifactStringActionField(object map[string]json.RawMessage, field string) (string, bool) {
	value, present := object[field]
	if !present {
		return "", false
	}
	var text string
	if json.Unmarshal(value, &text) != nil || strings.TrimSpace(text) == "" {
		return "", false
	}
	return text, true
}

func artifactUUIDActionField(object map[string]json.RawMessage, field string) (uuid.UUID, bool) {
	text, ok := artifactStringActionField(object, field)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func canonicalArtifactValues(values map[string]FieldValue) map[string]any {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		value := canonicalArtifactValue(values[key])
		if value != nil {
			result[key] = value
		}
	}
	return result
}

func canonicalArtifactValue(value FieldValue) any {
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

func canonicalArtifactCollections(collections map[string]CollectionActionPayload) map[string]any {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		result[key] = canonicalArtifactCollectionPayload(collections[key])
	}
	return result
}

func canonicalArtifactCollectionPayload(payload CollectionActionPayload) map[string]any {
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

func hashArtifactMutationPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *httpapi.APIError {
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
	return &httpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_mutation_payload",
		Message: "invalid mutation payload", Details: details,
	}
}
