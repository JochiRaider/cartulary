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
	mapping, err := decodeApprovedMapping(raw, limits)
	if err != nil || validateSourceColumnSampleCount(mapping.SourceColumns, recordCount) != nil {
		return approvedMapping{}, errInvalidImportObject
	}
	return mapping, nil
}
