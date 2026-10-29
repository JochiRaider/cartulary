package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

var errInvalidPreview = errors.New("network flow preview failed owner validation")

// decodePreviewObject is private to the owner boundary. Required, optional and
// nullable members are explicit; Go zero values never stand in for presence.
func decodePreviewObject(raw json.RawMessage, required, optional, nullable string, target any) (map[string]json.RawMessage, error) {
	object, err := strictjson.DecodeObject(bytes.NewReader(raw))
	if err != nil {
		return nil, errInvalidPreview
	}
	for _, key := range strings.Fields(required) {
		if _, ok := object[key]; !ok {
			return nil, errInvalidPreview
		}
	}
	allowed := strings.Fields(required + " " + optional)
	for key, value := range object {
		if !slices.Contains(allowed, key) || (bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !slices.Contains(strings.Fields(nullable), key)) {
			return nil, errInvalidPreview
		}
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return nil, errInvalidPreview
	}
	return object, nil
}

type importPreviewPayload struct {
	SchemaID             string            `json:"schema_id"`
	SourceContentSHA256  string            `json:"source_content_sha256"`
	SourceColumns        json.RawMessage   `json:"source_columns"`
	MaterializedMapping  json.RawMessage   `json:"materialized_mapping"`
	MappingFingerprint   string            `json:"mapping_fingerprint"`
	PreviewRecordCount   int               `json:"preview_record_count"`
	PreviewAcceptedCount int               `json:"preview_accepted_count"`
	PreviewRejectedCount int               `json:"preview_rejected_count"`
	Diagnostics          []json.RawMessage `json:"diagnostics"`
	DiagnosticsTruncated bool              `json:"diagnostics_truncated"`
}

func (f *importFacade) ValidateImportUnitMappingResult(result imports.ExtensionImportMappingResult) error {
	if result.OwnerResultSchemaID != importPreviewResultSchemaID || result.OwnerResult == nil {
		return errInvalidPreview
	}
	raw, err := json.Marshal(result.OwnerResult)
	if err != nil {
		return errInvalidPreview
	}
	var payload importPreviewPayload
	if _, err := decodePreviewObject(raw, "schema_id source_content_sha256 source_columns materialized_mapping mapping_fingerprint preview_record_count preview_accepted_count preview_rejected_count diagnostics diagnostics_truncated", "", "", &payload); err != nil {
		return err
	}
	if payload.SchemaID != importPreviewResultSchemaID || !validLowerSHA256(payload.SourceContentSHA256) || !validLowerSHA256(payload.MappingFingerprint) || payload.MappingFingerprint != result.MappingFingerprint || payload.PreviewRecordCount < 0 || payload.PreviewRecordCount > previewRecordLimit || payload.PreviewAcceptedCount < 0 || payload.PreviewRejectedCount < 0 || payload.PreviewAcceptedCount+payload.PreviewRejectedCount != payload.PreviewRecordCount || payload.DiagnosticsTruncated {
		return errInvalidPreview
	}
	approved, err := decodePreviewMapping(payload.MaterializedMapping, f.limits, payload.PreviewRecordCount)
	if err != nil {
		return err
	}
	outer, err := decodePreviewMapping(result.OwnerMapping, f.limits, payload.PreviewRecordCount)
	if err != nil {
		return err
	}
	// Typed canonical values admit formatting and field-mapping order differences.
	canonicalize := func(mapping approvedMapping) approvedMapping {
		slices.SortFunc(mapping.FieldMappings, func(a, b fieldMapping) int { return bytes.Compare(canonicalJSON(a), canonicalJSON(b)) })
		return mapping
	}
	if !bytes.Equal(canonicalJSON(canonicalize(approved)), canonicalJSON(canonicalize(outer))) || mappingFingerprint(approved, payload.SourceContentSHA256) != payload.MappingFingerprint {
		return errInvalidPreview
	}
	columns, err := decodePreviewColumns(payload.SourceColumns, f.limits, payload.PreviewRecordCount)
	if err != nil || !bytes.Equal(canonicalJSON(columns), canonicalJSON(approved.SourceColumns)) {
		return errInvalidPreview
	}
	// At most one per target plus the cross-field check, per bounded record.
	if len(payload.Diagnostics) > payload.PreviewRecordCount*(len(requiredCiscoFields())+3) {
		return errInvalidPreview
	}
	rejected := map[int64]bool{}
	counts := map[int64]int{}
	fieldCountRows := map[int64]bool{}
	seenTargets := map[int64]map[string]bool{}
	var previous rejectedRowDiagnostic
	mappings := sourceFieldMappings(approved)
	for i, raw := range payload.Diagnostics {
		d, err := decodePreviewDiagnostic(raw)
		if err != nil || validateDiagnostic(d, payload.SourceContentSHA256, payload.MappingFingerprint) != nil {
			return errInvalidPreview
		}
		if d.SourceRowNumber < 2 || d.SourceRowNumber > int64(payload.PreviewRecordCount+1) || (i > 0 && compareDiagnostics(previous, d) >= 0) {
			return errInvalidPreview
		}
		previous = d
		rejected[d.SourceRowNumber] = true
		counts[d.SourceRowNumber]++
		if d.ErrorCode == "network_flow_csv_field_count_mismatch" {
			fieldCountRows[d.SourceRowNumber] = true
		}
		if d.SourceColumnOrdinal != nil {
			ordinal := *d.SourceColumnOrdinal
			if ordinal > int64(len(columns)) || *d.RawHeaderSHA256 != columns[ordinal-1].RawHeaderSHA256 {
				return errInvalidPreview
			}
			if d.FieldKey != nil {
				mapped, ok := mappings[*d.FieldKey]
				if !ok || int64(mapped.SourceColumnOrdinal) != ordinal {
					return errInvalidPreview
				}
				if seenTargets[d.SourceRowNumber] == nil {
					seenTargets[d.SourceRowNumber] = map[string]bool{}
				}
				if seenTargets[d.SourceRowNumber][*d.FieldKey] {
					return errInvalidPreview
				}
				seenTargets[d.SourceRowNumber][*d.FieldKey] = true
			}
		}
	}
	if len(rejected) != payload.PreviewRejectedCount {
		return errInvalidPreview
	}
	for row := range fieldCountRows {
		if counts[row] != 1 {
			return errInvalidPreview
		}
	}
	sampleCount := payload.PreviewRecordCount - len(fieldCountRows)
	for _, column := range columns {
		if len(column.SampleValues) != sampleCount {
			return errInvalidPreview
		}
	}
	return nil
}

