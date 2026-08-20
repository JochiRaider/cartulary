package admission

import (
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/google/uuid"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
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

func DecodeTimelineCreateRequest(reader io.Reader) (timeline.CreateRequest, *httpapi.APIError) {
	schema, found := viewschema.Lookup(timeline.TimelineViewSchemaID)
	if !found {
		return timeline.CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return timeline.CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return timeline.CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request timeline.CreateRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return timeline.CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	var ok bool
	if request.DateEnteredText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.date_entered_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.date_entered_text", "invalid_value")
	}
	if request.AnalystText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.analyst_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.analyst_text", "invalid_value")
	}
	if request.MitreStageText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.mitre_stage_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.mitre_stage_text", "invalid_value")
	}
	if request.DeviceObjectText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.device_object_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.device_object_text", "invalid_value")
	}
	if request.IPAddressText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.ip_address_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.ip_address_text", "invalid_value")
	}
	if request.ActivityUTCText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.activity_utc_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.activity_utc_text", "invalid_value")
	}
	if request.ActivityLocalText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.activity_local_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.activity_local_text", "invalid_value")
	}
	if request.RawActivityText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.raw_activity_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.raw_activity_text", "invalid_value")
	}
	if request.ActivitySynopsisText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.activity_synopsis_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.activity_synopsis_text", "invalid_value")
	}
	if request.DataSourceText, ok = normalizeNullableTimelineVisibleTextField(raw, "timeline.data_source_text"); !ok {
		return timeline.CreateRequest{}, invalidMutationPayload("timeline.data_source_text", "invalid_value")
	}
	if request.HostRefs, apiErr = decodeCreateCollectionActionField(raw, "timeline.host_refs"); apiErr != nil {
		return timeline.CreateRequest{}, apiErr
	}
	if request.IdentityRefs, apiErr = decodeCreateCollectionActionField(raw, "timeline.identity_refs"); apiErr != nil {
		return timeline.CreateRequest{}, apiErr
	}
	if request.Tags, apiErr = decodeCreateCollectionActionField(raw, "timeline.tags"); apiErr != nil {
		return timeline.CreateRequest{}, apiErr
	}
	if request.AttachedEvidence, apiErr = decodeCreateCollectionActionField(raw, "timeline.attached_evidence_ids"); apiErr != nil {
		return timeline.CreateRequest{}, apiErr
	}
	if !schema.PermitsZeroFieldCreate && !timeline.CreateRequestHasUserValue(request) {
		return timeline.CreateRequest{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}

	return request, nil
}

