package networkflow

import "encoding/json"

// Descriptor admission is shared by source inspection, preview and stored
// approvals. Only the first two know the exact preview record count.
func decodeSourceColumns(raw json.RawMessage, limits EffectiveLimits) ([]sourceColumnDescriptor, error) {
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || len(values) < 1 || int64(len(values)) > limits.MaxColumnsPerCSV {
		return nil, errInvalidImportObject
	}
	columns := make([]sourceColumnDescriptor, 0, len(values))
	sampleCount := 0
	for i, raw := range values {
		var wire struct {
			sourceColumnDescriptor
			SampleValues json.RawMessage `json:"sample_values"`
		}
		if _, err := decodeClosedImportObject(raw, "source_column_ordinal raw_header_text normalized_header_for_suggestion raw_header_sha256 sample_values detected_empty_count", "", "", &wire); err != nil {
			return nil, errInvalidImportObject
		}
		column := wire.sourceColumnDescriptor
		if column.SourceColumnOrdinal != i+1 || invalidHeaderText(column.RawHeaderText, limits) || column.RawHeaderSHA256 != sha256Hex([]byte(column.RawHeaderText)) || column.NormalizedHeaderForSuggestion != sourceAliasMatchKey(column.RawHeaderText) || column.DetectedEmptyCount < 0 {
			return nil, errInvalidImportObject
		}
		var samples []json.RawMessage
		if json.Unmarshal(wire.SampleValues, &samples) != nil || samples == nil || len(samples) > previewRecordLimit {
			return nil, errInvalidImportObject
		}
		// Every column samples the same structurally valid preview records.
		if i == 0 {
			sampleCount = len(samples)
		} else if len(samples) != sampleCount {
			return nil, errInvalidImportObject
		}
		column.SampleValues = make([]safeSample, 0, len(samples))
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
			column.SampleValues = append(column.SampleValues, sample)
		}
		if column.DetectedEmptyCount != empty {
			return nil, errInvalidImportObject
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func validateSourceColumnSampleCount(columns []sourceColumnDescriptor, recordCount int) error {
	for _, column := range columns {
		if len(column.SampleValues) > recordCount {
			return errInvalidImportObject
		}
	}
	return nil
}

func decodePreviewColumns(raw json.RawMessage, limits EffectiveLimits, recordCount int) ([]sourceColumnDescriptor, error) {
	columns, err := decodeSourceColumns(raw, limits)
	if err != nil {
		return nil, err
	}
	if err := validateSourceColumnSampleCount(columns, recordCount); err != nil {
		return nil, err
	}
	return columns, nil
}
