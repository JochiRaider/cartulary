package networkflow_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
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
		SourceRefs   json.RawMessage `json:"source_refs"`
		Filters      json.RawMessage `json:"filters"`
		SemanticHash string          `json:"semantic_hash"`
	} `json:"expected"`
}

// The same database travels through historical SQL, migration, application
// admission and public operations. No serving path is replaced with a test fake.
func TestNetworkFlowCanonicalUpgradeServing(t *testing.T) {
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
				t.Fatal("migration changed immutable projection or lease")
			}
			if historicalLink != canonicalHistoricalLinkBytes(t, db) {
				t.Fatal("migration changed historical link receipt")
			}
			var refs json.RawMessage
			if err := db.QueryRowContext(t.Context(), "SELECT source_row_refs FROM network_flow_indicator_bindings").Scan(&refs); err != nil {
				t.Fatal(err)
			}
			requireUpgradeJSON(t, vector.Expected.SourceRefs, refs)
			borrowed, err := database.RuntimeDatabase()
			if err != nil {
				t.Fatal(err)
			}
			stage = "startup"
			h := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "nf-canonical-upgrade", Database: borrowed, TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{
				"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
				"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
				"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
				"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
			}})
			actor := uuid.MustParse("11111111-1111-4111-8111-111111111111")
			incident := uuid.MustParse("22222222-2222-4222-8222-222222222222")
			appsupport.SeedMembershipSQL(t, h.DB, incident, actor, "admin", actor)
			keys, err := authn.LoadMasterKeys(map[string]string{authn.AuthMasterKeyEnv: strings.TrimSpace(string(fixtures.MustRead("auth", "master-key.base64")))})
			if err != nil {
				t.Fatal(err)
			}
			const token = "canonical-upgrade-session"
			flowtest.SeedSession(t, h.DB, keys, actor, token, time.Now().UTC(), time.Now().UTC())
			csrf := authn.CSRFTokenForSessionToken(keys, token)
			r := &publicationFixtureRuntime{t: t, h: h, requestOptions: []func(*http.Request){httptestx.WithCookies(&http.Cookie{Name: authn.SessionCookieName, Value: token}, &http.Cookie{Name: authn.CSRFCookieName, Value: csrf}), httptestx.WithHeader(authn.CSRFHeaderName, csrf)}}
			rootPath := "/api/v1/incidents/" + incident.String() + "/network-flow"
			stage = "reference_reuse"
			link := map[string]any{"schema_id": "cartulary.network_flow.indicator_link_request.v1", "client_txn_id": "upgrade-reuse", "selector": map[string]any{"kind": "row_refs", "field_key": "network_flow.src_ip", "row_refs": refs}, "target": map[string]any{"mode": "existing_indicator", "indicator_id": "55555555-5555-4555-8555-555555555555"}, "observation_mode": "binding_only", "confirm_exact_value": "192.0.2.1"}
			linked := r.request(http.MethodPost, rootPath+"/indicator-links", link, http.StatusOK)
			if linked["binding"].(map[string]any)["network_flow_indicator_binding_id"] != "nfb_22222222222222222222222222222222" {
				t.Fatal("reference reuse duplicated binding")
			}
			requireUpgradeJSON(t, refs, mustJSON(t, linked["binding"].(map[string]any)["source_row_refs"]))
			linkedState := upgradeTables(t, db, []string{"network_flow_indicator_bindings", "route_idempotency", "deployment_admin_audit_events", "records", "indicators"})
			requireUpgradeJSON(t, mustJSON(t, linked), mustJSON(t, r.request(http.MethodPost, rootPath+"/indicator-links", link, http.StatusOK)))
			if !reflect.DeepEqual(linkedState, upgradeTables(t, db, []string{"network_flow_indicator_bindings", "route_idempotency", "deployment_admin_audit_events", "records", "indicators"})) {
				t.Fatal("link replay added durable effects")
			}
			stage = "historical_graph_replay"
			replayState := upgradeTables(t, db, []string{"network_flow_graph_views", "route_idempotency", "jobs", "extension_job_commit_proofs", "deployment_admin_audit_events"})
			for _, entry := range vector.Seed {
				if entry.Table != "route_idempotency" {
					continue
				}
				var route, client string
				_ = json.Unmarshal(entry.Row["route_key"], &route)
				_ = json.Unmarshal(entry.Row["client_txn_id"], &client)
				var expected map[string]any
				if err := json.Unmarshal(entry.Row["response_json"], &expected); err != nil {
					t.Fatal(err)
				}
				g := expected["graph_view"].(map[string]any)
				g["semantic_query_sha256"] = vector.Expected.SemanticHash
				g["semantic_query"].(map[string]any)["filters"] = vector.Expected.Filters
				path := rootPath + "/graph-views"
				method := http.MethodPost
				status := http.StatusAccepted
				body := map[string]any{"schema_id": "cartulary.network_flow.graph_view_create_request.v3", "client_txn_id": client, "display_name": g["display_name"], "semantic_query": g["semantic_query"]}
				if route == "nf.graph_views.patch" {
					path += "/" + g["graph_view_id"].(string)
					method = http.MethodPatch
					status = http.StatusOK
					body = map[string]any{"schema_id": "cartulary.network_flow.graph_view_rename_request.v2", "client_txn_id": client, "display_name": g["display_name"], "base_graph_view_version": g["graph_view_version"].(float64) - 1}
				}
				requireUpgradeJSON(t, mustJSON(t, expected), mustJSON(t, r.request(method, path, body, status)))
			}
			if !reflect.DeepEqual(replayState, upgradeTables(t, db, []string{"network_flow_graph_views", "route_idempotency", "jobs", "extension_job_commit_proofs", "deployment_admin_audit_events"})) {
				t.Fatal("historical replay added durable effects")
			}
			stage = "refresh"
			refresh := map[string]any{"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "upgrade-refresh", "base_graph_view_version": 2}
			refreshed := r.request(http.MethodPost, rootPath+"/graph-views/"+graphID+"/refresh", refresh, http.StatusAccepted)
			if refreshed["graph_view"].(map[string]any)["selected_result_binding"].(map[string]any)["projection_result_id"] != immutableID {
				t.Fatal("refresh discarded last-safe result")
			}
			job := r.waitJob(refreshed["job"].(map[string]any)["job_id"].(string))
			if job["status"] != "succeeded" {
				t.Fatalf("refresh job: %v", job)
			}
			result := r.request(http.MethodGet, rootPath+"/graph-views/"+graphID+"/result", nil, http.StatusOK)["result"].(map[string]any)["graph_projection_result"].(map[string]any)
			if result["projection_result_id"] == immutableID || len(result["vertices"].([]any)) != 0 || len(result["edges"].([]any)) != 0 {
				t.Fatal("refresh did not publish the new empty filtered graph")
			}
			if !reflect.DeepEqual(immutableBefore, upgradeImmutableState(t, db, immutableID)) {
				t.Fatal("startup/replay/refresh changed immutable historical result or lease")
			}
			afterRefresh := upgradeTables(t, db, []string{"network_flow_graph_views", "jobs", "extension_job_commit_proofs", "route_idempotency", "deployment_admin_audit_events"})
			requireUpgradeJSON(t, mustJSON(t, refreshed), mustJSON(t, r.request(http.MethodPost, rootPath+"/graph-views/"+graphID+"/refresh", refresh, http.StatusAccepted)))
			if !reflect.DeepEqual(afterRefresh, upgradeTables(t, db, []string{"network_flow_graph_views", "jobs", "extension_job_commit_proofs", "route_idempotency", "deployment_admin_audit_events"})) {
				t.Fatal("refresh replay duplicated effects")
			}
			if err := ValidateRetainedExtensionState(t.Context(), h.Pool); err != nil {
				t.Fatal(err)
			}
			if historicalLink != canonicalHistoricalLinkBytes(t, db) {
				t.Fatal("serving changed historical link receipt")
			}
			for _, definition := range appsupport.RecognizedJobDefinitions(t) {
				if definition.JobKind == GraphViewMaterializationJobKind {
					if err := ValidateSavedGraphAdmission(t.Context(), h.Pool, definition); err != nil {
						t.Fatal(err)
					}
				}
			}
			stage = "complete"
		})
	}
}

// Historical link receipts preserve their old references and request identity.
// Their shape is admitted independently from repaired live binding lineage.
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
	for _, table := range []string{"graph_projection_results", "graph_projection_result_vertices", "graph_projection_result_edges", "graph_projection_result_leases"} {
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
	upgradeExec(t, db, `INSERT INTO graph_projection_result_leases(lease_id,projection_result_id,lease_owner_id,lease_owner_resource_id,lease_purpose,leased_until,created_at,renewed_at) VALUES ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',$1,'snapshot_reporting','canonical-upgrade','render','2027-07-10T12:00:00Z','2026-07-10T12:00:00Z','2026-07-10T12:00:00Z')`, b.ProjectionResultID)
	return b.ProjectionResultID
}