func DecodeTimelinePatchRequest(reader io.Reader) (timeline.PatchRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return timeline.PatchRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"view_schema_id":   {},
		"base_row_version": {},
		"client_txn_id":    {},
		"changes":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return timeline.PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request timeline.PatchRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return timeline.PatchRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != timeline.TimelineViewSchemaID {
		return timeline.PatchRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, ok := raw["base_row_version"]; !ok {
		return timeline.PatchRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return timeline.PatchRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return timeline.PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return timeline.PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	value, ok := raw["changes"]
	if !ok {
		return timeline.PatchRequest{}, invalidMutationPayload("changes", "missing_required_field")
	}
	var rawChanges []json.RawMessage
	if err := json.Unmarshal(value, &rawChanges); err != nil {
		return timeline.PatchRequest{}, invalidMutationPayload("changes", "invalid_value")
	}
	if len(rawChanges) == 0 {
		return timeline.PatchRequest{}, invalidMutationPayload("changes", "empty_changes")
	}
	if len(rawChanges) > mutationpolicy.MaxPatchChanges {
		return timeline.PatchRequest{}, invalidMutationPayloadWithDetails("changes", "change_count_exceeded", map[string]any{
			"requested_count": len(rawChanges),
			"max_count":       mutationpolicy.MaxPatchChanges,
		})
	}

	seen := make(map[string]struct{}, len(rawChanges))
	request.CanonicalChange = make([]timeline.PatchChange, 0, len(rawChanges))
	for index, rawChange := range rawChanges {
		change, apiErr := decodePatchChange(rawChange)
		if apiErr != nil {
			return timeline.PatchRequest{}, apiErr
		}
		if _, ok := seen[change.FieldKey]; ok {
			return timeline.PatchRequest{}, invalidMutationPayload("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.CanonicalChange = append(request.CanonicalChange, change)
		_ = index
	}
	slices.SortFunc(request.CanonicalChange, func(left timeline.PatchChange, right timeline.PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func DecodeTimelineConflictResolveRequest(reader io.Reader, token string, claims conflicttokens.ConflictTokenClaims) (timeline.ConflictResolveRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return timeline.ConflictResolveRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"conflict_token":  {},
		"resolution_kind": {},
		"client_txn_id":   {},
		"resolved_value":  {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return timeline.ConflictResolveRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	request := timeline.ConflictResolveRequest{ConflictToken: token}
	if value, ok := raw["conflict_token"]; !ok {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ConflictToken); err != nil || request.ConflictToken != token {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, ok := raw["resolution_kind"]; !ok {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ResolutionKind); err != nil {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	resolvedValue, hasResolvedValue := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if hasResolvedValue {
			return timeline.ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !hasResolvedValue {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "missing_required_field")
	}

	field, ok := viewschema.LookupField(timeline.TimelineViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return timeline.ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	change := timeline.PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeCollectionActionPayload(claims.FieldKey, resolvedValue, claims.FieldKey, "resolved_value.actions")
		if apiErr != nil {
			return timeline.ConflictResolveRequest{}, apiErr
		}
		change.ActionPayload = payload
		change.CanonicalAny = payload.CanonicalValue()
	} else {
		if !mutationpolicy.IsDirectWritableField(claims.FieldKey) {
			return timeline.ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
		}
		textValue, ok := normalizeFieldTextValue(claims.FieldKey, resolvedValue)
		if !ok {
			return timeline.ConflictResolveRequest{}, invalidMutationPayload(claims.FieldKey, "invalid_value")
		}
		change.TextValue = textValue
		change.CanonicalAny = change.CanonicalValue()
	}
	request.ResolvedChange = &change
	request.CanonicalAny = change.CanonicalAny
	return request, nil
}

func DecodeTimelineActionRequest(reader io.Reader) (timeline.ActionRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return timeline.ActionRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return timeline.ActionRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request timeline.ActionRequest
	if value, ok := raw["base_row_version"]; !ok {
		return timeline.ActionRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return timeline.ActionRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return timeline.ActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return timeline.ActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	var ok bool
	if request.Reason, ok = normalizeNullableNoteField(raw, "reason"); !ok {
		return timeline.ActionRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	return request, nil
}

func DecodeTimelineSupersedeRequest(reader io.Reader) (timeline.SupersedeRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return timeline.SupersedeRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_row_version":      {},
		"client_txn_id":         {},
		"reason":                {},
		"replacement_record_id": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return timeline.SupersedeRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request timeline.SupersedeRequest
	if value, ok := raw["base_row_version"]; !ok {
		return timeline.SupersedeRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return timeline.SupersedeRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return timeline.SupersedeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return timeline.SupersedeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	value, ok := raw["reason"]
	if !ok {
		return timeline.SupersedeRequest{}, invalidMutationPayload("reason", "missing_required_field")
	}
	reason, ok := normalizeNoteValue(value)
	if !ok {
		return timeline.SupersedeRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	request.Reason = reason

	if replacementValue, ok := raw["replacement_record_id"]; ok {
		if string(replacementValue) == "null" {
			return timeline.SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "field_not_nullable")
		} else {
			var rawID string
			if err := json.Unmarshal(replacementValue, &rawID); err != nil {
				return timeline.SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "invalid_value")
			}
			replacementID, err := uuid.Parse(rawID)
			if err != nil {
				return timeline.SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "invalid_value")
			}
			request.ReplacementRecordID = &replacementID
		}
	}

	return request, nil
}

func DecodeTimelineTimeConversionProfilePutRequest(reader io.Reader) (timeline.TimeConversionProfilePutRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return timeline.TimeConversionProfilePutRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_profile_version": {},
		"enabled":              {},
		"local_offset_minutes": {},
		"local_label":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request timeline.TimeConversionProfilePutRequest
	if value, ok := raw["base_profile_version"]; !ok {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("base_profile_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseProfileVersion); err != nil || request.BaseProfileVersion < 1 {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("base_profile_version", "invalid_value")
	}
	if value, ok := raw["enabled"]; !ok {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("enabled", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Enabled); err != nil {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("enabled", "invalid_value")
	}
	offsetValue, ok := raw["local_offset_minutes"]
	if !ok {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("local_offset_minutes", "missing_required_field")
	}
	if string(offsetValue) != "null" {
		var offset int
		if err := json.Unmarshal(offsetValue, &offset); err != nil || offset < -840 || offset > 840 {
			return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("local_offset_minutes", "invalid_value")
		}
		request.LocalOffsetMinutes = &offset
	}
	if request.Enabled && request.LocalOffsetMinutes == nil {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("local_offset_minutes", "missing_required_field")
	}
	var okLabel bool
	if request.LocalLabel, okLabel = normalizeNullableLineField(raw, "local_label"); !okLabel {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("local_label", "invalid_value")
	}
	if _, ok := raw["local_label"]; !ok {
		return timeline.TimeConversionProfilePutRequest{}, invalidMutationPayload("local_label", "missing_required_field")
	}
	return request, nil
}

func decodePatchChange(raw json.RawMessage) (timeline.PatchChange, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return timeline.PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}

	allowed := map[string]struct{}{
		"field_key":      {},
		"value":          {},
		"action_payload": {},
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return timeline.PatchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}

	fieldValue, ok := object["field_key"]
	if !ok {
		return timeline.PatchChange{}, invalidMutationPayload("changes", "missing_field_key")
	}
	var fieldKey string
	if err := json.Unmarshal(fieldValue, &fieldKey); err != nil {
		return timeline.PatchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(timeline.TimelineViewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return timeline.PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}

	change := timeline.PatchChange{FieldKey: fieldKey}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return timeline.PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}

	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return timeline.PatchChange{}, invalidMutationPayload("action_payload", "missing_required_field")
		}
		payload, apiErr := decodeCollectionActionPayload(fieldKey, actionPayload, fieldKey, "changes.action_payload.actions")
		if apiErr != nil {
			return timeline.PatchChange{}, apiErr
		}
		change.ActionPayload = payload
		return change, nil
	}

	if !hasValue {
		return timeline.PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	if !mutationpolicy.IsDirectWritableField(fieldKey) {
		return timeline.PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	textValue, ok := normalizeFieldTextValue(fieldKey, value)
	if !ok {
		return timeline.PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	change.TextValue = textValue
	return change, nil
}

func normalizeFieldTextValue(fieldKey string, value json.RawMessage) (*string, bool) {
	if !mutationpolicy.IsDirectWritableField(fieldKey) {
		return nil, false
	}
	return normalizeNullableTimelineVisibleTextValue(value)
}

func decodeCreateCollectionActionField(raw map[string]json.RawMessage, fieldKey string) (*timeline.CollectionActionPayload, *httpapi.APIError) {
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

func decodeCollectionActionPayload(fieldKey string, raw json.RawMessage, invalidField string, actionsField string) (*timeline.CollectionActionPayload, *httpapi.APIError) {
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
	if len(rawActions) > mutationpolicy.MaxCollectionActions {
		return nil, invalidMutationPayloadWithDetails(actionsField, "collection_action_count_exceeded", map[string]any{
			"field_key":       fieldKey,
			"requested_count": len(rawActions),
			"max_count":       mutationpolicy.MaxCollectionActions,
		})
	}

	actions := make([]timeline.CollectionAction, 0, len(rawActions))
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
			actions = append(actions, timeline.CollectionAction{
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
			actions = append(actions, timeline.CollectionAction{
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
			actions = append(actions, timeline.CollectionAction{
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
			actions = append(actions, timeline.CollectionAction{Op: op, LinkedRecordID: &parsed})
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
			action := timeline.CollectionAction{Op: op, ItemRef: itemRef}
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
			actions = append(actions, timeline.CollectionAction{Op: op, ItemRef: itemRef})
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
			actions = append(actions, timeline.CollectionAction{Op: op, ItemRef: itemRef})
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
			actions = append(actions, timeline.CollectionAction{Op: op, ItemRef: itemRef})
		default:
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
	}
	return &timeline.CollectionActionPayload{Actions: actions}, nil
}

func normalizeCollectionToken(fieldKey string, rawText string) (string, bool) {
	if isTimelineTagCollection(fieldKey) {
		_, normalized, ok := fieldnorm.NormalizeTagLabel(rawText)
		return normalized, ok
	}
	return fieldnorm.NormalizeMentionToken(rawText)
}

func timelineCollectionPolicy(fieldKey string) (timeline.CollectionPolicy, bool) {
	policy, ok := timeline.LookupCollectionPolicy(fieldKey)
	if !ok {
		return timeline.CollectionPolicy{}, false
	}
	if policy.Family == timeline.CollectionFamilyMentionOrigin {
		return policy, true
	}
	if policy.AllowsLinksCollectionMutation() && (fieldKey == "timeline.tags" || fieldKey == "timeline.attached_evidence_ids") {
		return policy, true
	}
	return timeline.CollectionPolicy{}, false
}

func isTimelineMentionCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == timeline.CollectionFamilyMentionOrigin
}

func isTimelineTagCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == timeline.CollectionFamilyRecordTag
}

func isTimelineAttachedEvidenceCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == timeline.CollectionFamilyRecordRef && policy.LinkType == "attached_evidence"
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
	if !mutationpolicy.IsValidVisibleText(rawValue) {
		return nil, false
	}
	return &rawValue, true
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
