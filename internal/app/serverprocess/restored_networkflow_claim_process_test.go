package serverprocess

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	evidencemodule "github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
)

func restoredNetworkFlowClaimEnv() map[string]string {
	return map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}
}

func TestRestoredNetworkFlowClaimBoundary_Process(t *testing.T) {
	for _, scenario := range []string{"valid", "invalid_semantics", "unsupported_version"} {
		t.Run(scenario, func(t *testing.T) {
			var graphID, resultID, sourceState, sourceProof, jobID string
			fixture := captureRestoreSourceWithNetworkFlow(t, "inactive-"+scenario, func(source *processtest.Server, pool *pgxpool.Pool, login authcookietest.AuthCookies, incident string) {
				db := stdlib.OpenDBFromPool(pool)
				defer closeSQL(t, db)
				var actor string
				if err := pool.QueryRow(t.Context(), `SELECT id::text FROM users WHERE email=$1`, authenticationBootstrapAdminEmail).Scan(&actor); err != nil {
					t.Fatal(err)
				}
				table := networkflowsupport.SeedGraphSource(t, db, incident, actor)
				body := map[string]any{"schema_id": "cartulary.network_flow.graph_view_create_request.v3", "client_txn_id": "restored-claim-graph", "display_name": "Retained graph", "semantic_query": map[string]any{
					"schema_id": "cartulary.network_flow.graph_semantic_query.v2", "selected_table_ids": []string{table}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"},
				}}
				created := httptestx.RequireSuccessEnvelope(t, doJSON(t, source, http.MethodPost, "/api/v1/incidents/"+incident+"/network-flow/graph-views", body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusAccepted)["data"].(map[string]any)
				graphID = created["graph_view"].(map[string]any)["graph_view_id"].(string)
				jobID = created["job"].(map[string]any)["job_id"].(string)
				waitRestoredGraphFact(t, func() bool {
					var status string
					if err := pool.QueryRow(t.Context(), `SELECT status FROM jobs WHERE job_id::text=$1`, jobID).Scan(&status); err != nil {
						t.Fatal(err)
					}
					if status == "failed" || status == "canceled" {
						t.Fatal("source Graph materialization failed")
					}
					return status == "succeeded"
				})
				if err := pool.QueryRow(t.Context(), `SELECT selected_projection_result_id FROM network_flow_graph_views WHERE graph_view_id=$1`, graphID).Scan(&resultID); err != nil {
					t.Fatal(err)
				}
				source.Stop(t)
				switch scenario {
				case "invalid_semantics":
					if _, err := pool.Exec(t.Context(), `UPDATE network_flow_graph_views SET semantic_query_json=jsonb_set(semantic_query_json,'{aggregation,mode}','"unsupported_mode"'::jsonb) WHERE graph_view_id=$1`, graphID); err != nil {
						t.Fatal(err)
					}
				case "unsupported_version":
					if _, err := pool.Exec(t.Context(), `UPDATE extension_state_metadata SET state_version=1 WHERE profile_id='network_flow_activity'`); err != nil {
						t.Fatal(err)
					}
				}
				sourceState = retainedNetworkFlowState(t, pool)
				sourceProof = retainedNetworkFlowJobProof(t, pool, jobID)
			})

			target := prepareRestoreTarget(t, "inactive-target-"+scenario)
			core, err := projectionassembly.Build(target.Postgres)
			if err != nil {
				t.Fatal(err)
			}
			catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			runner := recovery.NewVersionedRestoreRunner(fixture.SourceStore, fixture.BackupStorage, recoveryExtensionCatalog(t), catalog)
			admitted := admitProcessRestore(t, fixture, target, runner, recovery.RestoreTarget{ProofScratch: recoverytest.ProofScratch(t), Postgres: target.RecoveryPool, ObjectStore: target.ObjectStore, ReferencePacks: openRecoveryPackStorage(t, target.Env), EvidenceObjects: evidencemodule.NewRecoveryProvider(target.Postgres), Projections: core.RecoveryPorts().Rebuilder}, application.OperationRestoreLatest)
			result, err := runner.RestoreLatestSuccessfulRetained(t.Context(), admitted.Fence.Target(), fixture.AsOf)
			if err != nil {
				t.Fatal("inactive structural restore invoked profile semantics", err)
			}
			admitted.Complete(t, result)
			if err := admitted.Admission.Release(t.Context()); err != nil {
				t.Fatal(err)
			}
			control, err := recovery.ReadTargetControl(t.Context(), target.Postgres)
			if err != nil || control.State != recovery.TargetComplete {
				t.Fatal("structural completion missing", err)
			}
			assertUnchanged := func() {
				t.Helper()
				if retainedNetworkFlowJobProof(t, target.Postgres, jobID) != sourceProof {
					t.Fatal("restore or claim changed the retained Job result/proof")
				}
				if retainedNetworkFlowState(t, target.Postgres) != sourceState {
					t.Fatal("inactive restore or rejected claim changed authoritative profile bytes")
				}
				current, err := recovery.ReadTargetControl(t.Context(), target.Postgres)
				if err != nil || current.Revision != control.Revision || current.State != control.State || current.CompletionEvidenceSHA256 == nil || *current.CompletionEvidenceSHA256 != *control.CompletionEvidenceSHA256 {
					t.Fatal("profile claim changed structural completion", err)
				}
			}
			assertUnchanged()
			requireRestoredGraphCount(t, target.Postgres, 0)
			delete(target.Env, recovery.RecoveryMasterKeyEnv)
			target.Env["CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED"] = "false"
			inactive := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
			t.Cleanup(func() { inactive.Stop(t) })
			inactive.WaitForReady(t)
			inactive.RequireStatus(t, "/readyz", http.StatusOK)
			assertUnchanged()
			requireRestoredGraphCount(t, target.Postgres, 0)
			inactive.Stop(t)
			for k, v := range restoredNetworkFlowClaimEnv() {
				target.Env[k] = v
			}
			if scenario != "valid" {
				failed := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
				t.Cleanup(func() { failed.Stop(t) })
				var exit *exec.ExitError
				if err := failed.WaitForExit(t); !errors.As(err, &exit) || exit.ExitCode() != 2 {
					t.Fatalf("invalid retained profile claim exit: %v", err)
				}
				failed.RequireDiagnosticsCode(t, "invalid_deployment_config")
				reason := "extension_admission_validation_failed"
				if scenario == "unsupported_version" {
					reason = "extension_state_version_unsupported"
				}
				failed.RequireDiagnosticsField(t, "$", reason)
				failed.RequireConnectionRefused(t, "/healthz")
				failed.RequireConnectionRefused(t, "/api/v1/incidents/"+fixture.IncidentID+"/network-flow/graph-views")
				failed.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/"+fixture.IncidentID)
				assertUnchanged()
				requireRestoredGraphCount(t, target.Postgres, 0)
				return
			}

			// A determinate derived rebuild failure must leave Core serving and the
			// structural completion intact. A sequence observes the real attempt
			// despite rollback; no production fault hook or fake rebuilder is used.
			fixtureAdmin, err := pgxpool.New(t.Context(), target.database.DSN)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(fixtureAdmin.Close)
			if _, err := fixtureAdmin.Exec(t.Context(), `CREATE SEQUENCE restored_graph_attempts; GRANT USAGE ON SEQUENCE restored_graph_attempts TO cartulary_runtime;
CREATE FUNCTION fail_restored_graph() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('restored_graph_attempts'); RAISE EXCEPTION 'injected derived rebuild failure'; END $$;
CREATE TRIGGER fail_restored_graph BEFORE INSERT ON graph_projection_results FOR EACH ROW EXECUTE FUNCTION fail_restored_graph();`); err != nil {
				t.Fatal(err)
			}
			failedRebuild := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
			t.Cleanup(func() { failedRebuild.Stop(t) })
			failedRebuild.WaitForReady(t)
			waitRestoredGraphFact(t, func() bool {
				var called bool
				if err := fixtureAdmin.QueryRow(t.Context(), `SELECT is_called FROM restored_graph_attempts`).Scan(&called); err != nil {
					t.Fatal(err)
				}
				return called
			})
			login := flowtest.LoginLocalUserWithSecondFactor(t, failedRebuild.Client, failedRebuild.BaseURL, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, fixture.AdminTOTPSecret))
			path := "/api/v1/incidents/" + fixture.IncidentID + "/network-flow/graph-views/" + graphID + "/result"
			httptestx.RequireErrorEnvelope(t, doJSON(t, failedRebuild, http.MethodGet, path, nil, httptestx.WithCookies(login.Session)), http.StatusBadGateway, "network_flow_graph_projection_failed")
			failedRebuild.RequireStatus(t, "/readyz", http.StatusOK)
			failedRebuild.Stop(t)
			assertUnchanged()
			requireRestoredGraphCount(t, target.Postgres, 0)
			if _, err := fixtureAdmin.Exec(t.Context(), `DROP TRIGGER fail_restored_graph ON graph_projection_results; DROP FUNCTION fail_restored_graph(); DROP SEQUENCE restored_graph_attempts;`); err != nil {
				t.Fatal(err)
			}
			ready := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
			t.Cleanup(func() { ready.Stop(t) })
			ready.WaitForReady(t)
			waitRestoredGraphFact(t, func() bool {
				var count int
				if err := target.Postgres.QueryRow(t.Context(), `SELECT count(*) FROM graph_projection_results WHERE projection_result_id=$1`, resultID).Scan(&count); err != nil {
					t.Fatal(err)
				}
				return count == 1
			})
			// Current authentication remains required; neither restored sessions nor
			// the previous process's login are used as a reconstruction capability.
			login = flowtest.LoginLocalUserWithSecondFactor(t, ready.Client, ready.BaseURL, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, fixture.AdminTOTPSecret))
			httptestx.RequireSuccessEnvelope(t, doJSON(t, ready, http.MethodGet, path, nil, httptestx.WithCookies(login.Session)), http.StatusOK)
			assertUnchanged()
		})
	}
}

