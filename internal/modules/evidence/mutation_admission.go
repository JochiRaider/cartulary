package evidence

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const maxMutationPatchChanges = 32

func canonicalHashBytes(payload any) []byte {
	data, _ := json.Marshal(payload)
	return data
}

func decodeEvidenceRawChanges(raw json.RawMessage) ([]json.RawMessage, *AdmissionFailure) {
	if raw == nil {
		return nil, newAdmissionFailure("changes", admissionMissingRequiredField)
	}
	var changes []json.RawMessage
	if json.Unmarshal(raw, &changes) != nil {
		return nil, newAdmissionFailure("changes", admissionInvalidValue)
	}
	if len(changes) == 0 {
		return nil, newAdmissionFailure("changes", admissionEmptyChanges)
	}
	if len(changes) > maxMutationPatchChanges {
		return nil, newAdmissionLimitFailure("changes", admissionChangeCountExceeded, len(changes), maxMutationPatchChanges)
	}
	return changes, nil
}

func decodeEvidencePatchChange(raw json.RawMessage) (patchChange, *AdmissionFailure) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return patchChange{}, newAdmissionFailure("changes", admissionInvalidChange)
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return patchChange{}, newAdmissionFailure("changes", admissionUnknownField)
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return patchChange{}, newAdmissionFailure("changes", admissionMissingFieldKey)
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return patchChange{}, newAdmissionFailure("field_key", admissionInvalidValue)
	}
	field, ok := viewschema.LookupField(ViewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return patchChange{}, newAdmissionFailure(fieldKey, admissionUnsupportedFieldKey)
	}
	value, hasValue := object["value"]
	_, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return patchChange{}, newAdmissionFailure("changes", admissionInvalidChange)
	}
	if !hasValue {
		return patchChange{}, newAdmissionFailure("value", admissionMissingRequiredField)
	}
	admitted, canonical, failure := decodeEvidenceValue(fieldKey, field, value, true)
	if failure != nil {
		return patchChange{}, failure
	}
	return patchChange{FieldKey: fieldKey, Value: &admitted, CanonicalValue: canonical}, nil
}

func decodeEvidenceValue(
	fieldKey string,
	field viewschema.Field,
	raw json.RawMessage,
	patch bool,
) (FieldValue, any, *AdmissionFailure) {
	if string(raw) == "null" {
		if !patch || field.Clearable {
			return FieldValue{}, nil, nil
		}
		return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionFieldNotNullable)
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		var rawTime string
		if json.Unmarshal(raw, &rawTime) != nil {
			return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
		}
		normalized, ok := fieldnorm.NormalizeTimestampInstant(rawTime)
		if !ok {
			return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
		}
		return FieldValue{Timestamp: &normalized}, normalized.Format(time.RFC3339Nano), nil
	}
	if field.DirectReferenceContractID != nil {
		var rawID string
		if json.Unmarshal(raw, &rawID) != nil {
			return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
		}
		parsed, err := uuid.Parse(rawID)
		if err != nil || parsed.String() != rawID {
			return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
		}
		return FieldValue{UUID: &parsed}, parsed.String(), nil
	}
	if field.ReadKind == "number" {
		parsed, ok := decodeEvidenceInteger(raw)
		if !ok {
			return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
		}
		return FieldValue{Number: &parsed}, parsed, nil
	}
	if field.ReadKind == "boolean" {
		parsed, ok := decodeEvidenceBoolean(raw)
		if !ok {
			return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
		}
		return FieldValue{Bool: &parsed}, parsed, nil
	}
	var rawText string
	if json.Unmarshal(raw, &rawText) != nil {
		return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
	}
	var normalized string
	var ok bool
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		normalized, ok = fieldnorm.NormalizeNote(rawText)
	} else {
		normalized, ok = fieldnorm.NormalizeLine(rawText)
	}
	if !ok {
		return FieldValue{}, nil, newAdmissionFailure(fieldKey, admissionInvalidValue)
	}
	return FieldValue{Text: &normalized}, normalized, nil
}

func decodeEvidenceInteger(raw json.RawMessage) (int64, bool) {
	var number int64
	if json.Unmarshal(raw, &number) == nil {
		return number, true
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return 0, false
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	return parsed, err == nil
}

func decodeEvidenceBoolean(raw json.RawMessage) (bool, bool) {
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

func canonicalEvidenceValue(value FieldValue) any {
	switch {
	case value.Text != nil:
		return *value.Text
	case value.Timestamp != nil:
		return value.Timestamp.UTC().Format(time.RFC3339Nano)
	case value.UUID != nil:
		return value.UUID.String()
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}
