package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestMappingPolicyJSONPresence(t *testing.T) {
	for _, profileID := range []string{sourceProfileCiscoSNANetFlowCSV, "cisco_sna_flow_search_csv_v1"} {
		t.Run(profileID, func(t *testing.T) {
			keys := []string{fieldFlowStartUTC, fieldFlowEndUTC, fieldSrcIP, fieldDstIP, fieldSrcPort, fieldDstPort, fieldIPProtocol, fieldBytesCount, fieldPacketsCount}
			profile := map[string]any{"schema_id": timestampProfileSchemaID, "mode": "cisco_sna_datetime_v1", "precision": "milliseconds"}
			if profileID == sourceProfileCiscoSNANetFlowCSV {
				keys = append(keys, fieldInputInterface, fieldOutputInterface)
				profile = map[string]any{"schema_id": timestampProfileSchemaID, "mode": "rfc3339", "precision": "milliseconds", "timezone": nil, "timezone_ruleset_id": nil, "ambiguous_local_time_policy": "reject", "local_time_gap_policy": "reject"}
			}
			fields := make([]map[string]any, len(keys))
			for i, key := range keys {
				policy := emptyPolicyInvalid
				if key == fieldInputInterface || key == fieldOutputInterface || profileID != sourceProfileCiscoSNANetFlowCSV && (key == fieldSrcPort || key == fieldDstPort) {
					policy = emptyPolicyNull
				}
				fields[i] = map[string]any{"mapping_kind": mappingKindSourceColumn, "field_key": key, "source_column_ordinal": i + 1, "transform_id": defaultTransformForField(profileID, key), "empty_value_policy": policy, "combinability": "single_source_only"}
			}
			candidate := map[string]any{"target_kind": targetKindNetworkFlowTable, "target_table_schema_id": targetTableSchemaID, "source_profile_id": profileID, "timestamp_profile": profile, "field_mappings": fields}
			stamp := "2026-10-29T00:00:00+0000"
			if profileID == sourceProfileCiscoSNANetFlowCSV {
				stamp = "2026-10-29T00:00:00Z"
			}
			values := []string{stamp, stamp, "192.0.2.1", "198.51.100.2", "12345", "443", "TCP", "100", "2"}
			if len(keys) == 11 {
				values = append(values, "inside", "outside")
			}
			parsed, err := parseCSVApply(strings.NewReader(strings.Join(keys, ",")+"\n"+strings.Join(values, ",")+"\n"), "", defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			encode := func(value any) json.RawMessage {
				t.Helper()
				raw, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				return raw
			}
			baseline, err := materializeApprovedMapping(encode(candidate), parsed.SourceColumns)
			if err != nil {
				t.Fatal(err)
			}
			baselineBytes := marshalApprovedMapping(baseline)
			baselineFingerprint, err := mappingFingerprint(baseline, parsed.SourceContentSHA256)
			if err != nil {
				t.Fatal(err)
			}
			for i, key := range keys {
				t.Run(key, func(t *testing.T) {
					original := fields[i]["empty_value_policy"]
					defer func() { fields[i]["empty_value_policy"] = original }()
					for _, value := range []any{nil, "", " ", " profile_default ", "unknown", 1, false, []any{}, map[string]any{}} {
						// Map values deliberately retain JSON null and the empty string.
						fields[i]["empty_value_policy"] = value
						_, err := materializeApprovedMapping(encode(candidate), parsed.SourceColumns)
						var conflict *mappingValidationError
						if !errors.As(err, &conflict) || conflict.Code != "network_flow_mapping_conflict" || conflict.ReasonCode != "invalid_empty_value_policy" || conflict.FieldKey != key {
							t.Errorf("explicit %s: expected field-specific invalid policy, got %v", encode(value), err)
						}
					}
					for _, form := range []string{"omitted", "profile_default", "explicit"} {
						switch form {
						case "omitted":
							delete(fields[i], "empty_value_policy")
						case "profile_default":
							fields[i]["empty_value_policy"] = "profile_default"
						case "explicit":
							fields[i]["empty_value_policy"] = original
						}
						got, err := materializeApprovedMapping(encode(candidate), parsed.SourceColumns)
						if err != nil {
							t.Fatalf("%s: %v", form, err)
						}
						fingerprint, err := mappingFingerprint(got, parsed.SourceContentSHA256)
						if err != nil || fingerprint != baselineFingerprint || !bytes.Equal(marshalApprovedMapping(got), baselineBytes) {
							t.Fatalf("%s changed approved bytes or fingerprint: %v", form, err)
						}
					}
					var approved map[string]any
					if err := json.Unmarshal(baselineBytes, &approved); err != nil {
						t.Fatal(err)
					}
					approvedField := approved["field_mappings"].([]any)[i].(map[string]any)
					for _, value := range []any{nil, "", " ", "profile_default", "unknown", 1, false, []any{}, map[string]any{}} {
						approvedField["empty_value_policy"] = value
						if _, err := decodeApprovedMapping(encode(approved), defaultLimits()); err == nil {
							t.Errorf("approved explicit %s admitted", encode(value))
						}
					}
					delete(approvedField, "empty_value_policy")
					if _, err := decodeApprovedMapping(encode(approved), defaultLimits()); err == nil {
						t.Error("approved omission admitted")
					}
				})
			}
			decoded, err := decodeApprovedMapping(baselineBytes, defaultLimits())
			if err != nil || !bytes.Equal(marshalApprovedMapping(decoded), baselineBytes) {
				t.Fatalf("valid approved round-trip changed: %v", err)
			}
			// Candidate policy handling must not hide the unavailable-kind error.
			for _, kind := range []string{"constant", "derived"} {
				fields[0]["mapping_kind"] = kind
				_, err := materializeApprovedMapping(encode(candidate), parsed.SourceColumns)
				var conflict *mappingValidationError
				if !errors.As(err, &conflict) || conflict.ReasonCode != "mapping_kind_unavailable" {
					t.Errorf("%s: expected unavailable mapping kind, got %v", kind, err)
				}
			}
		})
	}
}