func retainedNetworkFlowState(t testing.TB, pool *pgxpool.Pool) string {
	t.Helper()
	var state string
	if err := pool.QueryRow(t.Context(), `SELECT jsonb_build_array(
(SELECT jsonb_agg(to_jsonb(t) ORDER BY network_flow_table_id) FROM network_flow_tables t),
(SELECT jsonb_agg(to_jsonb(r) ORDER BY network_flow_row_id) FROM network_flow_rows r),
(SELECT jsonb_agg(to_jsonb(v) ORDER BY graph_view_id) FROM network_flow_graph_views v),
(SELECT to_jsonb(m) FROM extension_state_metadata m WHERE profile_id='network_flow_activity'))::text`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return state
}

func requireRestoredGraphCount(t testing.TB, pool *pgxpool.Pool, expected int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM graph_projection_results)+(SELECT count(*) FROM graph_projection_result_vertices)+(SELECT count(*) FROM graph_projection_result_edges)`).Scan(&count); err != nil || count != expected {
		t.Fatalf("derived Graph rows: count=%d expected=%d err=%v", count, expected, err)
	}
}

func waitRestoredGraphFact(t testing.TB, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("current Graph process did not reach the required boundary")
}

// Verify the actual worker's terminal digest after PostgreSQL JSONB storage and
// compare the complete retained proof across structural restore and later claims.
func retainedNetworkFlowJobProof(t testing.TB, pool *pgxpool.Pool, jobID string) string {
	t.Helper()
	var terminal, state []byte
	var digest string
	if err := pool.QueryRow(t.Context(), `SELECT p.terminal_result::text, p.terminal_result_sha256,
 jsonb_build_object('proof',to_jsonb(p),'public_result',j.result_summary_json)::text
 FROM extension_job_commit_proofs p JOIN jobs j ON j.job_id=p.job_id
 WHERE p.job_id::text=$1`, jobID).Scan(&terminal, &digest, &state); err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljson.Canonicalize(terminal)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(canonical)
	if hex.EncodeToString(hash[:]) != digest {
		t.Fatal("retained terminal digest does not use the current logical byte contract")
	}
	canonicalState, err := canonicaljson.Canonicalize(state)
	if err != nil {
		t.Fatal(err)
	}
	return string(canonicalState)
}
