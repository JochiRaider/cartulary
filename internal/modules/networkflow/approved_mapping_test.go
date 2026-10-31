package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

func approvedVariantFixture(t *testing.T, profileID string) (*importFacade, imports.ExtensionImportMappingResult, string) {
	t.Helper()
	return approvedVariantFixtureRecords(t, profileID, 1)
}

func approvedVariantFixtureRecords(t *testing.T, profileID string, recordCount int) (*importFacade, imports.ExtensionImportMappingResult, string) {
	t.Helper()
	mapping := approvedMappingFixture(profileID)
	mapping.DisplayNameOverride = stringPtr("flows")
	data := strings.TrimSuffix(validPreviewRow, "\n") + ",ignored\n"
	if profileID != sourceProfileCiscoSNANetFlowCSV {
		mapping.TimestampProfile = materializeTimestampProfile(timestampProfile{Mode: "cisco_sna_datetime_v1"})
		data = strings.ReplaceAll(data, "Z", "+0000")
	}
	mapping.FieldMappings = append(mapping.FieldMappings, fieldMapping{MappingKind: mappingKindIgnoredSourceColumn, SourceColumnOrdinal: 10, IgnoreReason: "user_ignored"})
	var candidate map[string]any
	if err := json.Unmarshal(marshalApprovedMapping(mapping), &candidate); err != nil {
		t.Fatal(err)
	}
	delete(candidate, "source_columns")
	raw, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	data = strings.Repeat(data, recordCount)
	if recordCount > 1 {
		// Equal sample counts do not require equal empty-value counts.
		data = strings.Replace(data, ",ignored\n", ",\n", 1)
	}
	source := strings.Join(requiredFields(sourceProfileCiscoSNANetFlowCSV), ",") + ",ignored\n" + data
	f := &importFacade{sourceStore: previewSource{csv: source}, limits: defaultLimits()}
	result, err := f.PrepareImportUnitMapping(context.Background(), imports.ExtensionImportMappingRequest{TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID, OwnerMappingSchemaID: mappingCandidateSchemaID, OwnerMapping: raw})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.ValidateImportUnitMappingResult(result); err != nil || result.OwnerResult["preview_accepted_count"] != int64(recordCount) {
		t.Fatalf("invalid positive preview control: %v, %#v", err, result.OwnerResult)
	}
	return f, result, source
}

type approvedVariantMutation struct {
	name, reason string
	raw          json.RawMessage
}

func approvedVariantMutations(t *testing.T, baseline json.RawMessage) []approvedVariantMutation {
	t.Helper()
	var original map[string]json.RawMessage
	if err := json.Unmarshal(baseline, &original); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(original["field_mappings"], &rows); err != nil {
		t.Fatal(err)
	}
	cases := mappingClosedObjectMutations(t, baseline, false)
	add := func(index int, name, reason, member string, value json.RawMessage) {
		t.Helper()
		var fields []map[string]json.RawMessage
		if err := json.Unmarshal(original["field_mappings"], &fields); err != nil {
			t.Fatal(err)
		}
		if value == nil {
			delete(fields[index], member)
		} else {
			fields[index][member] = value
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		object := make(map[string]json.RawMessage, len(original))
		for key, value := range original {
			object[key] = value
		}
		object["field_mappings"] = encoded
		raw, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, approvedVariantMutation{name, reason, raw})
	}
	for index, row := range rows {
		var kind string
		if err := json.Unmarshal(row["mapping_kind"], &kind); err != nil {
			t.Fatal(err)
		}
		// One source row suffices for shape; all source fields' policies have a
		// separate presence and profile-permission matrix.
		if kind == mappingKindSourceColumn && index != 0 {
			continue
		}

		if kind != mappingKindSourceColumn {
			for _, policy := range []string{`"profile_default"`, `"empty_string_is_null"`, `"empty_string_is_invalid"`, `"empty_string_preserved"`, `null`, `""`, `" "`, `"unknown"`, `1`, `false`, `[]`, `{}`} {
				add(index, kind+"/policy/"+policy, "variant_member_conflict", "empty_value_policy", json.RawMessage(policy))
			}
		}
		if kind == mappingKindIgnoredSourceColumn {
			add(index, kind+"/invalid_reason", "variant_member_conflict", "ignore_reason", json.RawMessage(`"unknown"`))
		} else {
			add(index, kind+"/invalid_combinability", "variant_member_conflict", "combinability", json.RawMessage(`"unknown"`))
		}
	}
	for _, kind := range []string{"constant", "derived", "unknown"} {
		add(0, "unavailable/"+kind, "mapping_kind_unavailable", "mapping_kind", json.RawMessage(fmt.Sprintf("%q", kind)))
	}
	return append(cases, approvedSampleCountMutations(t, baseline)...)
}

