package networkflow

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

type previewSource struct {
	ImportSourcePort
	csv string
}

func (s previewSource) OpenSourceStream(context.Context, string) (imports.ImportSourceStream, error) {
	return imports.ImportSourceStream{Reader: io.NopCloser(strings.NewReader(s.csv))}, nil
}

func preparePreviewFixture(t *testing.T, data string, limits EffectiveLimits, changes ...func(map[string]any)) (*importFacade, imports.ExtensionImportMappingResult) {
	t.Helper()
	mapping := approvedMappingFixture(sourceProfileCiscoSNANetFlowCSV)
	var candidate map[string]any
	if err := json.Unmarshal(marshalApprovedMapping(mapping), &candidate); err != nil {
		t.Fatal(err)
	}
	delete(candidate, "source_columns")
	for _, change := range changes {
		change(candidate)
	}
	raw, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	facade := &importFacade{sourceStore: previewSource{csv: strings.Join(requiredCiscoFields(), ",") + "\n" + data}, limits: limits}
	result, err := facade.PrepareImportUnitMapping(context.Background(), imports.ExtensionImportMappingRequest{
		TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID,
		OwnerMappingSchemaID: mappingCandidateSchemaID, OwnerMapping: raw,
	})
	if err != nil {
		t.Fatal(err)
	}
	return facade, result
}

const validPreviewRow = "2026-07-13T12:00:00Z,2026-07-13T12:01:00Z,192.0.2.10,198.51.100.2,443,51515,TCP,1,1\n"

