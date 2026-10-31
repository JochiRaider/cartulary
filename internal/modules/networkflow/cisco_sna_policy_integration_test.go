package networkflow_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestCiscoSNAFlowSearchPolicyAdmission(t *testing.T) {
	for _, index := range []int{0, 1, 2, 3, 6, 7, 8} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			assertCiscoPolicyRejection(t, "cisco_sna_flow_search_csv_v1", index, "empty_string_is_null")
		})
	}
}

func TestCiscoPolicyJSONPresenceAdmission(t *testing.T) {
	for _, profileID := range []string{"cisco_sna_netflow_csv_v1", "cisco_sna_flow_search_csv_v1"} {
		indices := []int{0, 4}
		if profileID == "cisco_sna_netflow_csv_v1" {
			indices = append(indices, 9)
		}
		for _, index := range indices {
			for _, policy := range []struct {
				name  string
				value any
			}{{"null", nil}, {"empty", ""}} {
				t.Run(fmt.Sprintf("%s/%d/%s", profileID, index, policy.name), func(t *testing.T) {
					assertCiscoPolicyRejection(t, profileID, index, policy.value)
				})
			}
		}
	}
}

func assertCiscoPolicyRejection(t *testing.T, profileID string, invalidIndex int, invalidPolicy any) {
	t.Helper()
	assertCiscoMappingRejection(t, profileID, invalidIndex, invalidPolicy, nil)
}

