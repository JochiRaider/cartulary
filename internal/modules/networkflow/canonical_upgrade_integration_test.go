package networkflow_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/stdlib"
)

type canonicalUpgradeVector struct {
	Seed []struct {
		Table string                     `json:"table"`
		Row   map[string]json.RawMessage `json:"row"`
	} `json:"seed"`
	Expected struct {
		SourceRefs json.RawMessage `json:"source_refs"`
	} `json:"expected"`
}

// The shipped physical migration must preserve immutable results and opaque
// receipts on success and rollback. This does not establish current serving
// compatibility of a database stopped at an earlier schema version.
func TestNetworkFlowCanonicalMigrationPreservesImmutableState_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	for _, reject := range []bool{true, false} {
		t.Run(fmt.Sprintf("reject_%t", reject), func(t *testing.T) {
			stage := "seed"
			root, err := networkflowfixtures.SelectedRunRoot()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				summary, _ := json.Marshal(map[string]any{"scenario": t.Name(), "last_stage": stage, "passed": !t.Failed()})
				if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("canonical-upgrade-%t.json", reject)), summary, 0600); err != nil {
					t.Error(err)
				}
			})
			database := harness.MigrationDatabaseThroughT(t, 33)
			db := database.SQL()
			raw, err := os.ReadFile("testdata/canonical_filter_order_migration.json")
			if err != nil {
				t.Fatal(err)
			}
			var vector canonicalUpgradeVector
			if err := json.Unmarshal(raw, &vector); err != nil {
				t.Fatal(err)
			}
			for _, entry := range vector.Seed {
				if entry.Table == "users" {
					flowtest.SeedActor(t, db, flowtest.ActorAttributes{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Email: "canonical-migration@example.invalid", DisplayName: "Migration fixture"})
					continue
				}
				columns := make([]string, 0, len(entry.Row))
				for column := range entry.Row {
					columns = append(columns, `"`+strings.ReplaceAll(column, `"`, `""`)+`"`)
				}
				sort.Strings(columns)
				table := `public."` + strings.ReplaceAll(entry.Table, `"`, `""`) + `"`
				names := strings.Join(columns, ",")
				upgradeExec(t, db, "INSERT INTO "+table+" ("+names+") SELECT "+names+" FROM jsonb_populate_record(NULL::"+table+",$1::jsonb)", string(mustJSON(t, entry.Row)))
			}
			// A retained application has a crypto identity in addition to DDL history.
			upgradeExec(t, db, `INSERT INTO application_crypto_format(singleton,format_id) VALUES (true,'cartulary.application_crypto_format.v1')`)
			upgradeExec(t, db, `INSERT INTO extension_state_metadata(profile_id,migration_lineage_id,state_version,metadata_version,created_at,updated_at) VALUES ('network_flow_activity','network_flow_activity.state_v1',5,1,now(),now())`)
			historicalLink := seedCanonicalHistoricalLink(t, db)
			const graphID = "nfgv_33333333333333333333333333333333"
			immutableID := seedCanonicalUpgradeProjection(t, db, graphID)
			immutableBefore := upgradeImmutableState(t, db, immutableID)
			stage = "migrate"
			if reject {
				upgradeExec(t, db, `UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,mapping_fingerprint}',to_jsonb(repeat('f',64)))`)
				tables := []string{"goose_db_version", "extension_state_metadata", "network_flow_tables", "network_flow_rows", "network_flow_rejected_row_diagnostics", "network_flow_indicator_bindings", "network_flow_graph_views", "jobs", "extension_job_commit_proofs", "route_idempotency", "import_units", "import_sessions"}
				before := upgradeTables(t, db, tables)
				if err := database.ApplyThrough(t.Context(), 34); err == nil {
					t.Fatal("invalid binding passed combined upgrade")
				}
				if !reflect.DeepEqual(before, upgradeTables(t, db, tables)) || !reflect.DeepEqual(immutableBefore, upgradeImmutableState(t, db, immutableID)) {
					t.Fatal("failed upgrade changed retained state")
				}
				stage = "rollback_verified"
				return
			}
			if err := database.ApplyThrough(t.Context(), 34); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(immutableBefore, upgradeImmutableState(t, db, immutableID)) {
				t.Fatal("migration changed immutable projection")
			}
			if historicalLink != canonicalHistoricalLinkBytes(t, db) {
				t.Fatal("migration changed historical link receipt")
			}
			var refs json.RawMessage
			if err := db.QueryRowContext(t.Context(), "SELECT source_row_refs FROM network_flow_indicator_bindings").Scan(&refs); err != nil {
				t.Fatal(err)
			}
			requireUpgradeJSON(t, vector.Expected.SourceRefs, refs)

			stage = "complete"
		})
	}
}

