package networkflow_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestNetworkFlowCanonicalEndpointStorage_Integration(t *testing.T) {
	h, actor, incident := startNetworkFlowStoreTest(t, "canonical-endpoints")
	AssertCanonicalEndpointSQLVectors(t, h.DB)
	store := newTestNetworkFlowStore(t, h.DB, revisionsupport.MustAppender(t))
	table := createTestTable(t, h.DB, store, actor.ID, incident, "canonical.csv", nil, 1)
	for _, mutation := range []map[string]any{
		{"endpoint_b_ip": "2001:0db8::1"},
		{"endpoint_b_ip": "2001:db8::1/128"},
		{"endpoint_a_ip": "192.0.2.10/32"},
		{"endpoint_a_ip": "2001:db8::2"},
		{"endpoint_a_port": -1},
		{"endpoint_b_port": 65536},
	} {
		mutation["network_flow_row_id"] = "nfr_" + strings.Repeat("f", 64)
		mutation["source_row_number"] = 2
		raw, err := json.Marshal(mutation)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := h.DB.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(t.Context(), `INSERT INTO network_flow_rows SELECT (jsonb_populate_record(NULL::network_flow_rows,to_jsonb(r)||$2::jsonb)).* FROM network_flow_rows r WHERE network_flow_table_id=$1`, table.TableID, raw)
		var failure *pgconn.PgError
		if !errors.As(err, &failure) || failure.Code != "23514" || failure.ConstraintName != "network_flow_rows_endpoints_ck" {
			_ = tx.Rollback(t.Context())
			t.Fatalf("invalid insert escaped endpoint invariant: %v", err)
		}
		if err := tx.Rollback(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := h.DB.QueryRow(t.Context(), `SELECT count(*) FROM network_flow_rows WHERE network_flow_table_id=$1`, table.TableID).Scan(&count); err != nil || count != 1 {
		t.Fatal("rejected insert changed state", count, err)
	}
	if err := ValidateExtensionState(t.Context(), newExtensionStateV3Reader(h.DB)); err != nil {
		t.Fatal("valid endpoints failed claim", err)
	}
}
