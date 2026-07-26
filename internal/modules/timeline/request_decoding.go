package timeline

import (
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func DecodeViewQueryRequest(reader io.Reader, viewSchemaID string) (viewschema.QueryMeta, *httpapi.APIError) {
	query, err := viewquery.Decode(reader, viewSchemaID)
	if err != nil {
		return viewschema.QueryMeta{}, invalidViewQueryValidation(err)
	}
	return query.Meta, nil
}

func DecodeTimelineCreateRequest(reader io.Reader) (CreateRequest, *httpapi.APIError) {
	schema, found := viewschema.Lookup(TimelineViewSchemaID)
	if !found {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader, invalidMutationPayload)
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

	var ok bool
	if request.DateEnteredText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.date_entered_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.date_entered_text", "invalid_value")
	}
	if request.AnalystText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.analyst_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.analyst_text", "invalid_value")
	}
	if request.MitreStageText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.mitre_stage_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.mitre_stage_text", "invalid_value")
	}
	if request.DeviceObjectText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.device_object_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.device_object_text", "invalid_value")
	}
	if request.IPAddressText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.ip_address_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.ip_address_text", "invalid_value")
	}
	if request.ActivityUTCText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.activity_utc_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.activity_utc_text", "invalid_value")
	}
	if request.ActivityLocalText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.activity_local_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.activity_local_text", "invalid_value")
	}
	if request.RawActivityText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.raw_activity_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.raw_activity_text", "invalid_value")
	}
	if request.ActivitySynopsisText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.activity_synopsis_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.activity_synopsis_text", "invalid_value")
	}
	if request.DataSourceText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.data_source_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.data_source_text", "invalid_value")
	}
	if request.HostRefs, apiErr = decodeCreateCollectionActionField(raw, "timeline.host_refs"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if request.IdentityRefs, apiErr = decodeCreateCollectionActionField(raw, "timeline.identity_refs"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if request.Tags, apiErr = decodeCreateCollectionActionField(raw, "timeline.tags"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if request.AttachedEvidence, apiErr = decodeCreateCollectionActionField(raw, "timeline.attached_evidence_ids"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if !schema.PermitsZeroFieldCreate && !CreateRequestHasUserValue(request) {
		return CreateRequest{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}

	return request, nil
}

func DecodeTimelinePatchRequest(reader io.Reader) (PatchRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
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
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != TimelineViewSchemaID {
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

	seen := make(map[string]struct{}, len(rawChanges))
	request.CanonicalChange = make([]PatchChange, 0, len(rawChanges))
	for index, rawChange := range rawChanges {
		change, apiErr := decodePatchChange(rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, ok := seen[change.FieldKey]; ok {
			return PatchRequest{}, invalidMutationPayload("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.CanonicalChange = append(request.CanonicalChange, change)
		_ = index
	}
	slices.SortFunc(request.CanonicalChange, func(left PatchChange, right PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func DecodeTimelineConflictResolveRequest(reader io.Reader, token string, claims TimelineConflictTokenClaims) (ConflictResolveRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
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

	field, ok := viewschema.LookupField(TimelineViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeCollectionActionPayload(claims.FieldKey, resolvedValue, claims.FieldKey, "resolved_value.actions")
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.ActionPayload = payload
		change.CanonicalAny = canonicalCollectionActionPayload(payload)
	} else {
		if _, ok := directWritableFieldKeys[claims.FieldKey]; !ok {
			return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
		}
		textValue, ok := normalizeFieldTextValue(claims.FieldKey, resolvedValue)
		if !ok {
			return ConflictResolveRequest{}, invalidMutationPayload(claims.FieldKey, "invalid_value")
		}
		change.TextValue = textValue
		change.CanonicalAny = canonicalChangeValue(change)
	}
	request.ResolvedChange = &change
	request.CanonicalAny = change.CanonicalAny
	return request, nil
}

func DecodeTimelineActionRequest(reader io.Reader) (ActionRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ActionRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request ActionRequest
	if value, ok := raw["base_row_version"]; !ok {
		return ActionRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return ActionRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return ActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	var ok bool
	if request.Reason, ok = normalizeNullableNoteField(raw, "reason"); !ok {
		return ActionRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	return request, nil
}

func DecodeTimelineSupersedeRequest(reader io.Reader) (SupersedeRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return SupersedeRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_row_version":      {},
		"client_txn_id":         {},
		"reason":                {},
		"replacement_record_id": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return SupersedeRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request SupersedeRequest
	if value, ok := raw["base_row_version"]; !ok {
		return SupersedeRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return SupersedeRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return SupersedeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return SupersedeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	value, ok := raw["reason"]
	if !ok {
		return SupersedeRequest{}, invalidMutationPayload("reason", "missing_required_field")
	}
	reason, ok := normalizeNoteValue(value)
	if !ok {
		return SupersedeRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	request.Reason = reason

	if replacementValue, ok := raw["replacement_record_id"]; ok {
		if string(replacementValue) == "null" {
			return SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "field_not_nullable")
		} else {
			var rawID string
			if err := json.Unmarshal(replacementValue, &rawID); err != nil {
				return SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "invalid_value")
			}
			replacementID, err := uuid.Parse(rawID)
			if err != nil {
				return SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "invalid_value")
			}
			request.ReplacementRecordID = &replacementID
		}
	}

	return request, nil
}

func DecodeTimelineTimeConversionProfilePutRequest(reader io.Reader) (TimeConversionProfilePutRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return TimeConversionProfilePutRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_profile_version": {},
		"enabled":              {},
		"local_offset_minutes": {},
		"local_label":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return TimeConversionProfilePutRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request TimeConversionProfilePutRequest
	if value, ok := raw["base_profile_version"]; !ok {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("base_profile_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseProfileVersion); err != nil || request.BaseProfileVersion < 1 {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("base_profile_version", "invalid_value")
	}
	if value, ok := raw["enabled"]; !ok {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("enabled", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Enabled); err != nil {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("enabled", "invalid_value")
	}
	offsetValue, ok := raw["local_offset_minutes"]
	if !ok {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("local_offset_minutes", "missing_required_field")
	}
	if string(offsetValue) != "null" {
		var offset int
		if err := json.Unmarshal(offsetValue, &offset); err != nil || offset < -840 || offset > 840 {
			return TimeConversionProfilePutRequest{}, invalidMutationPayload("local_offset_minutes", "invalid_value")
		}
		request.LocalOffsetMinutes = &offset
	}
	if request.Enabled && request.LocalOffsetMinutes == nil {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("local_offset_minutes", "missing_required_field")
	}
	var okLabel bool
	if request.LocalLabel, okLabel = normalizeNullableLineField(raw, "local_label"); !okLabel {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("local_label", "invalid_value")
	}
	if _, ok := raw["local_label"]; !ok {
		return TimeConversionProfilePutRequest{}, invalidMutationPayload("local_label", "missing_required_field")
	}
	return request, nil
}

func decodePatchChange(raw json.RawMessage) (PatchChange, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}

	allowed := map[string]struct{}{
		"field_key":      {},
		"value":          {},
		"action_payload": {},
	}
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
	field, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}

	change := PatchChange{FieldKey: fieldKey}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}

	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return PatchChange{}, invalidMutationPayload("action_payload", "missing_required_field")
		}
		payload, apiErr := decodeCollectionActionPayload(fieldKey, actionPayload, fieldKey, "changes.action_payload.actions")
		if apiErr != nil {
			return PatchChange{}, apiErr
		}
		change.ActionPayload = payload
		return change, nil
	}

	if !hasValue {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	if _, ok := directWritableFieldKeys[fieldKey]; !ok {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	textValue, ok := normalizeFieldTextValue(fieldKey, value)
	if !ok {
		return PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	change.TextValue = textValue
	return change, nil
}

func normalizeFieldTextValue(fieldKey string, value json.RawMessage) (*string, bool) {
	if _, ok := directWritableFieldKeys[fieldKey]; !ok {
		return nil, false
	}
	return normalizeNullableTimelineVisibleTextValue(value)
}

func decodeCreateCollectionActionField(raw map[string]json.RawMessage, fieldKey string) (*CollectionActionPayload, *httpapi.APIError) {
	value, ok := raw[fieldKey]
	if !ok {
		return nil, nil
	}
	payload, apiErr := decodeCollectionActionPayload(fieldKey, value, fieldKey, fieldKey+".actions")
	if apiErr != nil {
		return nil, apiErr
	}
	for _, action := range payload.Actions {
		if isTimelineAttachedEvidenceCollection(fieldKey) {
			if action.Op != "add_record_ref" {
				return nil, invalidMutationPayload(fieldKey, "invalid_value")
			}
			continue
		}
		if isTimelineTagCollection(fieldKey) {
			if action.Op != "add_tag" {
				return nil, invalidMutationPayload(fieldKey, "invalid_value")
			}
			continue
		}
		if action.Op != "add_token" && action.Op != "add_resolved_ref" {
			return nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
	}
	return payload, nil
}

func decodeCollectionActionPayload(fieldKey string, raw json.RawMessage, invalidField string, actionsField string) (*CollectionActionPayload, *httpapi.APIError) {
	policy, ok := timelineCollectionPolicy(fieldKey)
	if !ok {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}

	var payloadObject map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payloadObject); err != nil {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	if !objectHasOnlyFields(payloadObject, "kind", "actions") {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}

	var kind string
	if err := json.Unmarshal(payloadObject["kind"], &kind); err != nil {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	var rawActions []json.RawMessage
	if err := json.Unmarshal(payloadObject["actions"], &rawActions); err != nil {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	if kind != "collection_actions_v1" {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	if len(rawActions) == 0 {
		return nil, invalidMutationPayloadWithDetails(actionsField, "empty_collection_actions", map[string]any{
			"field_key": fieldKey,
		})
	}
	if len(rawActions) > maxCollectionActions {
		return nil, invalidMutationPayloadWithDetails(actionsField, "collection_action_count_exceeded", map[string]any{
			"field_key":       fieldKey,
			"requested_count": len(rawActions),
			"max_count":       maxCollectionActions,
		})
	}

	actions := make([]CollectionAction, 0, len(rawActions))
	for _, rawActionData := range rawActions {
		var rawAction map[string]json.RawMessage
		if err := json.Unmarshal(rawActionData, &rawAction); err != nil {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
		opValue, ok := rawAction["op"]
		if !ok {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}

		var op string
		if err := json.Unmarshal(opValue, &op); err != nil {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
		if !policy.AllowsOp(op) {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
		switch op {
		case "add_token":
			if !isTimelineMentionCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "raw_text"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			rawTextValue, ok := rawAction["raw_text"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var rawText string
			if err := json.Unmarshal(rawTextValue, &rawText); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			normalized, ok := normalizeCollectionToken(fieldKey, rawText)
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        rawText,
				NormalizedText: normalized,
			})
		case "add_tag":
			if !isTimelineTagCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "tag_name"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			tagNameValue, ok := rawAction["tag_name"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var tagName string
			if err := json.Unmarshal(tagNameValue, &tagName); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			label, normalized, ok := fieldnorm.NormalizeTagLabel(tagName)
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        label,
				NormalizedText: normalized,
			})
		case "add_resolved_ref":
			if !isTimelineMentionCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "raw_text", "resolved_record_id"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			rawTextValue, ok := rawAction["raw_text"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			resolvedRecordValue, ok := rawAction["resolved_record_id"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var rawText string
			if err := json.Unmarshal(rawTextValue, &rawText); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			normalized, ok := normalizeCollectionToken(fieldKey, rawText)
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var resolvedRecordID string
			if err := json.Unmarshal(resolvedRecordValue, &resolvedRecordID); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			parsed, err := uuid.Parse(resolvedRecordID)
			if err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        rawText,
				NormalizedText: normalized,
				ResolvedRecord: &parsed,
			})
		case "add_record_ref":
			if !isTimelineAttachedEvidenceCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "linked_record_id"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			linkedRecordValue, ok := rawAction["linked_record_id"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var linkedRecordID string
			if err := json.Unmarshal(linkedRecordValue, &linkedRecordID); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			parsed, err := uuid.Parse(linkedRecordID)
			if err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, LinkedRecordID: &parsed})
		case "resolve_item":
			if !isTimelineMentionCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref", "resolved_record_id"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			action := CollectionAction{Op: op, ItemRef: itemRef}
			resolvedRecordValue := rawAction["resolved_record_id"]
			var resolvedRecordID string
			if err := json.Unmarshal(resolvedRecordValue, &resolvedRecordID); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			parsed, err := uuid.Parse(resolvedRecordID)
			if err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			action.ResolvedRecord = &parsed
			actions = append(actions, action)
		case "dismiss_item", "revert_to_unresolved":
			if !isTimelineMentionCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, ItemRef: itemRef})
		case "remove_record_ref":
			if !isTimelineAttachedEvidenceCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, ItemRef: itemRef})
		case "remove_tag":
			if !isTimelineTagCollection(fieldKey) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, ItemRef: itemRef})
		default:
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
	}
	return &CollectionActionPayload{Actions: actions}, nil
}

func normalizeCollectionToken(fieldKey string, rawText string) (string, bool) {
	if isTimelineTagCollection(fieldKey) {
		_, normalized, ok := fieldnorm.NormalizeTagLabel(rawText)
		return normalized, ok
	}
	return fieldnorm.NormalizeMentionToken(rawText)
}

func timelineCollectionPolicy(fieldKey string) (CollectionPolicy, bool) {
	policy, ok := LookupCollectionPolicy(fieldKey)
	if !ok {
		return CollectionPolicy{}, false
	}
	if policy.Family == CollectionFamilyMentionOrigin {
		return policy, true
	}
	if policy.AllowsLinksCollectionMutation() && (fieldKey == "timeline.tags" || fieldKey == "timeline.attached_evidence_ids") {
		return policy, true
	}
	return CollectionPolicy{}, false
}

func isTimelineMentionCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyMentionOrigin
}

func isTimelineTagCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyRecordTag
}

func isTimelineAttachedEvidenceCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyRecordRef && policy.LinkType == "attached_evidence"
}

func objectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
	required := make([]string, 0, len(fields))
	required = append(required, fields...)
	return actionHasOnlyFields(object, required, nil)
}

func actionHasOnlyFields(action map[string]json.RawMessage, required []string, optional []string) bool {
	allowed := make(map[string]struct{}, len(required)+len(optional))
	for _, key := range required {
		allowed[key] = struct{}{}
		if _, ok := action[key]; !ok {
			return false
		}
	}
	for _, key := range optional {
		allowed[key] = struct{}{}
	}
	for key := range action {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}

func decodeObject(reader io.Reader, invalid func(string, string) *httpapi.APIError) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalid("", "request_not_object")
	}
	return raw, nil
}

func normalizeNullableTimelineVisibleTextField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok {
		return nil, true
	}
	return normalizeNullableTimelineVisibleTextValue(value)
}

func normalizeNullableTimelineVisibleTextValue(value json.RawMessage) (*string, bool) {
	if string(value) == "null" {
		return nil, true
	}
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil {
		return nil, false
	}
	if !validTimelineVisibleText(rawValue) {
		return nil, false
	}
	return &rawValue, true
}

func validTimelineVisibleText(value string) bool {
	if len([]rune(value)) > 32768 {
		return false
	}
	for _, r := range value {
		if r == 0 {
			return false
		}
		if (r < 0x20 || (r >= 0x7f && r <= 0x9f)) && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

func normalizeNullableLineField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok {
		return nil, true
	}
	return normalizeNullableLineValue(value)
}

func normalizeNullableLineValue(value json.RawMessage) (*string, bool) {
	if string(value) == "null" {
		return nil, true
	}
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil {
		return nil, false
	}
	normalized, ok := normalizeLine(rawValue)
	if !ok {
		return nil, false
	}
	return &normalized, true
}

func normalizeNullableNoteField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok {
		return nil, true
	}
	return normalizeNullableNoteValuePointer(value)
}

func normalizeNullableNoteValuePointer(value json.RawMessage) (*string, bool) {
	if string(value) == "null" {
		return nil, true
	}
	note, ok := normalizeNoteValue(value)
	if !ok {
		return nil, false
	}
	return &note, true
}

func normalizeNoteValue(value json.RawMessage) (string, bool) {
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil {
		return "", false
	}
	return normalizeNote(rawValue)
}

func normalizeLine(raw string) (string, bool) {
	return fieldnorm.NormalizeLine(raw)
}

func normalizeNote(raw string) (string, bool) {
	return fieldnorm.NormalizeNote(raw)
}