func assertCiscoMappingRejection(t *testing.T, profileID string, invalidIndex int, invalidPolicy any, mutate func(map[string]any) json.RawMessage) {
	t.Helper()
	headers := []string{"firstActiveTime", "lastActiveTime", "searchSubject.ipAddress", "peer.ipAddress", "searchSubject.portProtocol.port", "peer.portProtocol.port", "searchSubject.portProtocol.protocol", "connection.transferBytes", "connection.transferPackets"}
	keys := []string{"flow_start_utc", "flow_end_utc", "src_ip", "dst_ip", "src_port", "dst_port", "ip_protocol", "bytes_count", "packets_count"}
	transforms := []string{"timestamp_profile_v1", "timestamp_profile_v1", "ip_literal_v1", "ip_literal_v1", "port_number_v1", "port_number_v1", "protocol_number_or_token_v1", "uint64_decimal_string_v1", "uint64_decimal_string_v1"}
	stamp := "2026-10-29T00:00:00+0000"
	profile := map[string]any{"schema_id": "cartulary.network_flow.timestamp_profile.v1", "mode": "cisco_sna_datetime_v1", "precision": "milliseconds"}
	if profileID == "cisco_sna_netflow_csv_v1" {
		headers = []string{"Start Time", "End Time", "Source IP", "Destination IP", "Source Port", "Destination Port", "Protocol", "Bytes", "Packets", "Input Interface", "Output Interface"}
		keys = append(keys, "input_interface", "output_interface")
		transforms = append(transforms, "trim_ascii_space_v1", "trim_ascii_space_v1")
		stamp = "2026-10-29T00:00:00Z"
		profile = map[string]any{"schema_id": "cartulary.network_flow.timestamp_profile.v1", "mode": "rfc3339", "precision": "milliseconds", "timezone": nil, "timezone_ruleset_id": nil, "ambiguous_local_time_policy": "reject", "local_time_gap_policy": "reject"}
	}
	values := []string{stamp, stamp, "192.0.2.1", "198.51.100.2", "12345", "443", "TCP", "100", "2"}
	if len(keys) == 11 {
		values = append(values, "inside", "outside")
	}
	source := []byte(strings.Join(headers, ",") + "\n" + strings.Join(values, ",") + "\n")
	input := publicationFixtureCase{Name: "policy-admission", Filename: "policies.csv", ActorID: "00000000-0000-4000-8000-000000000001", IncidentID: "00000000-0000-4000-8000-000000000002", SessionID: "00000000-0000-4000-8000-000000000003", UnitID: "00000000-0000-4000-8000-000000000004", TableUUID: "00000000-0000-4000-8000-000000000005"}
	r := startPublicationFixture(t, input, source)
	path := "/api/v1/import-sessions/" + input.SessionID + "/units/" + input.UnitID
	inspection := r.request(http.MethodPost, path+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}, http.StatusOK)
	fields := []map[string]any{}
	for index, key := range keys {
		fields = append(fields, map[string]any{"mapping_kind": "source_column", "field_key": "network_flow." + key, "source_column_ordinal": index + 1, "transform_id": transforms[index], "combinability": "single_source_only"})
	}
	mapping := map[string]any{"target_kind": "network_flow_table", "target_table_schema_id": "cartulary.network_flow_table.v1", "source_profile_id": profileID, "timestamp_profile": profile, "field_mappings": fields}
	body := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v1", "owner_mapping": mapping}
	// The valid request proves the source and every other admission member
	// before the sole forbidden policy is introduced.
	valid := r.request(http.MethodPost, path+"/mapping-preview", body, http.StatusOK)
	owner := valid["owner_result"].(map[string]any)
	if owner["preview_accepted_count"] != float64(1) || owner["preview_rejected_count"] != float64(0) {
		t.Fatal("control source must have one valid, nonempty record")
	}
	fields[invalidIndex]["empty_value_policy"] = "profile_default"
	if suggested := r.request(http.MethodPost, path+"/mapping-preview", body, http.StatusOK); !reflect.DeepEqual(valid, suggested) {
		t.Fatal("profile_default changed preview or fingerprint")
	}
	delete(fields[invalidIndex], "empty_value_policy")
	state := func() any {
		return map[string]any{
			"owner":    r.durableState(),
			"sessions": r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(s) ORDER BY import_session_id),'[]') FROM import_sessions s`),
			"units":    r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(u) ORDER BY import_unit_id),'[]') FROM import_units u`),
			"receipts": r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY route_key,client_txn_id),'[]') FROM route_idempotency r`),
			"jobs":     r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(j) ORDER BY job_id),'[]') FROM jobs j`),
		}
	}
	before := state()
	wantOwner := map[string]any{"schema_id": "cartulary.network_flow.import_owner_error.v1", "owner_code": "network_flow_mapping_invalid", "retryable": false, "safe_details": map[string]any{"reason_code": "invalid_empty_value_policy", "field": "network_flow." + keys[invalidIndex]}}
	if mutate == nil {
		fields[invalidIndex]["empty_value_policy"] = invalidPolicy
	} else {
		// The mutation owns a separate decoded copy; the valid control stays intact.
		encoded, err := json.Marshal(mapping)
		if err != nil {
			t.Fatal(err)
		}
		var copy map[string]any
		if err := json.Unmarshal(encoded, &copy); err != nil {
			t.Fatal(err)
		}
		body["owner_mapping"] = mutate(copy)
		wantOwner["safe_details"] = map[string]any{"reason_code": "variant_member_conflict"}
	}
	checkRejected := func(method, endpoint string, request map[string]any) {
		t.Helper()
		failure := r.errorRequest(method, endpoint, request, http.StatusBadRequest, "invalid_import_request")
		details := failure["details"].(map[string]any)
		if details["field"] != "owner_mapping" || details["reason_code"] != "owner_preview_validation_failed" || !reflect.DeepEqual(details["owner_error"], wantOwner) || failure["retryable"] != false {
			t.Fatalf("unexpected policy rejection: %#v", failure)
		}
		if repeated := r.errorRequest(method, endpoint, request, http.StatusBadRequest, "invalid_import_request"); !reflect.DeepEqual(failure, repeated) {
			t.Fatal("repeated rejection changed error")
		}
		if !reflect.DeepEqual(before, state()) {
			t.Fatal("policy rejection changed durable state")
		}
	}
	checkRejected(http.MethodPost, path+"/mapping-preview", body)
	approval := map[string]any{}
	for key, value := range body {
		approval[key] = value
	}
	columns := []map[string]any{}
	for _, value := range inspection["owner_result"].(map[string]any)["source_columns"].([]any) {
		column := value.(map[string]any)
		columns = append(columns, map[string]any{"source_column_ordinal": column["source_column_ordinal"], "source_header_text": column["raw_header_text"], "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"})
	}
	approval["client_txn_id"], approval["header_row_ref"], approval["data_start_row_ref"], approval["source_columns"] = "policy-approval", 1, 2, columns
	checkRejected(http.MethodPut, path+"/mapping", approval)
	// Repair only the policy, keeping the rejected approval transaction ID.
	fields[invalidIndex]["empty_value_policy"] = "empty_string_is_invalid"
	if keys[invalidIndex] == "input_interface" || keys[invalidIndex] == "output_interface" || profileID == "cisco_sna_flow_search_csv_v1" && (keys[invalidIndex] == "src_port" || keys[invalidIndex] == "dst_port") {
		fields[invalidIndex]["empty_value_policy"] = "empty_string_is_null"
	}
	body["owner_mapping"], approval["owner_mapping"] = mapping, mapping
	corrected := r.request(http.MethodPost, path+"/mapping-preview", body, http.StatusOK)
	if !reflect.DeepEqual(valid, corrected) {
		t.Fatal("equivalent explicit default changed preview or fingerprint")
	}
	accepted := r.request(http.MethodPut, path+"/mapping", approval, http.StatusOK)
	after := state()
	if replay := r.request(http.MethodPut, path+"/mapping", approval, http.StatusOK); !reflect.DeepEqual(accepted, replay) || !reflect.DeepEqual(after, state()) {
		t.Fatal("corrected approval replay changed state or response")
	}
	var stored map[string]any
	value := r.jsonQuery(`SELECT approved_mapping_json FROM import_units WHERE import_unit_id=$1`, input.UnitID)
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &stored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored["owner_mapping"], owner["materialized_mapping"]) {
		t.Fatal("stored owner mapping differs from valid preview")
	}
}
