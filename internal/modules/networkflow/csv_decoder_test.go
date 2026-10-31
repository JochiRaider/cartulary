package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

// Called by the registered CSV selector; expected decoded cells are authored
// independently of the decoder and exercised at every input byte split.
func assertIncrementalCSVDialect(t *testing.T) {
	for _, test := range []struct {
		name, source string
		rows         [][]string
	}{
		{"record delimiters", "a,b\r1,2\r\n3,4\n5,6\n", [][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}}},
		{"blank records", "a\n\n\r\n\r", [][]string{{""}, {""}, {""}}},
		{"quotes and Unicode", "a,b\n\"界\r\n🙂\r\",\"a\"\"b\"\n x\"y,z\"z", [][]string{{"界\r\n🙂\r", "a\"b"}, {" x\"y", "z\"z"}}},
		{"BOM and terminal comma", "\ufeffa,b\n\"\",", [][]string{{"", ""}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for split := 0; split <= len(test.source); split++ {
				reader := io.MultiReader(strings.NewReader(test.source[:split]), strings.NewReader(test.source[split:]))
				parsed, err := parseCSVApply(iotest.OneByteReader(reader), fmt.Sprintf("%x", sha256.Sum256([]byte(test.source))), defaultLimits())
				if err != nil {
					t.Fatalf("split %d: %v", split, err)
				}
				var rows [][]string
				for i, record := range parsed.Records {
					if record.SourceRowNumber != int64(i+2) {
						t.Fatal("logical provenance changed")
					}
					rows = append(rows, record.Fields)
				}
				if !reflect.DeepEqual(rows, test.rows) {
					t.Fatalf("split %d rows=%#v want=%#v", split, rows, test.rows)
				}
			}
		})
	}
	for _, test := range []struct{ source, code, reason string }{
		{"", "network_flow_csv_empty_file", "zero_bytes"},
		{"\ufeff", "network_flow_csv_empty_file", "bom_only"},
		{"a\n", "network_flow_no_data_rows", "header_only"},
		{"a\x00\nx", "network_flow_invalid_header", "forbidden_header_control"},
		{"a\n\"x\"y", "network_flow_csv_malformed_quote", "invalid_after_closing_quote"},
		{"a\n\"x", "network_flow_csv_malformed_quote", "unclosed_quote"},
		{"a\nx\ufeff", "network_flow_invalid_utf8", "bom_not_at_offset_zero"},
		{"a\nx\xff", "network_flow_invalid_utf8", "invalid_utf8_sequence"},
	} {
		_, err := parseCSVApply(iotest.OneByteReader(strings.NewReader(test.source)), "", defaultLimits())
		var invalid *sourceValidationError
		if !errors.As(err, &invalid) || invalid.Code != test.code || invalid.ReasonCode != test.reason {
			t.Fatalf("source=%q err=%#v", test.source, err)
		}
	}
	t.Run("preview raw tail", func(t *testing.T) {
		prefix := "a\n" + strings.Repeat("ok\n", 50)
		for _, tail := range []string{"\"unclosed", "\xff", "\ufeff", "\"closed\"bad"} {
			source := prefix + tail
			parsed, err := parseCSVPreview(iotest.OneByteReader(strings.NewReader(source)), fmt.Sprintf("%x", sha256.Sum256([]byte(source))), defaultLimits())
			if err != nil || len(parsed.Records) != 50 {
				t.Fatalf("tail became preview failure: %v", err)
			}
			if _, err := parseCSVApply(strings.NewReader(source), "", defaultLimits()); err == nil {
				t.Fatal("apply ignored invalid tail")
			}
			if _, err := parseCSVPreview(strings.NewReader("a\n"+strings.Repeat("ok\n", 49)+tail), "", defaultLimits()); err == nil {
				t.Fatal("preview ignored record 50")
			}
		}
		failure := errors.New("tail read failed")
		if _, err := parseCSVPreview(io.MultiReader(strings.NewReader(prefix), iotest.ErrReader(failure)), "", defaultLimits()); !errors.Is(err, failure) {
			t.Fatalf("raw tail failure hidden: %v", err)
		}
		if _, err := parseCSVPreview(strings.NewReader(prefix), strings.Repeat("0", 64), defaultLimits()); !errors.Is(err, errSourceChanged) {
			t.Fatalf("tail digest ignored: %v", err)
		}
	})
	t.Run("bounded decoded fields and exact counts", func(t *testing.T) {
		limits := defaultLimits()
		limits.MaxRawCellScalarLength = 4
		for _, count := range []int{3, 4, 5, 65537} {
			value := strings.Repeat("界", count)
			parsed, err := parseCSVApply(strings.NewReader("a\n\""+value+"\""), "", limits)
			if err != nil {
				t.Fatal(err)
			}
			cell := parsed.Records[0].Cells[0]
			if cell.ScalarCount != int64(count) || cell.OverLimit != (count > 4) || cell.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(value))) || len(parsed.Records[0].Fields[0]) > 12 {
				t.Fatalf("cell accounting: %#v", cell)
			}
			if count > 4 && parsed.SourceColumns[0].SampleValues[0].SafeSample != nil {
				t.Fatal("oversized sample exposed")
			}
		}
		limits.MaxRowsPerCSV = 1
		_, err := parseCSVApply(strings.NewReader("a\nx\n\"unterminated"), "", limits)
		var invalid *sourceValidationError
		if !errors.As(err, &invalid) || invalid.ReasonCode != "unclosed_quote" {
			t.Fatalf("row limit hid malformed framing: %v", err)
		}
		limits.MaxHeaderScalarLength = 1
		_, err = parseCSVApply(strings.NewReader("aa\nx"), "", limits)
		if !errors.As(err, &invalid) || invalid.ReasonCode != "header_scalar_limit_exceeded" || invalid.ActualValue != 2 {
			t.Fatalf("header limit: %#v", err)
		}
		limits.MaxColumnsPerCSV = 1
		_, err = parseCSVApply(strings.NewReader("a,b\nx,y"), "", limits)
		if !errors.As(err, &invalid) || invalid.ReasonCode != "column_limit_exceeded" || invalid.ActualValue != 2 {
			t.Fatalf("column limit: %#v", err)
		}
	})
}

