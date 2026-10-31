package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/networkflowtz"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

type fixtureMappingCase struct {
	Name    string          `json:"name"`
	Source  string          `json:"source"`
	Mapping json.RawMessage `json:"mapping"`
}

func TestNetworkFlowAliasFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-011-alias-collision", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		outcomes, err := mappingFixtureOutcomes(bundle)
		if err != nil {
			return nil, err
		}
		data, err := bundle.Read("source/alias-keys.json")
		if err != nil {
			return nil, err
		}
		var inputs []string
		if err := json.Unmarshal(data, &inputs); err != nil {
			return nil, err
		}
		for index, input := range inputs {
			outcomes = append(outcomes, map[string]any{
				"case": fmt.Sprintf("alias-key-%02d", index+1), "request": input,
				"result": sourceAliasMatchKey(input),
			})
		}
		return outcomes, nil
	})
}

func TestNetworkFlowUptimeFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-012-sys-uptime-timestamps", 3, mappingFixtureOutcomes)
}

func TestNetworkFlowTimestampFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-022-timestamp-rulesets", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		data, err := bundle.Read("source/timezone-provenance.json")
		if err != nil {
			return nil, err
		}
		var provenance struct {
			Ruleset string `json:"ruleset_id"`
			Archive struct {
				SHA256 string `json:"sha256"`
				Size   int64  `json:"size_bytes"`
			} `json:"source_archive"`
		}
		if err := json.Unmarshal(data, &provenance); err != nil {
			return nil, err
		}
		archive, err := bundle.Read("source/tzdata2026c.tar.gz")
		if err != nil {
			return nil, err
		}
		if provenance.Ruleset != networkflowtz.RulesetID || provenance.Archive.SHA256 != sha256Hex(archive) || provenance.Archive.Size != int64(len(archive)) {
			return nil, fmt.Errorf("timestamp source archive does not match pinned provenance")
		}
		outcomes, err := mappingFixtureOutcomes(bundle)
		if err != nil {
			return nil, err
		}
		return append([]map[string]any{{"case": "ruleset-source", "request": map[string]any{"ruleset_id": provenance.Ruleset, "source": "source/tzdata2026c.tar.gz"}, "result": map[string]any{"archive_sha256": sha256Hex(archive), "size_bytes": len(archive), "embedded_ruleset_id": networkflowtz.RulesetID}}}, outcomes...), nil
	})
}

func TestNetworkFlowOrdinalFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-003-duplicate-headers", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		assertRetainedRawCanonical(t)
		return mappingFixtureOutcomes(bundle)
	})
}