func decodePreviewMapping(raw json.RawMessage, limits EffectiveLimits, recordCount int) (approvedMapping, error) {
	var mapping approvedMapping
	object, err := decodePreviewObject(raw, "target_kind target_table_schema_id source_profile_id parser_profile_id unknown_column_policy timestamp_profile source_columns field_mappings", "display_name_override", "", &mapping)
	if err != nil {
		return mapping, err
	}
	if validateTimestampProfileJSONShape(raw, true) != nil || validateApprovedMapping(mapping) != nil {
		return mapping, errInvalidPreview
	}
	if mapping.DisplayNameOverride != nil {
		normalized, err := normalizeExplicitDisplayName(*mapping.DisplayNameOverride)
		if err != nil || normalized != *mapping.DisplayNameOverride {
			return mapping, errInvalidPreview
		}
	}
	columns, err := decodePreviewColumns(object["source_columns"], limits, recordCount)
	if err != nil {
		return mapping, err
	}
	mapping.SourceColumns = columns
	var fields []json.RawMessage
	if json.Unmarshal(object["field_mappings"], &fields) != nil || len(fields) > len(columns)+1 {
		return mapping, errInvalidPreview
	}
	for i, field := range mapping.FieldMappings {
		var required string
		switch field.MappingKind {
		case mappingKindSourceColumn:
			required = "mapping_kind field_key source_column_ordinal transform_id empty_value_policy combinability"
		case mappingKindIgnoredSourceColumn:
			required = "mapping_kind source_column_ordinal ignore_reason"
		case mappingKindSystemDerivation:
			required = "mapping_kind field_key derivation_id combinability"
		default:
			return mapping, errInvalidPreview
		}
		if field.MappingKind == mappingKindSourceColumn && (field.Combinability != "single_source_only") {
			return mapping, errInvalidPreview
		}
		if field.MappingKind == mappingKindIgnoredSourceColumn && field.IgnoreReason != "user_ignored" && field.IgnoreReason != "policy_accounting" {
			return mapping, errInvalidPreview
		}
		var decoded fieldMapping
		if _, err := decodePreviewObject(fields[i], required, "", "", &decoded); err != nil {
			return mapping, err
		}
	}
	return mapping, nil
}

func decodePreviewColumns(raw json.RawMessage, limits EffectiveLimits, recordCount int) ([]sourceColumnDescriptor, error) {
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || len(values) < 1 || int64(len(values)) > limits.MaxColumnsPerCSV {
		return nil, errInvalidPreview
	}
	columns := make([]sourceColumnDescriptor, 0, len(values))
	for i, raw := range values {
		var column sourceColumnDescriptor
		object, err := decodePreviewObject(raw, "source_column_ordinal raw_header_text normalized_header_for_suggestion raw_header_sha256 sample_values detected_empty_count", "", "", &column)
		if err != nil || column.SourceColumnOrdinal != i+1 || invalidHeaderText(column.RawHeaderText, limits) || column.RawHeaderSHA256 != sha256Hex([]byte(column.RawHeaderText)) || column.NormalizedHeaderForSuggestion != sourceAliasMatchKey(column.RawHeaderText) || len(column.SampleValues) > recordCount || column.DetectedEmptyCount < 0 {
			return nil, errInvalidPreview
		}
		var samples []json.RawMessage
		if json.Unmarshal(object["sample_values"], &samples) != nil {
			return nil, errInvalidPreview
		}
		empty := 0
		for _, raw := range samples {
			var sample safeSample
			if _, err := decodePreviewObject(raw, "safe_sample raw_value_sha256", "", "safe_sample", &sample); err != nil || sample.RawValueSHA256 == nil || !validLowerSHA256(*sample.RawValueSHA256) {
				return nil, errInvalidPreview
			}
			if sample.SafeSample != nil && (!safeNumericSample(*sample.SafeSample) || sha256Hex([]byte(*sample.SafeSample)) != *sample.RawValueSHA256) {
				return nil, errInvalidPreview
			}
			if *sample.RawValueSHA256 == sha256Hex(nil) {
				empty++
			}
		}
		if column.DetectedEmptyCount != empty {
			return nil, errInvalidPreview
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func decodePreviewDiagnostic(raw json.RawMessage) (rejectedRowDiagnostic, error) {
	var d rejectedRowDiagnostic
	_, err := decodePreviewObject(raw, "diagnostic_id source_row_number source_column_ordinal raw_header_sha256 field_key error_code reason_code safe_sample raw_value_sha256 message_key message_args message limit_name limit_value actual_value", "", "source_column_ordinal raw_header_sha256 field_key safe_sample raw_value_sha256 limit_name limit_value actual_value", &d)
	return d, err
}

func validatePreviewDiagnosticShape(raw json.RawMessage) error {
	_, err := decodePreviewDiagnostic(raw)
	return err
}
