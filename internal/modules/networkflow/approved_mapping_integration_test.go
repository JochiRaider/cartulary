package networkflow_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type approvedApplyAdmissionCase struct {
	name  string
	apply func(json.RawMessage) (json.RawMessage, error)
}

func approvedApplyAdmissionCases() []approvedApplyAdmissionCase {
	var cases []approvedApplyAdmissionCase
	add := func(name string, change func(map[string]any) error) {
		cases = append(cases, approvedApplyAdmissionCase{name, func(raw json.RawMessage) (json.RawMessage, error) {
			var mapping map[string]any
			if err := json.Unmarshal(raw, &mapping); err != nil {
				return nil, err
			}
			if err := change(mapping); err != nil {
				return nil, err
			}
			return json.Marshal(mapping)
		}})
	}
	for _, kind := range []string{"ignored_source_column", "system_derivation"} {
		for _, policy := range []struct {
			name  string
			value any
		}{{"null", nil}, {"empty", ""}, {"default", "profile_default"}} {
			add(kind+"/"+policy.name, func(mapping map[string]any) error {
				for _, item := range mapping["field_mappings"].([]any) {
					field := item.(map[string]any)
					if field["mapping_kind"] == kind {
						field["empty_value_policy"] = policy.value
						return nil
					}
				}
				return fmt.Errorf("positive control has no %s row", kind)
			})
		}
	}
	for _, name := range []string{"root_alias", "timestamp_alias", "row_alias", "column_alias", "sample_alias", "missing_target", "null_display_name", "null_samples", "empty_precision"} {
		add(name, func(mapping map[string]any) error {
			column := mapping["source_columns"].([]any)[0].(map[string]any)
			switch name {
			case "root_alias":
				mapping["FIELD_MAPPINGS"] = mapping["field_mappings"]
			case "timestamp_alias":
				profile := mapping["timestamp_profile"].(map[string]any)
				profile["PRECISION"] = profile["precision"]
			case "row_alias":
				field := mapping["field_mappings"].([]any)[0].(map[string]any)
				field["EMPTY_VALUE_POLICY"] = field["empty_value_policy"]
			case "column_alias":
				column["RAW_HEADER_TEXT"] = column["raw_header_text"]
			case "sample_alias":
				sample := column["sample_values"].([]any)[0].(map[string]any)
				sample["SAFE_SAMPLE"] = sample["safe_sample"]
			case "missing_target":
				delete(mapping, "target_kind")
			case "null_display_name":
				mapping["display_name_override"] = nil
			case "null_samples":
				column["sample_values"] = nil
			case "empty_precision":
				mapping["timestamp_profile"].(map[string]any)["precision"] = ""
			}
			return nil
		})
	}
	for _, position := range []string{"first", "middle", "last"} {
		for _, direction := range []string{"shorter", "longer"} {
			add("sample_count/"+position+"/"+direction, func(mapping map[string]any) error {
				columns := mapping["source_columns"].([]any)
				index := map[string]int{"first": 0, "middle": len(columns) / 2, "last": len(columns) - 1}[position]
				column := columns[index].(map[string]any)
				samples := column["sample_values"].([]any)
				// This fixture has one nonempty value per column, so either edit
				// preserves its valid zero detected_empty_count.
				if len(samples) != 1 || column["detected_empty_count"] != float64(0) {
					return fmt.Errorf("invalid sample-count positive control")
				}
				if direction == "shorter" {
					column["sample_values"] = samples[:0]
				} else {
					column["sample_values"] = append(samples, samples[0])
				}
				return nil
			})
		}
	}
	return cases
}

