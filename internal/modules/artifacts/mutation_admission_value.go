package artifacts

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func decodeArtifactValue(fieldKey string, field viewschema.Field, raw json.RawMessage, patch bool) (fieldValue, any, *AdmissionError) {
	if string(raw) == "null" {
		if !patch || field.Clearable {
			return fieldValue{}, nil, nil
		}
		return fieldValue{}, nil, newAdmissionError(fieldKey, admissionFieldNotNullable)
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		normalized, ok := fieldnorm.NormalizeTimestampInstant(text)
		if !ok {
			return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		return fieldValue{Timestamp: &normalized}, normalized.Format(time.RFC3339Nano), nil
	}
	if field.DirectReferenceContractID != nil || strings.HasSuffix(fieldKey, "_user_id") {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		parsed, err := uuid.Parse(strings.TrimSpace(text))
		if err != nil || (field.DirectReferenceContractID != nil && parsed.String() != text) {
			return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		return fieldValue{UUID: &parsed}, parsed.String(), nil
	}
	if field.ReadKind == "number" {
		parsed, ok := decodeArtifactInteger(raw)
		if !ok {
			return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		return fieldValue{Number: &parsed}, parsed, nil
	}
	if field.ReadKind == "boolean" {
		parsed, ok := decodeArtifactBoolean(raw)
		if !ok {
			return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		return fieldValue{Bool: &parsed}, parsed, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	var normalized string
	var ok bool
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		normalized, ok = fieldnorm.NormalizeNote(text)
	} else {
		normalized, ok = fieldnorm.NormalizeLine(text)
	}
	if !ok {
		return fieldValue{}, nil, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	return fieldValue{Text: &normalized}, normalized, nil
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