func approvedSampleCountMutations(t *testing.T, baseline json.RawMessage) []approvedVariantMutation {
	t.Helper()
	var cases []approvedVariantMutation
	for _, position := range []string{"first", "middle", "last"} {
		for _, direction := range []string{"shorter", "longer"} {
			mapping, err := decodeApprovedMapping(baseline, defaultLimits())
			if err != nil {
				t.Fatalf("invalid sample-count control: %v", err)
			}
			index := map[string]int{"first": 0, "middle": len(mapping.SourceColumns) / 2, "last": len(mapping.SourceColumns) - 1}[position]
			column := &mapping.SourceColumns[index]
			if direction == "shorter" {
				column.SampleValues = column.SampleValues[:len(column.SampleValues)-1]
			} else {
				column.SampleValues = append(column.SampleValues, column.SampleValues[0])
			}
			column.DetectedEmptyCount = 0
			for _, sample := range column.SampleValues {
				if *sample.RawValueSHA256 == sha256Hex(nil) {
					column.DetectedEmptyCount++
				}
			}
			cases = append(cases, approvedVariantMutation{"sample_count/" + position + "/" + direction, "variant_member_conflict", marshalApprovedMapping(mapping)})
		}
	}
	return cases
}

func TestApprovedMappingFieldVariants(t *testing.T) {
	for _, profileID := range []string{sourceProfileCiscoSNANetFlowCSV, "cisco_sna_flow_search_csv_v1"} {
		t.Run(profileID, func(t *testing.T) {
			_, result, source := approvedVariantFixture(t, profileID)
			mapping, err := decodeApprovedMapping(result.OwnerMapping, defaultLimits())
			if err != nil || !bytes.Equal(marshalApprovedMapping(mapping), result.OwnerMapping) {
				t.Fatalf("valid approved round-trip changed: %v", err)
			}
			escaped := bytes.Replace(result.OwnerMapping, []byte(`"target_kind"`), []byte(`"\u0074arget_kind"`), 1)
			if decoded, err := decodeApprovedMapping(escaped, defaultLimits()); err != nil || !bytes.Equal(marshalApprovedMapping(decoded), result.OwnerMapping) {
				t.Fatalf("exact escaped member rejected or changed identity: %v", err)
			}
			fingerprint, err := mappingFingerprint(mapping, sha256Hex([]byte(source)))
			if err != nil || fingerprint != result.MappingFingerprint {
				t.Fatalf("valid fingerprint changed: %v", err)
			}
			t.Run("shared descriptor limits", func(t *testing.T) {
				for _, limited := range []EffectiveLimits{
					func() EffectiveLimits {
						limits := defaultLimits()
						limits.MaxColumnsPerCSV = int64(len(mapping.SourceColumns) - 1)
						return limits
					}(),
					func() EffectiveLimits { limits := defaultLimits(); limits.MaxHeaderScalarLength = 1; return limits }(),
				} {
					if _, err := decodeApprovedMapping(result.OwnerMapping, limited); err == nil {
						t.Fatal("approved descriptor bypassed effective limits")
					}
				}
				_, fifty, _ := approvedVariantFixtureRecords(t, profileID, previewRecordLimit)
				bounded, err := decodeApprovedMapping(fifty.OwnerMapping, defaultLimits())
				if err != nil {
					t.Fatal(err)
				}
				for _, column := range bounded.SourceColumns {
					if len(column.SampleValues) != previewRecordLimit {
						t.Fatal("fixed-bound control did not sample every column")
					}
				}
				if bounded.SourceColumns[0].DetectedEmptyCount != 0 || bounded.SourceColumns[9].DetectedEmptyCount != 1 {
					t.Fatal("fixed-bound control lost differing empty-value counts")
				}
				for _, columns := range [][]sourceColumnDescriptor{mapping.SourceColumns[:1], bounded.SourceColumns[:1]} {
					if _, err := decodeSourceColumns(mappingTestJSON(t, columns), defaultLimits()); err != nil {
						t.Fatalf("single-column sample control rejected: %v", err)
					}
				}
				// Stored admission has a fixed bound; only preview knows actual count.
				if _, err := decodeApprovedMapping(marshalApprovedMapping(bounded), defaultLimits()); err != nil {
					t.Fatalf("fixed-bound control rejected: %v", err)
				}
				if _, err := decodePreviewMapping(marshalApprovedMapping(bounded), defaultLimits(), 1); err == nil {
					t.Fatal("preview admitted more samples than actual records")
				}
				for _, mutation := range approvedSampleCountMutations(t, fifty.OwnerMapping) {
					t.Run(mutation.name, func(t *testing.T) {
						var conflict *mappingValidationError
						if _, err := decodeApprovedMapping(mutation.raw, defaultLimits()); !errors.As(err, &conflict) || conflict.Code != "network_flow_mapping_conflict" || conflict.ReasonCode != mutation.reason {
							t.Fatalf("inconsistent bounded sample counts admitted: %v", err)
						}
					})
				}
				// Every column exceeds the bound, so inequality cannot mask a missing bound.
				for i := range bounded.SourceColumns {
					column := &bounded.SourceColumns[i]
					column.SampleValues = append(column.SampleValues, column.SampleValues[0])
					if *column.SampleValues[0].RawValueSHA256 == sha256Hex(nil) {
						column.DetectedEmptyCount++
					}
				}
				if _, err := decodeApprovedMapping(marshalApprovedMapping(bounded), defaultLimits()); err == nil {
					t.Fatal("approved sample bound exceeded")
				}
			})
			stream := &applySourceFixture{source: source, capability: imports.ImportSourceCapability{SourceStreamRef: "impsrc_" + strings.Repeat("1", 32), SourceContentSHA256: sha256Hex([]byte(source)), SourceMediaType: "text/csv", SourceByteSize: int64(len(source))}}
			f := &importFacade{sourceStore: stream, limits: defaultLimits(), safeDigester: &queryCompletionProbe{}, now: func() time.Time { return time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC) }}
			request := imports.ExtensionImportApplyRequest{SourceCapability: stream.capability, ExpectedSourceContentSHA256: stream.capability.SourceContentSHA256, MappingFingerprint: fingerprint, OwnerMapping: result.OwnerMapping}
			if prepared, err := f.prepareImportApply(context.Background(), nil, request); err != nil || prepared.params.VisitRows == nil || stream.opens != 1 {
				t.Fatalf("valid apply control failed: %v", err)
			}
			for _, mutation := range approvedVariantMutations(t, result.OwnerMapping) {
				t.Run(mutation.name, func(t *testing.T) {
					_, err := decodeApprovedMapping(mutation.raw, defaultLimits())
					var conflict *mappingValidationError
					if !errors.As(err, &conflict) || conflict.Code != "network_flow_mapping_conflict" || conflict.ReasonCode != mutation.reason {
						t.Errorf("approved admission: expected %s, got %v", mutation.reason, err)
					}
					stream.opens, stream.closes = 0, 0
					request.OwnerMapping = mutation.raw // Keep the valid fingerprint.
					prepared, err := f.prepareImportApply(context.Background(), nil, request)
					var owner *networkFlowImportOwnerError
					if !errors.As(err, &owner) || owner.ownerCode != "network_flow_mapping_invalid" || owner.safeDetails["reason_code"] != mutation.reason || owner.retryable || prepared.params.VisitRows != nil || stream.opens != 0 || stream.closes != 0 {
						t.Errorf("apply must reject before source access/publication: opens=%d err=%v", stream.opens, err)
					}
				})
			}
		})
	}
}