func assertRawCellDiagnostics(t *testing.T) {
	limits := defaultLimits()
	limits.MaxRawCellScalarLength = 24
	long := strings.Repeat("7", 25)
	header := strings.Join(requiredCiscoFields(), ",") + ",unmapped,ignored"
	row := "2026-07-13T12:00:00Z,2026-07-13T12:01:00Z,192.0.2.10,198.51.100.2," + long + ",51515,TCP,1,1," + long + "," + long + "\n"
	parsed, err := parseCSVPreview(strings.NewReader(header+"\n"+row), "", limits)
	if err != nil {
		t.Fatal(err)
	}
	mapping := approvedMappingFixture(sourceProfileCiscoSNANetFlowCSV)
	mapping.SourceColumns = parsed.SourceColumns
	fingerprint := mustMappingFingerprint(mapping, parsed.SourceContentSHA256)
	result, err := validatePreview(parsed, mapping, fingerprint, limits)
	if err != nil || len(result.Diagnostics) != 3 || result.RejectedRowCount != 1 {
		t.Fatalf("raw-cell diagnostics: %#v %v", result, err)
	}
	for i, d := range result.Diagnostics {
		if d.ReasonCode != "cell_scalar_limit_exceeded" || d.SafeSample != nil || *d.RawValueSHA256 != sha256Hex([]byte(long)) || *d.ActualValue != 25 || validateDiagnostic(d, parsed.SourceContentSHA256, fingerprint) != nil {
			t.Fatalf("unsafe/inexact diagnostic: %#v", d)
		}
		if (i == 0) != (d.FieldKey != nil) {
			t.Fatal("incorrect target attribution")
		}
	}
	for _, policy := range []string{unknownColumnPolicyPreserve, unknownColumnPolicyIgnore} {
		mapping.UnknownColumnPolicy = policy
		if policy == unknownColumnPolicyIgnore {
			for _, ordinal := range []int{10, 11} {
				mapping.FieldMappings = append(mapping.FieldMappings, materializeFieldMapping(fieldMapping{MappingKind: mappingKindIgnoredSourceColumn, SourceColumnOrdinal: ordinal}))
			}
		}
		if err := validateApprovedMapping(mapping); err != nil {
			t.Fatal(err)
		}
		result, err = validatePreview(parsed, mapping, mustMappingFingerprint(mapping, parsed.SourceContentSHA256), limits)
		if err != nil || len(result.Diagnostics) != 3 {
			t.Fatalf("unmapped policy bypassed raw limit: %s %v", policy, err)
		}
	}
	parsed, err = parseCSVPreview(strings.NewReader(header+"\n"+strings.TrimSpace(row)+",extra\n"), "", limits)
	if err != nil {
		t.Fatal(err)
	}
	result, err = validatePreview(parsed, mapping, mustMappingFingerprint(mapping, parsed.SourceContentSHA256), limits)
	if err != nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].ErrorCode != "network_flow_csv_field_count_mismatch" {
		t.Fatalf("field-count precedence: %#v %v", result, err)
	}
	if bytes.Contains(mustCanonicalJSON(result.Diagnostics), []byte(long)) {
		t.Fatal("raw oversized value leaked")
	}
	t.Run("complete preview beyond target-only bound", func(t *testing.T) {
		columns := append([]string{}, requiredCiscoFields()...)
		for i := 0; i < 30; i++ {
			columns = append(columns, fmt.Sprintf("extra%d", i))
		}
		data := strings.TrimSuffix(validPreviewRow, "\n") + strings.Repeat(","+long, 30) + "\n"
		source := strings.Join(columns, ",") + "\n" + strings.Repeat(data, 50)
		candidate := approvedMappingFixture(sourceProfileCiscoSNANetFlowCSV)
		var raw map[string]any
		if err := json.Unmarshal(marshalApprovedMapping(candidate), &raw); err != nil {
			t.Fatal(err)
		}
		delete(raw, "source_columns")
		facade := &importFacade{sourceStore: previewSource{csv: source}, limits: limits}
		preview, err := facade.PrepareImportUnitMapping(context.Background(), imports.ExtensionImportMappingRequest{
			TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID, OwnerMappingSchemaID: mappingCandidateSchemaID, OwnerMapping: canonicalTestValue(json.Marshal(raw)),
		})
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(preview.OwnerResult)
		var payload importPreviewPayload
		if err := json.Unmarshal(encoded, &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Diagnostics) != 1500 || payload.PreviewRejectedCount != 50 || payload.DiagnosticsTruncated {
			t.Fatalf("preview diagnostics lost: %d", len(payload.Diagnostics))
		}
		if err := facade.ValidateImportUnitMappingResult(preview); err != nil {
			t.Fatalf("complete preview rejected: %v", err)
		}
	})
}

// Full materialization is a test fixture convenience; production apply streams.
func parseCSVApply(reader io.Reader, expectedSHA256 string, limits EffectiveLimits) (parsedCSV, error) {
	d, err := newCSVDecoder(reader, expectedSHA256, limits, parseModeApply)
	if err != nil {
		return parsedCSV{}, err
	}
	result := parsedCSV{Records: []csvRecord{}}
	for {
		record, err := d.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return parsedCSV{}, err
		}
		result.Records = append(result.Records, record)
	}
	result.SourceColumns, result.SourceContentSHA256 = d.Columns, d.SourceSHA256
	return result, nil
}
