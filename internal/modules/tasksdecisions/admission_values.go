package tasksdecisions

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

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
	if !isRecordRefCollectionField(fieldKey) {
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
		if json.Unmarshal(object["op"], &op) != nil || !allowsCollectionOp(fieldKey, op) {
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
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return false
	}
	_, ok := catalog.SurfaceByViewID(viewSchemaID)
	return ok
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