func assertApprovedPreviewFieldVariants(t *testing.T) {
	t.Helper()
	for _, profileID := range []string{sourceProfileCiscoSNANetFlowCSV, "cisco_sna_flow_search_csv_v1"} {
		f, result, _ := approvedVariantFixture(t, profileID)
		reordered := rewriteMappingObject(t, result.OwnerMapping, nil, func(object map[string]json.RawMessage) json.RawMessage {
			var rows []json.RawMessage
			if err := json.Unmarshal(object["field_mappings"], &rows); err != nil {
				t.Fatal(err)
			}
			for left, right := 0, len(rows)-1; left < right; left, right = left+1, right-1 {
				rows[left], rows[right] = rows[right], rows[left]
			}
			object["field_mappings"] = mappingTestJSON(t, rows)
			return mappingTestJSON(t, object)
		})
		control := result
		control.OwnerMapping = reordered
		if err := f.ValidateImportUnitMappingResult(control); err != nil {
			t.Fatalf("permitted member and mapping-row reordering rejected: %v", err)
		}
		for _, mutation := range approvedVariantMutations(t, result.OwnerMapping) {
			boundaries := []string{"outer", "inner", "both"}
			if strings.HasPrefix(mutation.name, "sample_count/") {
				boundaries = append(boundaries, "all_descriptors")
			}
			for _, boundary := range boundaries {
				t.Run(profileID+"/"+mutation.name+"/"+boundary, func(t *testing.T) {
					bad := result
					bad.OwnerResult = make(map[string]any, len(result.OwnerResult))
					for key, value := range result.OwnerResult {
						bad.OwnerResult[key] = value
					}
					if boundary != "inner" {
						bad.OwnerMapping = mutation.raw
					}
					if boundary != "outer" {
						bad.OwnerResult["materialized_mapping"] = mutation.raw
					}
					if boundary == "all_descriptors" {
						var mapping map[string]json.RawMessage
						if err := json.Unmarshal(mutation.raw, &mapping); err != nil {
							t.Fatal(err)
						}
						bad.OwnerResult["source_columns"] = mapping["source_columns"]
					}
					if err := f.ValidateImportUnitMappingResult(bad); !errors.Is(err, errInvalidImportObject) {
						t.Fatalf("malformed approved result admitted or wrong error: %v", err)
					}
				})
			}
		}
		t.Run(profileID+"/equal_counts_below_actual", func(t *testing.T) {
			mapping, err := decodeApprovedMapping(result.OwnerMapping, defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			for i := range mapping.SourceColumns {
				mapping.SourceColumns[i].SampleValues = []safeSample{}
				mapping.SourceColumns[i].DetectedEmptyCount = 0
			}
			raw := marshalApprovedMapping(mapping)
			if _, err := decodeApprovedMapping(raw, defaultLimits()); err != nil {
				t.Fatalf("internally consistent zero-count approved control rejected: %v", err)
			}
			bad := result
			bad.OwnerMapping = raw
			bad.OwnerResult = make(map[string]any, len(result.OwnerResult))
			for key, value := range result.OwnerResult {
				bad.OwnerResult[key] = value
			}
			bad.OwnerResult["materialized_mapping"] = raw
			bad.OwnerResult["source_columns"] = mapping.SourceColumns
			if err := f.ValidateImportUnitMappingResult(bad); !errors.Is(err, errInvalidImportObject) {
				t.Fatalf("preview ignored actual sampled-record count: %v", err)
			}
		})
	}
}