func mappingFixtureOutcomes(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
	data, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var cases []fixtureMappingCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}
	outcomes := []map[string]any{}
	for _, test := range cases {
		source, err := bundle.Read(test.Source)
		if err != nil {
			return nil, err
		}
		facade := &importFacade{sourceStore: previewSource{csv: string(source)}, limits: defaultLimits()}
		inspection, err := facade.InspectImportUnitSource(context.Background(), imports.ExtensionImportSourceRequest{
			TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID,
		})
		if err != nil {
			if translated, ok := facade.TranslateImportUnitError(err); ok {
				return nil, fmt.Errorf("%s: source inspection: %s %v", test.Name, translated.OwnerError.OwnerCode, translated.OwnerError.SafeDetails)
			}
			return nil, fmt.Errorf("%s: source inspection: %w", test.Name, err)
		}
		if err := facade.ValidateImportUnitSourceResult(inspection); err != nil {
			return nil, err
		}
		prepared, err := facade.PrepareImportUnitMapping(context.Background(), imports.ExtensionImportMappingRequest{
			TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID,
			OwnerMappingSchemaID: mappingCandidateSchemaID, OwnerMapping: test.Mapping,
		})
		outcome := map[string]any{"case": test.Name, "request": test, "inspection": inspection}
		if err != nil {
			translated, ok := facade.TranslateImportUnitError(err)
			if !ok {
				return nil, fmt.Errorf("%s: unexpected mapping error: %w", test.Name, err)
			}
			if err := facade.ValidateImportUnitError(translated.OwnerError); err != nil {
				return nil, err
			}
			outcome["result"] = map[string]any{"status": "rejected", "error": translated}
		} else {
			if err := facade.ValidateImportUnitMappingResult(prepared); err != nil {
				return nil, err
			}
			var mapping approvedMapping
			if err := json.Unmarshal(prepared.OwnerMapping, &mapping); err != nil {
				return nil, err
			}
			parsed, err := parseCSVPreview(strings.NewReader(string(source)), sha256Hex(source), defaultLimits())
			if err != nil {
				return nil, err
			}
			records := []map[string]any{}
			for _, record := range parsed.Records {
				row, diagnostics, err := validateCSVRecord(record, parsed.SourceColumns, mapping, prepared.MappingFingerprint, parsed.SourceContentSHA256, defaultLimits(), sourceFieldMappings(mapping))
				if err != nil {
					return nil, err
				}
				if len(diagnostics) > 0 {
					resources := []map[string]any{}
					for _, diagnostic := range diagnostics {
						resources = append(resources, diagnosticResource(diagnostic))
					}
					records = append(records, map[string]any{"status": "rejected", "source_row_number": record.SourceRowNumber, "diagnostics": resources})
				} else {
					records = append(records, normalizedFixtureRecord(row))
				}
			}
			outcome["result"] = map[string]any{"status": "accepted", "prepared": prepared, "normalized_records": records}
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes, nil
}

// Only semantic input and source digests belong to normalization. Storage
// attribution, identities and observation provenance are compared after apply.
func normalizedFixtureRecord(row normalizedRow) map[string]any {
	return map[string]any{
		"status": "accepted", "source_row_number": row.SourceRowNumber,
		"source_row_digest_sha256":     row.SourceRowDigestSHA256,
		"normalized_row_digest_sha256": row.NormalizedRowDigestSHA256,
		"unmapped_raw":                 row.UnmappedRaw,
		"values": map[string]any{
			fieldFlowStartUTC: formatTimestamp(row.FlowStartUTC), fieldFlowEndUTC: formatTimestamp(row.FlowEndUTC),
			fieldSrcIP: row.SrcIP, fieldDstIP: row.DstIP, fieldSrcPort: row.SrcPort, fieldDstPort: row.DstPort,
			fieldIPProtocol: row.IPProtocol, fieldBytesCount: row.BytesCount, fieldPacketsCount: row.PacketsCount,
			fieldExporterID: row.ExporterID, fieldInputInterface: row.InputInterface, fieldOutputInterface: row.OutputInterface,
			fieldTCPFlags: row.TCPFlags, fieldApplicationLabel: row.ApplicationLabel,
		},
	}
}

func assertRetainedRawCanonical(t *testing.T) {
	t.Helper()
	const emptySHA = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	value := func(ordinal int) string {
		return fmt.Sprintf(`{"source_column_ordinal":%d,"raw_header_text":"","raw_header_sha256":"%s","decoded_value":"","decoded_value_sha256":"%s"}`, ordinal, emptySHA, emptySHA)
	}
	want := `{"2":` + value(2) + `,"10":` + value(10) + `}`
	var values retainedRawValues
	if err := json.Unmarshal([]byte(`{"10":`+value(10)+`,"2":`+value(2)+`}`), &values); err != nil {
		t.Fatal(err)
	}
	actual, err := canonicalJSON(values)
	if err != nil || string(actual) != want {
		t.Fatalf("typed ordinal canonical order: %s, %v", actual, err)
	}
	invalid := []string{
		`null`, `[]`, `{"2":null}`, `{"02":` + value(2) + `}`, `{"0":` + value(0) + `}`,
		`{"-2":` + value(-2) + `}`, `{"2":` + value(10) + `}`, `{"2":` + value(2) + `,"2":` + value(2) + `}`,
		strings.Replace(want, `"source_column_ordinal":2`, `"source_column_ordinal":2.0`, 1),
		strings.Replace(want, `"raw_header_text":"",`, ``, 1),
		strings.Replace(want, `"raw_header_text":""`, `"raw_header_text":null`, 1),
		strings.Replace(want, `"raw_header_text":""`, `"raw_header_text":"changed"`, 1),
		strings.Replace(want, `"decoded_value":""`, `"decoded_value":"changed"`, 1),
		strings.Replace(want, `"decoded_value":""`, `"decoded_value":"","extra":1`, 1),
	}
	for _, raw := range invalid {
		var value retainedRawValues
		if err := json.Unmarshal([]byte(raw), &value); err == nil {
			t.Fatal("malformed retained provenance admitted:", raw)
		}
	}
	for _, value := range []retainedRawValues{nil, {0: {SourceColumnOrdinal: 0}}, {2: {SourceColumnOrdinal: 3}}, {2: {SourceColumnOrdinal: 2}}} {
		if data, err := canonicalJSON(value); err == nil || data != nil {
			t.Fatal("malformed typed provenance produced canonical bytes")
		}
	}
}