// Historical link receipts are opaque preservation fixtures, not current
// admission or replay fixtures.
func seedCanonicalHistoricalLink(t *testing.T, db *sql.DB) string {
	t.Helper()
	raw, err := os.ReadFile("testdata/canonical_historical_link_receipt.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Hash     string          `json:"request_sha256"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	upgradeExec(t, db, `INSERT INTO route_idempotency(route_key,scope_key,client_txn_id,actor_user_id,request_hash,status_code,response_json) VALUES ('nf.indicator_links.create','22222222-2222-4222-8222-222222222222','historical-link','11111111-1111-4111-8111-111111111111',decode($1,'hex'),201,$2)`, fixture.Hash, string(fixture.Response))
	return canonicalHistoricalLinkBytes(t, db)
}

func canonicalHistoricalLinkBytes(t *testing.T, db *sql.DB) string {
	t.Helper()
	var raw string
	if err := db.QueryRowContext(t.Context(), `SELECT to_jsonb(r)::text FROM route_idempotency r WHERE route_key='nf.indicator_links.create' AND client_txn_id='historical-link'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func upgradeExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}
func upgradeTables(t *testing.T, db *sql.DB, tables []string) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, table := range tables {
		var raw string
		if err := db.QueryRowContext(t.Context(), "SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]')::text FROM "+table+" v").Scan(&raw); err != nil {
			t.Fatal(err)
		}
		result[table] = raw
	}
	return result
}
func upgradeImmutableState(t *testing.T, db *sql.DB, id string) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, table := range []string{"graph_projection_results", "graph_projection_result_vertices", "graph_projection_result_edges"} {
		var raw string
		if err := db.QueryRowContext(t.Context(), "SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]')::text FROM "+table+" v WHERE projection_result_id=$1", id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		result[table] = raw
	}
	return result
}
func requireUpgradeJSON(t *testing.T, want, got []byte) {
	t.Helper()
	var a, b any
	if err := json.Unmarshal(want, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(got, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("upgrade JSON mismatch: expected %s got %s", want, got)
	}
}

func seedCanonicalUpgradeProjection(t *testing.T, db *sql.DB, graphID string) string {
	t.Helper()
	input, err := os.ReadFile("testdata/canonical_upgrade_projection.json")
	if err != nil {
		t.Fatal(err)
	}
	projected, err := graphprojection.ProjectV2(t.Context(), graphprojection.InvocationContextV2{GraphViewID: graphID, SourceOwnerID: ProfileID}, input)
	if err != nil {
		t.Fatal(err)
	}
	completed, err := projected.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	completed.PublishedAt = time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	connection, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	err = connection.Raw(func(driver any) error {
		tx, err := driver.(*stdlib.Conn).Conn().Begin(t.Context())
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		publisher, err := postgresresult.NewPublisher(tx)
		if err != nil {
			return err
		}
		if err = publisher.PublishResult(t.Context(), completed); err != nil {
			return err
		}
		return tx.Commit(t.Context())
	})
	closeErr := connection.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	b := completed.Binding
	upgradeExec(t, db, `UPDATE network_flow_graph_views SET selected_projection_result_id=$1,selected_source_snapshot_id=$2,selected_projection_schema_id=$3,selected_projection_version=$4,selected_normalized_configuration_sha256=$5,selected_normalized_source_sha256=$6,selected_canonical_output_sha256=$7 WHERE graph_view_id=$8`, b.ProjectionResultID, b.SourceSnapshotID, b.ProjectionSchemaID, b.ProjectionVersion, b.NormalizedConfigurationSHA256, b.NormalizedSourceSHA256, b.CanonicalOutputSHA256, graphID)
	return b.ProjectionResultID
}
