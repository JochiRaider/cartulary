package networkflow

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCiscoSNAFlowSearchProfile(t *testing.T) {
	assertCiscoDefaultPolicies(t)
	const profileID = "cisco_sna_flow_search_csv_v1"
	profile := timestampProfile{SchemaID: timestampProfileSchemaID, Mode: "cisco_sna_datetime_v1", Precision: "milliseconds"}
	for value, want := range map[string]string{
		"2026-10-29T00:00:07.117+0000": "2026-10-29T00:00:07.117Z",
		"2026-10-29T02:30:07.100+0230": "2026-10-29T00:00:07.1Z",
		"0001-01-01T00:00:00+0000":     "0001-01-01T00:00:00Z",
		"9999-12-31T23:59:59.999+0000": "9999-12-31T23:59:59.999Z",
	} {
		got, err := parseTimestampForRecord(value, profile, nil)
		if err != nil || formatTimestamp(got) != want {
			t.Fatalf("%s: %v, %v", value, got, err)
		}
		if _, err := parseExactRFC3339(value, timestampProfile{Mode: "rfc3339", Precision: "milliseconds"}); err == nil {
			t.Fatal("compact offset relaxed RFC3339")
		}
	}
	for _, value := range []string{
		"2026-10-29T00:00:00Z", "2026-10-29T00:00:00+00:00", "2026-10-29T00:00:00", "2026-10-29t00:00:00+0000",
		"2026-10-29T00:00:00-0000", "2026-10-29T24:00:00+0000", "2026-10-29T00:00:60+0000", "2026-02-30T00:00:00+0000",
		"2026-10-29T00:00:00+2400", "2026-10-29T00:00:00+0060", "2026-10-29T00:00:00.1234+0000",
		"0001-01-01T00:00:00+0100", "9999-12-31T23:59:59-0100", "2026-10-29T00:00:00+0000 ",
	} {
		if _, err := parseTimestampForRecord(value, profile, nil); err == nil {
			t.Fatalf("invalid source timestamp admitted: %s", value)
		}
	}
	registry, ok := mappingSourceProfile(profileID)
	if !ok || len(sourceProfileResources()) != 2 {
		t.Fatal("Flow Search profile unavailable")
	}
	profiles := sourceProfileResources()
	if profiles[0].(map[string]any)["source_profile_id"] != profileID || profiles[1].(map[string]any)["source_profile_id"] != sourceProfileCiscoSNANetFlowCSV {
		t.Fatal("discovery is not ordered by profile ID")
	}
	if len(profiles[0].(map[string]any)["optional_field_keys"].([]string)) != 0 || len(profiles[1].(map[string]any)["optional_field_keys"].([]string)) != 2 {
		t.Fatal("discovery advertises unsupported optional fields")
	}
	var headers []string
	var fields []fieldMapping
	for _, field := range registry.Fields {
		if field.Requirement != "required" {
			continue
		}
		headers = append(headers, field.Aliases[0])
		fields = append(fields, fieldMapping{MappingKind: mappingKindSourceColumn, FieldKey: field.FieldKey, SourceColumnOrdinal: len(headers), TransformID: field.TransformID, Combinability: "single_source_only"})
	}
	source := strings.Join(headers, ",") + "\n2026-10-29T00:00:00.000+0000,2026-10-29T00:00:01.000+0000,192.0.2.1,192.0.2.2,,,ICMP,100,2\n"
	parsed, err := parseCSVApply(strings.NewReader(source), "", defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	candidate := mappingCandidate{TargetKind: targetKindNetworkFlowTable, TargetTableSchemaID: targetTableSchemaID, SourceProfileID: profileID, TimestampProfile: profile, FieldMappings: candidateFieldMappingsForTest(t, fields)}
	raw, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := materializeApprovedMapping(raw, parsed.SourceColumns)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := mappingFingerprint(mapping, parsed.SourceContentSHA256)
	if err != nil {
		t.Fatal(err)
	}
	row, diagnostics, err := validateCSVRecord(parsed.Records[0], parsed.SourceColumns, mapping, fingerprint, parsed.SourceContentSHA256, defaultLimits(), sourceFieldMappings(mapping))
	if err != nil || len(diagnostics) != 0 || row.SrcPort != nil || row.DstPort != nil || row.IPProtocol != 1 {
		t.Fatalf("nullable ICMP: %#v %#v %v", row, diagnostics, err)
	}
	if sourceMappableField(profileID, fieldInputInterface) || sourceMappableField(profileID, fieldExporterID) {
		t.Fatal("endpoint-relative data became a mapped role")
	}
	for _, member := range []string{"timezone", "timezone_ruleset_id", "netflow_export_time_column_ordinal"} {
		var request map[string]any
		if err := json.Unmarshal(raw, &request); err != nil {
			t.Fatal(err)
		}
		request["timestamp_profile"].(map[string]any)[member] = nil
		invalid, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := materializeApprovedMapping(invalid, parsed.SourceColumns); err == nil {
			t.Fatalf("compact mode accepted %s", member)
		}
	}
	// An already approved mapping keeps its profile's interpretation. Selecting a
	// different profile cannot silently admit a mode that profile never supported.
	mapping.SourceProfileID = sourceProfileCiscoSNANetFlowCSV
	if err := validateApprovedMapping(mapping); err == nil {
		t.Fatal("old profile accepted new timestamp mode")
	}
}

func assertCiscoDefaultPolicies(t *testing.T) {
	t.Helper()
	for _, profileID := range []string{sourceProfileCiscoSNANetFlowCSV, "cisco_sna_flow_search_csv_v1"} {
		registry, _ := mappingSourceProfile(profileID)
		flowSearch := profileID == "cisco_sna_flow_search_csv_v1"
		stamp := "2026-10-29T00:00:00Z"
		profile := materializeTimestampProfile(timestampProfile{Mode: "rfc3339"})
		if flowSearch {
			stamp = "2026-10-29T00:00:00+0000"
			profile = materializeTimestampProfile(timestampProfile{Mode: "cisco_sna_datetime_v1"})
		}
		headers, fields := []string{}, []fieldMapping{}
		for _, field := range registry.Fields {
			if field.Requirement == "required" || field.Requirement == "optional_map_when_present" {
				headers = append(headers, field.Aliases[0])
				fields = append(fields, fieldMapping{MappingKind: mappingKindSourceColumn, FieldKey: field.FieldKey, SourceColumnOrdinal: len(headers), TransformID: field.TransformID, Combinability: "single_source_only"})
			}
		}
		values := []string{stamp, stamp, "192.0.2.1", "192.0.2.2", "", "", "ICMP", "100", "2"}
		if !flowSearch {
			values = append(values, "", "")
		}
		source := strings.Join(headers, ",") + "\n" + strings.Join(values, ",") + "\n"
		parsed, err := parseCSVApply(strings.NewReader(source), "", defaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		var baseline approvedMapping
		var baselineFingerprint string
		for _, policy := range []string{"", "profile_default", "explicit"} {
			candidateFields := append([]fieldMapping(nil), fields...)
			for i := range candidateFields {
				field := &candidateFields[i]
				field.EmptyValuePolicy = policy
				if policy == "explicit" {
					field.EmptyValuePolicy = emptyPolicyInvalid
					if field.FieldKey == fieldInputInterface || field.FieldKey == fieldOutputInterface || flowSearch && (field.FieldKey == fieldSrcPort || field.FieldKey == fieldDstPort) {
						field.EmptyValuePolicy = emptyPolicyNull
					}
				}
			}
			candidate := mappingCandidate{TargetKind: targetKindNetworkFlowTable, TargetTableSchemaID: targetTableSchemaID, SourceProfileID: profileID, TimestampProfile: profile, FieldMappings: candidateFieldMappingsForTest(t, candidateFields)}
			raw, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			approved, err := materializeApprovedMapping(raw, parsed.SourceColumns)
			if err != nil {
				t.Fatalf("%s/%q: %v", profileID, policy, err)
			}
			fingerprint, err := mappingFingerprint(approved, parsed.SourceContentSHA256)
			if err != nil {
				t.Fatal(err)
			}
			if policy == "" {
				baseline, baselineFingerprint = approved, fingerprint
			} else if !reflect.DeepEqual(baseline, approved) || fingerprint != baselineFingerprint {
				t.Fatalf("%s/%q changed approved defaults or fingerprint", profileID, policy)
			}
			for _, field := range approved.FieldMappings {
				if field.EmptyValuePolicy == "profile_default" {
					t.Fatal("unmaterialized approved policy")
				}
			}
		}
		assertCiscoPolicyAdmission(t, baseline)
		validate := func(cells []string, mapping approvedMapping) (normalizedRow, []rejectedRowDiagnostic) {
			t.Helper()
			input, err := parseCSVApply(strings.NewReader(strings.Join(headers, ",")+"\n"+strings.Join(cells, ",")+"\n"), "", defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			record := input.Records[0]
			row, diagnostics, err := validateCSVRecord(record, parsed.SourceColumns, mapping, baselineFingerprint, parsed.SourceContentSHA256, defaultLimits(), sourceFieldMappings(mapping))
			if err != nil {
				t.Fatal(err)
			}
			return row, diagnostics
		}
		row, diagnostics := validate(values, baseline)
		if flowSearch {
			if len(diagnostics) != 0 || row.SrcPort != nil || row.DstPort != nil {
				t.Fatalf("nullable default: %#v", diagnostics)
			}
		} else if len(diagnostics) != 2 || diagnostics[0].ErrorCode != "network_flow_invalid_port" || diagnostics[1].ErrorCode != "network_flow_invalid_port" {
			t.Fatalf("original blank ports: %#v", diagnostics)
		}
		zero := append([]string(nil), values...)
		zero[4], zero[5] = "0", "0"
		row, diagnostics = validate(zero, baseline)
		if len(diagnostics) != 0 || row.SrcPort == nil || *row.SrcPort != 0 || row.DstPort == nil || *row.DstPort != 0 {
			t.Fatalf("zero ports: %#v", diagnostics)
		}
		invalid := append([]string(nil), zero...)
		invalid[4], invalid[5] = "-1", "65536"
		if _, diagnostics = validate(invalid, baseline); len(diagnostics) != 2 {
			t.Fatalf("invalid nonempty ports: %#v", diagnostics)
		}
		for _, index := range []int{0, 1, 2, 3, 6, 7, 8} {
			empty := append([]string(nil), zero...)
			empty[index] = ""
			if _, diagnostics = validate(empty, baseline); len(diagnostics) != 1 {
				t.Fatalf("%s required field %d: %#v", profileID, index, diagnostics)
			}
		}
		if flowSearch {
			explicit := baseline
			explicit.FieldMappings = append([]fieldMapping(nil), baseline.FieldMappings...)
			for i := range explicit.FieldMappings {
				if explicit.FieldMappings[i].FieldKey == fieldSrcPort || explicit.FieldMappings[i].FieldKey == fieldDstPort {
					explicit.FieldMappings[i].EmptyValuePolicy = emptyPolicyInvalid
				}
			}
			if err := validateApprovedMapping(explicit); err != nil {
				t.Fatal(err)
			}
			if _, diagnostics = validate(values, explicit); len(diagnostics) != 2 {
				t.Fatal("explicit invalid-empty port policy was overridden")
			}
			unsupported := explicit
			unsupported.FieldMappings = append(append([]fieldMapping(nil), explicit.FieldMappings...), fieldMapping{MappingKind: mappingKindSourceColumn, FieldKey: fieldInputInterface, SourceColumnOrdinal: 1, TransformID: "trim_ascii_space_v1", EmptyValuePolicy: emptyPolicyNull, Combinability: "single_source_only"})
			if err := validateApprovedMapping(unsupported); err == nil {
				t.Fatal("Flow Search interface mapping admitted")
			}
		}
	}
}

func assertCiscoPolicyAdmission(t *testing.T, baseline approvedMapping) {
	t.Helper()
	for index, field := range baseline.FieldMappings {
		if field.MappingKind != mappingKindSourceColumn {
			continue
		}
		for _, policy := range []string{emptyPolicyInvalid, emptyPolicyNull} {
			t.Run(baseline.SourceProfileID+"/"+field.FieldKey+"/"+policy, func(t *testing.T) {
				mapping := baseline
				mapping.FieldMappings = append([]fieldMapping(nil), baseline.FieldMappings...)
				mapping.FieldMappings[index].EmptyValuePolicy = policy
				forbidden := baseline.SourceProfileID == "cisco_sna_flow_search_csv_v1" && policy == emptyPolicyNull && field.FieldKey != fieldSrcPort && field.FieldKey != fieldDstPort
				check := func(boundary string, err error) {
					t.Helper()
					if !forbidden {
						if err != nil {
							t.Errorf("%s rejected a permitted explicit policy: %v", boundary, err)
						}
						return
					}
					var conflict *mappingValidationError
					if !errors.As(err, &conflict) || conflict.Code != "network_flow_mapping_conflict" || conflict.ReasonCode != "invalid_empty_value_policy" || conflict.FieldKey != field.FieldKey {
						t.Errorf("%s: expected field-specific policy conflict, got %v", boundary, err)
					}
				}
				check("approved validation", validateApprovedMapping(mapping))
				decoded, err := decodeApprovedMapping(marshalApprovedMapping(mapping), defaultLimits())
				check("approved decoding", err)
				if !forbidden && err == nil && !reflect.DeepEqual(mapping, decoded) {
					t.Error("approved decoding changed a permitted mapping")
				}
				candidate := mappingCandidate{TargetKind: mapping.TargetKind, TargetTableSchemaID: mapping.TargetTableSchemaID, SourceProfileID: mapping.SourceProfileID, ParserProfileID: mapping.ParserProfileID, UnknownColumnPolicy: mapping.UnknownColumnPolicy, TimestampProfile: mapping.TimestampProfile, FieldMappings: candidateFieldMappingsForTest(t, mapping.FieldMappings)}
				raw, err := json.Marshal(candidate)
				if err != nil {
					t.Fatal(err)
				}
				materialized, err := materializeApprovedMapping(raw, mapping.SourceColumns)
				check("candidate materialization", err)
				if !forbidden && err == nil && !reflect.DeepEqual(mapping, materialized) {
					t.Error("materialization changed a permitted explicit policy")
				}
			})
		}
	}
}

func candidateFieldMappingsForTest(t *testing.T, fields []fieldMapping) []candidateFieldMapping {
	t.Helper()
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	var candidates []candidateFieldMapping
	if err := json.Unmarshal(raw, &candidates); err != nil {
		t.Fatal(err)
	}
	return candidates
}