func TestImportPreviewAdmission(t *testing.T) {
	t.Run("unit resource failures cannot be row diagnostics", func(t *testing.T) {
		limits := defaultLimits()
		limits.MaxRawCellScalarLength = 24
		facade, result := preparePreviewFixture(t, strings.Replace(validPreviewRow, "443", strings.Repeat("7", 25), 1), limits)
		if err := facade.ValidateImportUnitMappingResult(result); err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(result.OwnerResult["diagnostics"])
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics []json.RawMessage
		if err := json.Unmarshal(raw, &diagnostics); err != nil || len(diagnostics) != 1 {
			t.Fatalf("diagnostic fixture: %s, %v", raw, err)
		}
		valid, err := decodeDiagnosticResource(diagnostics[0])
		if err != nil {
			t.Fatal(err)
		}
		source := result.OwnerResult["source_content_sha256"].(string)
		if err := validateDiagnostic(valid, source, result.MappingFingerprint); err != nil {
			t.Fatal(err)
		}
		for reason, limit := range map[string]string{
			"header_scalar_limit_exceeded": "network_flow.max_header_scalar_length",
			"column_limit_exceeded":        "network_flow.max_columns_per_csv",
			"row_limit_exceeded":           "network_flow.max_rows_per_csv",
			"accepted_row_limit_exceeded":  "network_flow.max_accepted_rows_per_table",
		} {
			t.Run(reason, func(t *testing.T) {
				forged := valid
				forged.ReasonCode = reason
				forged.LimitName = &limit
				forged.MessageKey = "network_flow.diagnostic." + forged.ErrorCode + "." + reason
				forged.Message = forged.MessageKey
				forged.DiagnosticID = diagnosticID(source, result.MappingFingerprint, forged.SourceRowNumber, forged.SourceColumnOrdinal, forged.FieldKey, forged.ErrorCode, reason)
				if err := validateDiagnostic(forged, source, result.MappingFingerprint); err == nil {
					t.Fatal("unit failure admitted as a row diagnostic")
				}
				result.OwnerResult["diagnostics"] = []map[string]any{diagnosticResource(forged)}
				if err := facade.ValidateImportUnitMappingResult(result); err == nil {
					t.Fatal("public owner-result boundary admitted unit failure")
				}
			})
		}
	})
	t.Run("incremental owner dialect", assertIncrementalCSVDialect)
	t.Run("raw cell diagnostics", assertRawCellDiagnostics)
	t.Run("fragmented source preserves identity and preview agreement", func(t *testing.T) {
		source := strings.Join(requiredCiscoFields(), ",") + "\n" + validPreviewRow
		digest := sha256Hex([]byte(source))
		want, err := parseCSVPreview(strings.NewReader(source), digest, defaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		for _, reader := range []io.Reader{
			iotest.OneByteReader(strings.NewReader(source)),
			iotest.HalfReader(strings.NewReader(source)),
			iotest.DataErrReader(strings.NewReader(source)),
		} {
			got, err := parseCSVPreview(reader, digest, defaultLimits())
			if err != nil || string(canonicalTestValue(json.Marshal(got))) != string(canonicalTestValue(json.Marshal(want))) {
				t.Fatalf("fragmented source changed preview: %v", err)
			}
		}
		if _, err := parseCSVPreview(strings.NewReader(source), strings.Repeat("0", 64), defaultLimits()); err == nil {
			t.Fatal("source digest mismatch admitted")
		}
		if _, err := parseCSVPreview(io.MultiReader(strings.NewReader(source), iotest.ErrReader(io.ErrUnexpectedEOF)), digest, defaultLimits()); err == nil {
			t.Fatal("source read failure admitted")
		}
	})
	t.Run("zero record result satisfies the declared shape", func(t *testing.T) {
		facade, result := preparePreviewFixture(t, validPreviewRow, defaultLimits())
		mapping, err := decodeApprovedMapping(result.OwnerMapping)
		if err != nil {
			t.Fatal(err)
		}
		for i := range mapping.SourceColumns {
			mapping.SourceColumns[i].SampleValues = []safeSample{}
			mapping.SourceColumns[i].DetectedEmptyCount = 0
		}
		sourceDigest := sha256Hex([]byte(strings.Join(requiredCiscoFields(), ",") + "\n"))
		result.OwnerMapping = marshalApprovedMapping(mapping)
		result.MappingFingerprint = mustMappingFingerprint(mapping, sourceDigest)
		result.OwnerResult["source_content_sha256"] = sourceDigest
		result.OwnerResult["mapping_fingerprint"] = result.MappingFingerprint
		result.OwnerResult["materialized_mapping"] = mapping
		result.OwnerResult["source_columns"] = mapping.SourceColumns
		result.OwnerResult["preview_record_count"] = 0
		result.OwnerResult["preview_accepted_count"] = 0
		if err := facade.ValidateImportUnitMappingResult(result); err != nil {
			t.Fatal(err)
		}
		result.OwnerResult["preview_record_count"] = -1
		if err := facade.ValidateImportUnitMappingResult(result); err == nil {
			t.Fatal("negative record count admitted")
		}
	})
	t.Run("diagnostic identity and privacy", func(t *testing.T) {
		column := int64(4)
		field := fieldSrcPort
		got := diagnosticID(strings.Repeat("1", 64), strings.Repeat("2", 64), 3, &column, &field, "network_flow_invalid_port", "out_of_range")
		const want = "nfd_d00a0f3a8ca225d1715c1c96c00d530a6d80998258ffbbb4d552596c147676ff"
		if got != want {
			t.Fatalf("diagnostic identity = %s, want %s", got, want)
		}
		if got := diagnosticID(strings.Repeat("1", 64), strings.Repeat("2", 64), 3, nil, nil, "network_flow_csv_field_count_mismatch", "field_count_mismatch"); got != "nfd_0ae3cdcbefd521cc610900df4cbe180766603775f321c051163a531be1c19de4" {
			t.Fatalf("nullable diagnostic vector: %s", got)
		}

		for _, raw := range []string{"192.0.2.1", "secret-host", "", "=1+1"} {
			d := diagnostic(3, &column, stringPtr(strings.Repeat("a", 64)), stringPtr(fieldSrcIP), "network_flow_invalid_ip", "invalid_syntax", raw)
			d.DiagnosticID = diagnosticID(strings.Repeat("1", 64), strings.Repeat("2", 64), d.SourceRowNumber, d.SourceColumnOrdinal, d.FieldKey, d.ErrorCode, d.ReasonCode)
			if d.SafeSample != nil || d.RawValueSHA256 == nil || d.Message != d.MessageKey {
				t.Fatal("unsafe diagnostic")
			}
			if err := validateDiagnostic(d, strings.Repeat("1", 64), strings.Repeat("2", 64)); err != nil {
				t.Fatal(err)
			}
		}
		mapping := approvedMappingFixture(sourceProfileCiscoSNANetFlowCSV)
		record := csvRecord{SourceRowNumber: 2, Fields: []string{strings.Repeat("x", 257)}}
		_, textDiagnostic := mappedValue(record, mapping, fieldMapping{SourceColumnOrdinal: 1, TransformID: transformTrimASCIISpace}, fieldInputInterface)
		if textDiagnostic == nil || textDiagnostic.ErrorCode != "network_flow_invalid_text" || textDiagnostic.SafeSample != nil || !validDiagnosticReason(textDiagnostic.ErrorCode, textDiagnostic.ReasonCode) {
			t.Fatal("invalid optional text lost its owner diagnostic")
		}
		d := fieldCountDiagnostic(csvRecord{SourceRowNumber: 2, RawFieldCount: 1}, 9)
		if d.SafeSample != nil || d.RawValueSHA256 != nil {
			t.Fatal("row-level diagnostic exposed fabricated scalar")
		}
	})
	t.Run("complete preview diagnostics independent of apply retention", func(t *testing.T) {
		invalid := strings.ReplaceAll(strings.ReplaceAll(validPreviewRow, "443", "99999"), "192.0.2.10", "secret-host")
		data := invalid + "short,row\n" + validPreviewRow
		for _, limit := range []int64{0, 2, 3, 4} {
			limits := defaultLimits()
			limits.MaxRejectedRowDiagnostics = limit

			facade, result := preparePreviewFixture(t, data, limits)
			if err := facade.ValidateImportUnitMappingResult(result); err != nil {
				t.Fatal(err)
			}
			if result.OwnerResult["diagnostics_truncated"] != false || result.OwnerResult["preview_rejected_count"] != int64(2) || result.OwnerResult["preview_accepted_count"] != int64(1) {
				t.Fatal("preview accounting", result.OwnerResult)
			}
			preview := result.OwnerResult["diagnostics"].([]map[string]any)
			if len(preview) != 3 {
				t.Fatalf("preview omitted diagnostics at limit %d", limit)
			}

		}
		for _, count := range []int{50, 51} {
			_, result := preparePreviewFixture(t, strings.Repeat(invalid, count), defaultLimits())
			if result.OwnerResult["preview_record_count"] != 50 || len(result.OwnerResult["diagnostics"].([]map[string]any)) != 100 {
				t.Fatal("preview record bound")
			}
		}
	})
	t.Run("strict result and nested contracts", func(t *testing.T) {
		malformed := strings.Replace(validPreviewRow, "192.0.2.10", "private-host", 1)
		facade, good := preparePreviewFixture(t, malformed, defaultLimits())
		if err := facade.ValidateImportUnitMappingResult(good); err != nil {
			t.Fatal(err)
		}
		clone := func() imports.ExtensionImportMappingResult {
			_, result := preparePreviewFixture(t, malformed, defaultLimits())
			raw, _ := json.Marshal(result.OwnerResult)
			if err := json.Unmarshal(raw, &result.OwnerResult); err != nil {
				t.Fatal(err)
			}
			return result
		}
		reject := func(result imports.ExtensionImportMappingResult) {
			t.Helper()
			if err := facade.ValidateImportUnitMappingResult(result); err == nil {
				t.Fatal("corrupt result admitted")
			}
		}
		for key := range good.OwnerResult {
			t.Run("missing "+key, func(t *testing.T) { bad := clone(); delete(bad.OwnerResult, key); reject(bad) })
			t.Run("null "+key, func(t *testing.T) { bad := clone(); bad.OwnerResult[key] = nil; reject(bad) })
		}
		for _, target := range []string{"column", "sample", "diagnostic", "mapping", "field"} {
			selectObject := func(result imports.ExtensionImportMappingResult) map[string]any {
				switch target {
				case "column":
					return result.OwnerResult["source_columns"].([]any)[0].(map[string]any)
				case "sample":
					return result.OwnerResult["source_columns"].([]any)[0].(map[string]any)["sample_values"].([]any)[0].(map[string]any)
				case "diagnostic":
					return result.OwnerResult["diagnostics"].([]any)[0].(map[string]any)
				case "mapping":
					return result.OwnerResult["materialized_mapping"].(map[string]any)
				default:
					return result.OwnerResult["materialized_mapping"].(map[string]any)["field_mappings"].([]any)[0].(map[string]any)
				}
			}
			for key := range selectObject(clone()) {
				t.Run(target+" missing "+key, func(t *testing.T) { bad := clone(); delete(selectObject(bad), key); reject(bad) })
			}
			bad := clone()
			selectObject(bad)["unexpected"] = "private-source"
			reject(bad)
		}
		for key, value := range map[string]any{"source_row_number": 1, "source_column_ordinal": nil, "diagnostic_id": "nfd_" + strings.Repeat("a", 64), "message": "private-host", "message_args": map[string]any{"secret": "private-host"}, "reason_code": "unknown", "safe_sample": "private-host", "raw_value_sha256": nil, "limit_name": "secret"} {
			t.Run("diagnostic value "+key, func(t *testing.T) {
				bad := clone()
				bad.OwnerResult["diagnostics"].([]any)[0].(map[string]any)[key] = value
				reject(bad)
			})
		}
		bad := clone()
		bad.OwnerMapping = json.RawMessage(`{}`)
		reject(bad)
		bad = clone()
		bad.OwnerMapping = append(bad.OwnerMapping, []byte(` {}`)...)
		reject(bad)
		bad = clone()
		bad.OwnerResult["diagnostics"] = []any{}
		reject(bad)
		bad = clone()
		bad.OwnerResult["preview_rejected_count"] = nil
		reject(bad)
		// Equivalent formatting is admitted after typed canonicalization.
		formatted := clone()
		var buf strings.Builder
		raw, _ := json.MarshalIndent(json.RawMessage(formatted.OwnerMapping), "", "  ")
		buf.Write(raw)
		formatted.OwnerMapping = json.RawMessage(buf.String())
		if err := facade.ValidateImportUnitMappingResult(formatted); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("normalized display override roundtrip", func(t *testing.T) {
		facade, result := preparePreviewFixture(t, validPreviewRow, defaultLimits(), func(mapping map[string]any) { mapping["display_name_override"] = "  Named table  " })
		if err := facade.ValidateImportUnitMappingResult(result); err != nil {
			t.Fatal(err)
		}
		mapping, err := decodeApprovedMapping(result.OwnerMapping)
		if err != nil || mapping.DisplayNameOverride == nil || *mapping.DisplayNameOverride != "Named table" {
			t.Fatal("display name was not materialized")
		}
	})

	t.Run("null policy on a required scalar rejects the row", func(t *testing.T) {
		facade, result := preparePreviewFixture(t, strings.Replace(validPreviewRow, "2026-07-13T12:00:00Z", "", 1), defaultLimits(), func(mapping map[string]any) {
			for _, row := range mapping["field_mappings"].([]any) {
				field := row.(map[string]any)
				if field["field_key"] == fieldFlowStartUTC {
					field["empty_value_policy"] = emptyPolicyNull
				}
			}
		})
		if err := facade.ValidateImportUnitMappingResult(result); err != nil {
			t.Fatal(err)
		}
		if result.OwnerResult["preview_rejected_count"] != int64(1) {
			t.Fatal("non-nullable empty scalar admitted")
		}
	})

	facade, result := preparePreviewFixture(t, validPreviewRow, defaultLimits())
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		t.Fatalf("prepared preview did not roundtrip: %v", err)
	}
	for _, field := range []string{"schema_id", "mapping_fingerprint", "source_content_sha256", "preview_record_count", "preview_accepted_count", "materialized_mapping", "source_columns"} {
		t.Run(field, func(t *testing.T) {
			_, invalid := preparePreviewFixture(t, validPreviewRow, defaultLimits())
			invalid.OwnerResult[field] = "invalid"
			if err := facade.ValidateImportUnitMappingResult(invalid); err == nil {
				t.Fatal("corrupted result accepted")
			}
		})
	}
	for _, raw := range []string{`null`, `[]`, `{`, `{} {}`, `{"x":1,"x":2}`, `{"x":{"y":1,"y":2}}`, `{}`} {
		if _, err := decodeDiagnosticResource(json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid diagnostic admitted: %s", raw)
		}
	}
}
