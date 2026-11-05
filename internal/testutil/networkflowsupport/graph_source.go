package networkflowsupport

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// SeedGraphSource installs immutable source fixtures for process recovery tests.
// Commands and materialization under test still run through the public server.
func SeedGraphSource(t testing.TB, db *sql.DB, incident, actor string) string {
	t.Helper()
	session, unit := uuid.NewString(), uuid.NewString()
	table := "nft_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	digest := strings.Repeat("a", 64)
	fields := []string{"flow_start_utc", "flow_end_utc", "input_first_ip", "input_second_ip", "input_first_port", "input_second_port", "ip_protocol", "bytes_count", "packets_count"}
	transforms := []string{"timestamp_profile_v1", "timestamp_profile_v1", "ip_literal_v1", "ip_literal_v1", "port_number_v1", "port_number_v1", "protocol_number_or_token_v1", "uint64_decimal_string_v1", "uint64_decimal_string_v1"}
	columns, mappings := []any{}, []any{}
	for i, field := range fields {
		key := "network_flow." + field
		columns = append(columns, map[string]any{"source_column_ordinal": i + 1, "raw_header_text": key, "raw_header_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(key))), "normalized_header_for_suggestion": key, "detected_empty_count": 0, "sample_values": []any{}})
		mappings = append(mappings, map[string]any{"mapping_kind": "source_column", "field_key": key, "source_column_ordinal": i + 1, "transform_id": transforms[i], "empty_value_policy": "empty_string_is_invalid", "combinability": "single_source_only"})
	}
	mappings = append(mappings, map[string]any{"mapping_kind": "system_derivation", "field_key": "network_flow.observation_source_ref", "derivation_id": "network_flow.observation_source_ref.v1", "combinability": "single_source_only"})
	owner := map[string]any{"endpoint_derivation_id": "declared_source_destination_v1", "target_kind": "network_flow_table", "target_table_schema_id": "cartulary.network_flow_table.v2", "source_profile_id": "cisco_sna_netflow_csv_v2", "parser_profile_id": "rfc4180_headered_csv_v1", "unknown_column_policy": "preserve_unmapped_raw", "timestamp_profile": map[string]any{"schema_id": "cartulary.network_flow.timestamp_profile.v1", "mode": "rfc3339", "precision": "microseconds", "timezone": nil, "timezone_ruleset_id": nil, "ambiguous_local_time_policy": "reject", "local_time_gap_policy": "reject"}, "source_columns": columns, "field_mappings": mappings}
	approved, _ := json.Marshal(map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.approved_mapping.v2", "owner_mapping": owner, "source_columns": []any{}})
	inputs := []any{}
	for _, i := range []int{2, 4, 3, 5} {
		inputs = append(inputs, map[string]any{"input_key": "network_flow." + fields[i], "source_column_ordinal": i + 1, "raw_header_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte("network_flow."+fields[i]))), "raw_value_sha256": digest})
	}
	evidence, _ := json.Marshal(map[string]any{"derivation_id": "declared_source_destination_v1", "first_endpoint": "a", "first_role": "", "second_role": "", "inputs": inputs})
	statements := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO import_sessions(import_session_id,incident_id,created_by_user_id,client_txn_id,assistant_profile,source_file_kind,original_filename,source_content_sha256,source_media_type,source_byte_size,parser_profile_id,parser_version,session_status) VALUES($1,$2,$3,$4,'network_flow_test','csv','process.csv',$5,'text/csv',12,'rfc4180_headered_csv_v1','test','ready_to_apply')`, []any{session, incident, actor, "source-" + unit, digest}},
		{`INSERT INTO import_units(import_unit_id,import_session_id,unit_status,locator_kind,locator,source_rect_a1,header_row_ref,data_start_row_ref,inferred_row_count,inferred_column_count,warning_codes,mapping_fingerprint,approved_mapping_json,columns_json,source_rows_json,preview_rows_json,approved_target_kind,approved_extension_profile_id,discovery_sequence) VALUES($1,$2,'ready','csv','unit-1','A1:Z2',1,2,1,9,'{}',$3,$4,'[]','[]','[]','network_flow_table','network_flow_activity',1)`, []any{unit, session, digest, approved}},
		{`INSERT INTO network_flow_tables(network_flow_table_id,incident_id,display_name,table_status,source_import_session_id,source_import_unit_id,source_content_sha256,source_filename_display,source_filename_digest,source_filename_digest_key_id,mapping_fingerprint,source_profile_id,parser_profile_id,row_count_accepted,row_count_rejected,created_by_user_id) VALUES($1,$2,'process source','active',$3,$4,$5,'process.csv',$5,'test-key',$5,'cisco_sna_netflow_csv_v2','rfc4180_headered_csv_v1',1,0,$6)`, []any{table, incident, session, unit, digest, actor}},
		{`INSERT INTO network_flow_rows(network_flow_row_id,network_flow_table_id,incident_id,source_row_number,source_row_digest_sha256,normalized_row_digest_sha256,mapping_fingerprint,flow_start_utc,flow_end_utc,endpoint_a_ip,endpoint_b_ip,direction_kind,from_endpoint,direction_reason,direction_evidence,ip_protocol,bytes_count,packets_count,observation_source_ref,created_by_user_id) VALUES($1,$2,$3,1,$4,$4,$4,now(),now(),'192.0.2.1','192.0.2.2','source_to_destination','a','',$6,6,'12','1','{}',$5)`, []any{"nfr_" + digest, table, incident, digest, actor, evidence}},
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(context.Background(), statement.sql, statement.args...); err != nil {
			t.Fatalf("seed graph process source: %v", err)
		}
	}
	return table
}
