package networkflow_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestNetworkFlowDirectionEvidenceStorage_Integration(t *testing.T) {
	h, actor, incident := startNetworkFlowStoreTest(t, "direction-evidence")
	store := newTestNetworkFlowStore(t, h.DB, revisionsupport.MustAppender(t))
	table := createTestTable(t, h.DB, store, actor.ID, incident, "evidence.csv", nil, 1)
	for _, tc := range []struct {
		name, expression string
		valid            bool
	}{
		{"current evidence", "direction_evidence", true},
		{"SQL null", "NULL::jsonb", false},
		{"JSON null", "'null'::jsonb", false},
		{"array root", "'[]'::jsonb", false},
		{"extra root member", `direction_evidence || '{"extra":true}'::jsonb`, false},
		{"missing root member", "direction_evidence - 'first_role'", false},
		{"scalar input", `jsonb_set(direction_evidence,'{inputs,0}','0')`, false},
		{"extra input member", `jsonb_set(direction_evidence,'{inputs,0}',(direction_evidence#>'{inputs,0}') || '{"extra":true}')`, false},
		{"missing input member", `jsonb_set(direction_evidence,'{inputs,0}',(direction_evidence#>'{inputs,0}') - 'raw_header_sha256')`, false},
		{"missing input", `jsonb_set(direction_evidence,'{inputs}',(direction_evidence->'inputs') - 0)`, false},
		{"duplicate ordinal", `jsonb_set(direction_evidence,'{inputs,1,source_column_ordinal}',direction_evidence#>'{inputs,0,source_column_ordinal}')`, false},
		{"mistyped ordinal", `jsonb_set(direction_evidence,'{inputs,0,source_column_ordinal}','true')`, false},
		{"ordinal outside int64", `jsonb_set(direction_evidence,'{inputs,0,source_column_ordinal}','9223372036854775808')`, false},
		{"wrong input key", `jsonb_set(direction_evidence,'{inputs,0,input_key}','"network_flow.input_second_ip"')`, false},
		{"invalid hash", `jsonb_set(direction_evidence,'{inputs,0,raw_header_sha256}','"invalid"')`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var valid bool
			// Expressions are fixed test cases; the retained table binding remains a parameter.
			err := h.DB.QueryRow(t.Context(), `SELECT public.network_flow_direction_evidence_is_valid_v1(`+tc.expression+`,direction_kind,from_endpoint,direction_reason,false) FROM network_flow_rows WHERE network_flow_table_id=$1`, table.TableID).Scan(&valid)
			if err != nil || valid != tc.valid {
				t.Fatalf("closed evidence admission=%v want=%v: %v", valid, tc.valid, err)
			}
		})
	}
}
