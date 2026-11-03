package database_migrations_test

import (
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"strings"
	"testing"
)

func TestGraphLeaseRetirementRejectsPopulatedState_Integration(t *testing.T) {
	for _, state := range []string{"empty", "live", "expired"} {
		t.Run(state, func(t *testing.T) {
			database := pgtest.Start(t).MigrationDatabaseThroughT(t, 35)
			db := database.SQL()
			if state != "empty" {
				if _, err := db.ExecContext(t.Context(), `INSERT INTO application_crypto_format(singleton,format_id) VALUES (true,'cartulary.application_crypto_format.v1') ON CONFLICT DO NOTHING;
INSERT INTO graph_projection_results(projection_result_id,graph_view_id,source_owner_id,source_snapshot_id,projection_schema_id,projection_version,normalized_configuration_sha256,normalized_source_sha256,canonical_output_sha256,vertex_count,edge_count,result_json,published_at)
VALUES ('gpres_' || repeat('a',64),'retirement-fixture','network_flow_activity','snapshot','graph_projection.v2','network_flow_activity.v1',repeat('1',64),repeat('2',64),repeat('3',64),0,0,'{}',now());`); err != nil {
					t.Fatal(err)
				}
				until := "2099-01-01T00:00:00Z"
				if state == "expired" {
					until = "2020-01-02T00:00:00Z"
				}
				if _, err := db.ExecContext(t.Context(), `INSERT INTO graph_projection_result_leases(lease_id,projection_result_id,lease_owner_id,lease_owner_resource_id,lease_purpose,leased_until,created_at,renewed_at) VALUES ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa','gpres_' || repeat('a',64),'snapshot_reporting','retirement-private-sentinel','render',$1,'2020-01-01T00:00:00Z','2020-01-01T00:00:00Z')`, until); err != nil {
					t.Fatal(err)
				}
			}
			var beforeLedger string
			if err := db.QueryRowContext(t.Context(), `SELECT jsonb_agg(to_jsonb(v) ORDER BY id)::text FROM goose_db_version v`).Scan(&beforeLedger); err != nil {
				t.Fatal(err)
			}
			err := database.ApplyThrough(t.Context(), 36)
			var present bool
			if scanErr := db.QueryRowContext(t.Context(), `SELECT to_regclass('public.graph_projection_result_leases') IS NOT NULL`).Scan(&present); scanErr != nil {
				t.Fatal(scanErr)
			}
			if state == "empty" {
				if err != nil || present {
					t.Fatalf("empty retirement failed: present=%t err=%v", present, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "graph result lease retirement requires an empty table") || strings.Contains(err.Error(), "retirement-private-sentinel") || !present {
				t.Fatalf("populated retirement did not reject safely: present=%t err=%v", present, err)
			}
			var afterLedger string
			var leases, results int
			if err := db.QueryRowContext(t.Context(), `SELECT (SELECT jsonb_agg(to_jsonb(v) ORDER BY id)::text FROM goose_db_version v),(SELECT count(*) FROM graph_projection_result_leases),(SELECT count(*) FROM graph_projection_results)`).Scan(&afterLedger, &leases, &results); err != nil {
				t.Fatal(err)
			}
			if beforeLedger != afterLedger || leases != 1 || results != 1 {
				t.Fatal("rejected retirement changed history or retained rows")
			}
		})
	}
}
