package networkflow

import (
	"encoding/json"
	"strings"
	"testing"
)

func testRetainedImportResultAdmission(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	table := "nft_" + strings.Repeat("a", 32)
	valid := map[string]any{
		"schema_id": "cartulary.network_flow.import_unit_result.v2", "import_session_id": id, "import_unit_id": id,
		"source_content_sha256": strings.Repeat("b", 64), "mapping_fingerprint": strings.Repeat("c", 64),
		"source_profile_id": "cisco_sna_flow_search_csv_v2", "parser_profile_id": "rfc4180_headered_csv_v1",
		"table_ref": map[string]any{"kind": "network_flow_table", "id": table, "route": "/api/v1/incidents/" + id + "/network-flow/tables/" + table,
			"display_name": "Flows", "row_count_accepted": 10, "row_count_rejected": 0, "diagnostics_truncated": false, "table_version": 1},
	}
	raw, err := json.Marshal(valid)
	if err != nil || validateRetainedImportResult(raw) != nil {
		t.Fatal("current retained result rejected", err)
	}
	for _, change := range []string{"old schema", "old profile", "parser", "hash", "missing", "extra", "null", "nested extra", "route", "negative", "fraction", "version"} {
		t.Run(change, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			ref := value["table_ref"].(map[string]any)
			switch change {
			case "old schema":
				value["schema_id"] = "cartulary.network_flow.import_unit_result.v1"
			case "old profile":
				value["source_profile_id"] = "cisco_sna_flow_search_csv_v1"
			case "parser":
				value["parser_profile_id"] = "unknown"
			case "hash":
				value["mapping_fingerprint"] = "invalid"
			case "missing":
				delete(value, "import_unit_id")
			case "extra":
				value["extra"] = true
			case "null":
				value["import_session_id"] = nil
			case "nested extra":
				ref["extra"] = true
			case "route":
				ref["route"] = "/external/" + table
			case "negative":
				ref["row_count_accepted"] = -1
			case "fraction":
				ref["row_count_rejected"] = 0.5
			case "version":
				ref["table_version"] = 0
			}
			invalid, err := json.Marshal(value)
			if err != nil || validateRetainedImportResult(invalid) == nil {
				t.Fatal("malformed retained result admitted", err)
			}
		})
	}
}
