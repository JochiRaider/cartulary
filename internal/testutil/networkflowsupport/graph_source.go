package networkflowsupport

import (
	"context"
	"database/sql"
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
	statements := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO import_sessions(import_session_id,incident_id,created_by_user_id,client_txn_id,assistant_profile,source_file_kind,original_filename,source_content_sha256,source_media_type,source_byte_size,parser_profile_id,parser_version,session_status) VALUES($1,$2,$3,$4,'network_flow_test','csv','process.csv',$5,'text/csv',12,'rfc4180_headered_csv_v1','test','ready_to_apply')`, []any{session, incident, actor, "source-" + unit, digest}},
		{`INSERT INTO import_units(import_unit_id,import_session_id,unit_status,locator_kind,locator,source_rect_a1,header_row_ref,data_start_row_ref,inferred_row_count,inferred_column_count,warning_codes,mapping_fingerprint,approved_mapping_json,columns_json,source_rows_json,preview_rows_json,approved_target_kind,approved_extension_profile_id,discovery_sequence) VALUES($1,$2,'ready','csv','unit-1','A1:Z2',1,2,1,9,'{}',$3,'{}','[]','[]','[]','network_flow_table','network_flow_activity',1)`, []any{unit, session, digest}},
		{`INSERT INTO network_flow_tables(network_flow_table_id,incident_id,display_name,table_status,source_import_session_id,source_import_unit_id,source_content_sha256,source_filename_display,source_filename_digest,source_filename_digest_key_id,mapping_fingerprint,source_profile_id,parser_profile_id,row_count_accepted,row_count_rejected,created_by_user_id) VALUES($1,$2,'process source','active',$3,$4,$5,'process.csv',$5,'test-key',$5,'cisco_sna_netflow_csv_v1','rfc4180_headered_csv_v1',1,0,$6)`, []any{table, incident, session, unit, digest, actor}},
		{`INSERT INTO network_flow_rows(network_flow_row_id,network_flow_table_id,incident_id,source_row_number,source_row_digest_sha256,normalized_row_digest_sha256,mapping_fingerprint,flow_start_utc,flow_end_utc,src_ip,dst_ip,ip_protocol,bytes_count,packets_count,observation_source_ref,created_by_user_id) VALUES($1,$2,$3,1,$4,$4,$4,now(),now(),'192.0.2.1','192.0.2.2',6,'12','1','{}',$5)`, []any{"nfr_" + digest, table, incident, digest, actor}},
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(context.Background(), statement.sql, statement.args...); err != nil {
			t.Fatalf("seed graph process source: %v", err)
		}
	}
	return table
}
