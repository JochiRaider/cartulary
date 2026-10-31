package networkflow

import (
	"encoding/json"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

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
		return errInvalidImportObject
	}
	raw, err := json.Marshal(result.OwnerResult)
	if err != nil {
		return errInvalidImportObject
	}
	var payload importPreviewPayload
	if _, err := decodeClosedImportObject(raw, "schema_id source_content_sha256 source_columns materialized_mapping mapping_fingerprint preview_record_count preview_accepted_count preview_rejected_count diagnostics diagnostics_truncated", "", "", &payload); err != nil {
		return err
	}
	if payload.SchemaID != importPreviewResultSchemaID || !validLowerSHA256(payload.SourceContentSHA256) || !validLowerSHA256(payload.MappingFingerprint) || payload.MappingFingerprint != result.MappingFingerprint || payload.PreviewRecordCount < 0 || payload.PreviewRecordCount > previewRecordLimit || payload.PreviewAcceptedCount < 0 || payload.PreviewRejectedCount < 0 || payload.PreviewAcceptedCount+payload.PreviewRejectedCount != payload.PreviewRecordCount || payload.DiagnosticsTruncated {
		return errInvalidImportObject
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
		slices.SortFunc(mapping.FieldMappings, compareFieldMappings)
		return mapping
	}
	fingerprint, err := mappingFingerprint(approved, payload.SourceContentSHA256)
	if err != nil {
		return err
	}
	if !canonicalEqual(canonicalize(approved), canonicalize(outer)) || fingerprint != payload.MappingFingerprint {
		return errInvalidImportObject
	}
	columns, err := decodePreviewColumns(payload.SourceColumns, f.limits, payload.PreviewRecordCount)
	if result.HeaderRowRef != 1 || result.DataStartRowRef != 2 || !slices.Equal(result.SourceColumns, importSourceColumns(columns)) {
		return errInvalidImportObject
	}
	if err != nil || !canonicalEqual(columns, approved.SourceColumns) {
		return errInvalidImportObject
	}
	// Raw-cell diagnostics include unmapped and ignored columns as well as targets.
	if len(payload.Diagnostics) > payload.PreviewRecordCount*(int(f.limits.MaxColumnsPerCSV)+len(approved.FieldMappings)+1) {
		return errInvalidImportObject
	}
	rejected := map[int64]bool{}
	counts := map[int64]int{}
	fieldCountRows := map[int64]bool{}
	seenTargets := map[int64]map[string]bool{}
	var previous rejectedRowDiagnostic
	mappings := sourceFieldMappings(approved)
	for i, raw := range payload.Diagnostics {
		d, err := decodeDiagnosticResource(raw)
		if err != nil || validateDiagnostic(d, payload.SourceContentSHA256, payload.MappingFingerprint) != nil {
			return errInvalidImportObject
		}
		if d.SourceRowNumber < 2 || d.SourceRowNumber > int64(payload.PreviewRecordCount+1) || (i > 0 && compareDiagnostics(previous, d) >= 0) {
			return errInvalidImportObject
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
				return errInvalidImportObject
			}
			if d.ReasonCode == "cell_scalar_limit_exceeded" {
				if d.LimitValue == nil || *d.LimitValue != f.limits.MaxRawCellScalarLength {
					return errInvalidImportObject
				}
				for key, mapped := range mappings {
					if int64(mapped.SourceColumnOrdinal) == ordinal && (d.FieldKey == nil || *d.FieldKey != key) {
						return errInvalidImportObject
					}
				}
			}
			if d.FieldKey != nil {
				mapped, ok := mappings[*d.FieldKey]
				if !ok || int64(mapped.SourceColumnOrdinal) != ordinal {
					return errInvalidImportObject
				}
				if seenTargets[d.SourceRowNumber] == nil {
					seenTargets[d.SourceRowNumber] = map[string]bool{}
				}
				if seenTargets[d.SourceRowNumber][*d.FieldKey] {
					return errInvalidImportObject
				}
				seenTargets[d.SourceRowNumber][*d.FieldKey] = true
			}
		}
	}
	if len(rejected) != payload.PreviewRejectedCount {
		return errInvalidImportObject
	}
	for row := range fieldCountRows {
		if counts[row] != 1 {
			return errInvalidImportObject
		}
	}
	sampleCount := payload.PreviewRecordCount - len(fieldCountRows)
	for _, column := range columns {
		if len(column.SampleValues) != sampleCount {
			return errInvalidImportObject
		}
	}
	return nil
}

