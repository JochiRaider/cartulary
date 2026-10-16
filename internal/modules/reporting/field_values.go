package reporting

import (
	"encoding/json"
	"unicode/utf8"

	strictjson "github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// admitReportingFieldValue applies the common field-value contract before
// materialization or rendering. It never repairs source data or converts a
// container to text. Canonical admission preserves exact integer tokens.
func admitReportingFieldValue(value any, multiline bool) (any, bool) {
	raw, err := canonicalJSON(value)
	if err != nil {
		return nil, false
	}
	value, err = strictjson.DecodeStrict(raw)
	if err != nil {
		return nil, false
	}
	if values, ok := value.([]any); ok {
		for _, item := range values {
			if !validReportingFieldScalar(item, multiline) {
				return nil, false
			}
		}
		return value, true
	}
	return value, validReportingFieldScalar(value, multiline)
}

func validReportingFieldScalar(value any, multiline bool) bool {
	switch value := value.(type) {
	case nil, bool, json.Number:
		return true
	case string:
		return validReportingText(value, multiline)
	default:
		return false
	}
}

func validReportingText(value string, multiline bool) bool {
	if !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if r == '\n' && multiline {
			continue
		}
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			return false
		}
	}
	return true
}

// Frozen records and events duplicate source fields for deterministic lookup.
// All copies must satisfy the same scalar boundary at admission, including
// recovery and replay paths that never invoke a renderer.
func validateSnapshotFieldValues(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	validate := func(fields []ReportingField, multiline bool) {
		keys := make(map[string]bool, len(fields))
		for index, field := range fields {
			if keys[field.FieldKey] {
				issues.add(newRenderValidationError("export_model_invalid", "duplicate_stable_id"))
			}
			keys[field.FieldKey] = true
			if !validReportingField(field, index+1, multiline) {
				issues.add(newRenderValidationError("export_model_invalid", "invalid_scalar_value"))
			}
		}
	}
	var blocks func([]ReportingBlock)
	blocks = func(items []ReportingBlock) {
		for _, block := range items {
			validate(block.Fields, block.BlockKind == "paragraph")
			blocks(block.Children)
		}
	}
	for _, section := range model.Sections {
		blocks(section.Blocks)
	}
	for _, record := range model.Records {
		validate(record.Fields, true)
	}
	for _, event := range model.TimelineEvents {
		validate(event.Fields, true)
	}
	return issues.finish()
}

func validReportingField(field ReportingField, ordinal int, multiline bool) bool {
	value, valid := admitReportingFieldValue(field.Value, multiline)
	if !valid || field.SchemaID != "cartulary.reporting_field.v1" || !reportingMemberName.MatchString(field.FieldKey) ||
		field.FieldOrdinal != ordinal || field.RawValueSHA256 != nil ||
		(field.DisplayLabel != nil && !validReportingText(*field.DisplayLabel, false)) ||
		!orderedReportingRefs(field.SourceRefs) || !orderedReportingRefs(field.SupportRefs) ||
		field.DisclosurePartitionRefs == nil || validateDisclosurePartitionSet(field.DisclosurePartitionRefs) != nil {
		return false
	}
	unchanged := field.RedactedValueState == "unchanged"
	allowed := field.RedactedValueState == "allowed"
	masked := field.RedactedValueState == "masked" || field.RedactedValueState == "stubbed" || field.RedactedValueState == "tokenized"
	presentAction := unchanged || allowed || masked || field.RedactedValueState == "truncated"
	switch field.SourceValueState {
	case "present", "superseded":
		return presentAction && value != nil
	case "null":
		return (unchanged || allowed) && value == nil
	case "missing":
		return unchanged && value == nil
	case "unavailable":
		return (unchanged && value == nil) || (field.RedactedValueState == "stubbed" && value != nil)
	case "withheld":
		return masked && value != nil
	case "derived":
		return presentAction
	default:
		return false
	}
}

func orderedReportingRefs(refs []string) bool {
	if refs == nil {
		return false
	}
	for index, ref := range refs {
		if ref == "" || !validReportingText(ref, false) || (index > 0 && refs[index-1] >= ref) {
			return false
		}
	}
	return true
}