func TestApprovedMappingApplyAdmission(t *testing.T) {
	for _, profileID := range []string{"cisco_sna_netflow_csv_v1", "cisco_sna_flow_search_csv_v1"} {
		for _, mutation := range approvedApplyAdmissionCases() {
			t.Run(profileID+"/"+mutation.name, func(t *testing.T) {
				keys := []string{"flow_start_utc", "flow_end_utc", "src_ip", "dst_ip", "src_port", "dst_port", "ip_protocol", "bytes_count", "packets_count"}
				transforms := []string{"timestamp_profile_v1", "timestamp_profile_v1", "ip_literal_v1", "ip_literal_v1", "port_number_v1", "port_number_v1", "protocol_number_or_token_v1", "uint64_decimal_string_v1", "uint64_decimal_string_v1"}
				stamp := "2026-10-29T00:00:00Z"
				timestamp := map[string]any{"schema_id": "cartulary.network_flow.timestamp_profile.v1", "mode": "rfc3339", "precision": "milliseconds", "timezone": nil, "timezone_ruleset_id": nil, "ambiguous_local_time_policy": "reject", "local_time_gap_policy": "reject"}
				if profileID == "cisco_sna_flow_search_csv_v1" {
					stamp = "2026-10-29T00:00:00+0000"
					timestamp = map[string]any{"schema_id": "cartulary.network_flow.timestamp_profile.v1", "mode": "cisco_sna_datetime_v1", "precision": "milliseconds"}
				}
				values := []string{stamp, stamp, "192.0.2.1", "198.51.100.2", "12345", "443", "TCP", "100", "2", "ignored"}
				source := []byte(strings.Join(keys, ",") + ",ignored\n" + strings.Join(values, ",") + "\n")
				input := publicationFixtureCase{Name: "approved-policy", Filename: "policies.csv", ActorID: "00000000-0000-4000-8000-000000000001", IncidentID: "00000000-0000-4000-8000-000000000002", SessionID: "00000000-0000-4000-8000-000000000003", UnitID: "00000000-0000-4000-8000-000000000004", TableUUID: "00000000-0000-4000-8000-000000000005"}
				r := startPublicationFixture(t, input, source)
				path := "/api/v1/import-sessions/" + input.SessionID + "/units/" + input.UnitID
				fields := []map[string]any{}
				for i, key := range keys {
					fields = append(fields, map[string]any{"mapping_kind": "source_column", "field_key": "network_flow." + key, "source_column_ordinal": i + 1, "transform_id": transforms[i], "combinability": "single_source_only"})
				}
				fields = append(fields, map[string]any{"mapping_kind": "ignored_source_column", "source_column_ordinal": 10, "ignore_reason": "user_ignored"})
				mapping := map[string]any{"target_kind": "network_flow_table", "target_table_schema_id": "cartulary.network_flow_table.v1", "source_profile_id": profileID, "timestamp_profile": timestamp, "field_mappings": fields}
				body := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v1", "owner_mapping": mapping}
				preview := r.request(http.MethodPost, path+"/mapping-preview", body, http.StatusOK)["owner_result"].(map[string]any)
				if preview["preview_accepted_count"] != float64(1) || preview["preview_rejected_count"] != float64(0) {
					t.Fatal("invalid positive control")
				}
				columns := []map[string]any{}
				for _, item := range preview["source_columns"].([]any) {
					column := item.(map[string]any)
					columns = append(columns, map[string]any{"source_column_ordinal": column["source_column_ordinal"], "source_header_text": column["raw_header_text"], "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"})
				}
				body["client_txn_id"], body["header_row_ref"], body["data_start_row_ref"], body["source_columns"] = "approved-policy", 1, 2, columns
				r.request(http.MethodPut, path+"/mapping", body, http.StatusOK)
				r.request(http.MethodPost, path+"/select", map[string]any{"client_txn_id": "select-policy"}, http.StatusOK)
				before := r.durableState()
				r.source.mu.Lock()
				r.source.active = &publicationSourceChangeCase{Kind: "approved_mapping"}
				r.source.changeApprovedMapping = mutation.apply
				r.source.mu.Unlock()
				applyPath := "/api/v1/import-sessions/" + input.SessionID + "/apply"
				applyBody := map[string]any{"client_txn_id": "apply-policy"}
				job := r.request(http.MethodPost, applyPath, applyBody, http.StatusAccepted)
				jobID := job["job_id"].(string)
				terminal := r.waitJob(jobID)
				after := r.durableState()
				if terminal["status"] != "failed" {
					t.Errorf("malformed approved mapping applied: %#v", terminal)
				}
				for _, member := range []string{"tables", "rows", "diagnostics", "audit"} {
					if !reflect.DeepEqual(before[member], after[member]) {
						t.Errorf("rejected apply changed %s", member)
					}
				}
				outcomes := after["outcomes"].([]any)
				if len(outcomes) != 1 {
					t.Fatalf("expected one failed unit outcome, got %#v", outcomes)
				}
				outcome := outcomes[0].(map[string]any)
				details, ok := outcome["error_details"].(map[string]any)
				if !ok {
					t.Fatalf("missing owner error: %#v", outcome)
				}
				owner, ok := details["owner_error"].(map[string]any)
				if !ok {
					t.Fatalf("missing owner error translation: %#v", details)
				}
				safe := owner["safe_details"].(map[string]any)
				if outcome["error_code"] != "import_apply_blocked" || outcome["error_retryable"] != false || owner["owner_code"] != "network_flow_mapping_invalid" || safe["reason_code"] != "variant_member_conflict" {
					t.Errorf("unexpected rejection: %#v", outcome)
				}
				r.source.mu.Lock()
				mutated, opens := r.source.requestMutated, r.source.opens
				r.source.mu.Unlock()
				if !mutated || opens != 0 {
					t.Errorf("apply did not reject injected mapping before source access: mutated=%v opens=%d", mutated, opens)
				}
				if replay := r.request(http.MethodPost, applyPath, applyBody, http.StatusAccepted); !reflect.DeepEqual(job, replay) || !reflect.DeepEqual(terminal, r.waitJob(jobID)) || !reflect.DeepEqual(after, r.durableState()) {
					t.Fatal("failed apply replay changed response or state")
				}
			})
		}
	}
}