func decodePreviewMapping(raw json.RawMessage, limits EffectiveLimits, recordCount int) (approvedMapping, error) {
	var mapping approvedMapping
	object, err := decodeClosedImportObject(raw, "target_kind target_table_schema_id source_profile_id parser_profile_id unknown_column_policy timestamp_profile source_columns field_mappings", "display_name_override", "", &mapping)
	if err != nil {
		return mapping, err
	}
	if validateTimestampProfileJSONShape(raw, true) != nil || validateApprovedMapping(mapping) != nil {
		return mapping, errInvalidImportObject
	}
	if mapping.DisplayNameOverride != nil {
		normalized, err := normalizeExplicitDisplayName(*mapping.DisplayNameOverride)
		if err != nil || normalized != *mapping.DisplayNameOverride {
			return mapping, errInvalidImportObject
		}
	}
	columns, err := decodePreviewColumns(object["source_columns"], limits, recordCount)
	if err != nil {
		return mapping, err
	}
	mapping.SourceColumns = columns
	var fields []json.RawMessage
	if json.Unmarshal(object["field_mappings"], &fields) != nil || len(fields) > len(columns)+1 {
		return mapping, errInvalidImportObject
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
			return mapping, errInvalidImportObject
		}
		if field.MappingKind == mappingKindSourceColumn && (field.Combinability != "single_source_only") {
			return mapping, errInvalidImportObject
		}
		if field.MappingKind == mappingKindIgnoredSourceColumn && field.IgnoreReason != "user_ignored" && field.IgnoreReason != "policy_accounting" {
			return mapping, errInvalidImportObject
		}
		var decoded fieldMapping
		if _, err := decodeClosedImportObject(fields[i], required, "", "", &decoded); err != nil {
			return mapping, err
		}
	}
	return mapping, nil
}

func decodePreviewColumns(raw json.RawMessage, limits EffectiveLimits, recordCount int) ([]sourceColumnDescriptor, error) {
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || len(values) < 1 || int64(len(values)) > limits.MaxColumnsPerCSV {
		return nil, errInvalidImportObject
	}
	columns := make([]sourceColumnDescriptor, 0, len(values))
	for i, raw := range values {
		var column sourceColumnDescriptor
		object, err := decodeClosedImportObject(raw, "source_column_ordinal raw_header_text normalized_header_for_suggestion raw_header_sha256 sample_values detected_empty_count", "", "", &column)
		if err != nil || column.SourceColumnOrdinal != i+1 || invalidHeaderText(column.RawHeaderText, limits) || column.RawHeaderSHA256 != sha256Hex([]byte(column.RawHeaderText)) || column.NormalizedHeaderForSuggestion != sourceAliasMatchKey(column.RawHeaderText) || len(column.SampleValues) > recordCount || column.DetectedEmptyCount < 0 {
			return nil, errInvalidImportObject
		}
		var samples []json.RawMessage
		if json.Unmarshal(object["sample_values"], &samples) != nil {
			return nil, errInvalidImportObject
		}
		empty := 0
		for _, raw := range samples {
			var sample safeSample
			if _, err := decodeClosedImportObject(raw, "safe_sample raw_value_sha256", "", "safe_sample", &sample); err != nil || sample.RawValueSHA256 == nil || !validLowerSHA256(*sample.RawValueSHA256) {
				return nil, errInvalidImportObject
			}
			if sample.SafeSample != nil && (!safeNumericSample(*sample.SafeSample) || sha256Hex([]byte(*sample.SafeSample)) != *sample.RawValueSHA256) {
				return nil, errInvalidImportObject
			}
			if *sample.RawValueSHA256 == sha256Hex(nil) {
				empty++
			}
		}
		if column.DetectedEmptyCount != empty {
			return nil, errInvalidImportObject
		}
		columns = append(columns, column)
	}
	return columns, nil
}
