package imports_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type currentImportIDs struct{ actor, incident, session, unit, job, txn string }

func currentExec(t testing.TB, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}

func currentImportOptions() appsupport.ServerOptions {
	return appsupport.ServerOptions{Prefix: "current-import", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}}
}

// Construct all workflow state through current public admission. Only the
// worker claim is paused; no fixture manufactures plans, mappings or receipts.
func admitCurrentWorkflow(t *testing.T, runtime *appsupport.Runtime, family string) (*appsupport.ServerHarness, appsupport.ServerOptions, currentImportIDs) {
	t.Helper()
	options := currentImportOptions()
	first := runtime.StartServer(t, options)
	ids := currentImportIDs{txn: "txn-current-apply"}
	ids.actor = flowtest.SeedLocalUserFlags(t, first.DB, "current-import@example.test", "Current import", "Current-import-test-42!", false, true, true)
	login := flowtest.LoginLocalUser(t, http.DefaultClient, first.Server.HTTP.URL, "current-import@example.test", "Current-import-test-42!", nil)
	incident := scenariotest.CreateIncident(t, first.Server, login, map[string]any{"client_txn_id": "current-incident", "incident_key": "IR-CURRENT", "title": "Current import"})
	ids.incident = incident["incident_id"].(string)
	if family == "generic" {
		ids.session, ids.unit = startCSVImportSession(t, first.Server.HTTP.URL, login, ids.incident, "current-upload", "summary,optional\ncurrent row\n", "current.csv")
		assertCurrentGenericPreview(t, first, login, ids)
		body := map[string]any{"client_txn_id": "current-mapping", "target_view_schema_id": "cartulary.view.timeline.v2", "header_row_ref": 1, "data_start_row_ref": 2, "unknown_column_policy": "preserve_raw_capture", "source_columns": []map[string]any{{"source_column_ordinal": 1, "source_header_text": "summary", "field_key": "timeline.activity_synopsis_text", "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"}, {"source_column_ordinal": 2, "source_header_text": "optional", "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"}}}
		response := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPut, "/api/v1/import-sessions/"+ids.session+"/units/"+ids.unit+"/mapping", body)
		httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
		assertCurrentGenericPreview(t, first, login, ids)
	} else {
		ids.session, ids.unit = startCSVImportSession(t, first.Server.HTTP.URL, login, ids.incident, "current-upload", sourceAdmissionHeader+sourceAdmissionRow, "current.csv")
		base := "/api/v1/import-sessions/" + ids.session + "/units/" + ids.unit
		inspection := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPost, base+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"})
		httptestx.RequireSuccessEnvelope(t, inspection, http.StatusOK)
		preview := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPost, base+"/mapping-preview", networkFlowMappingPreviewPayload())
		httptestx.RequireSuccessEnvelope(t, preview, http.StatusOK)
		mapping := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPut, base+"/mapping", networkFlowMappingPayload("current-mapping"))
		httptestx.RequireSuccessEnvelope(t, mapping, http.StatusOK)
	}
	selected := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+ids.session+"/units/"+ids.unit+"/select", map[string]any{"client_txn_id": "current-select"})
	httptestx.RequireSuccessEnvelope(t, selected, http.StatusOK)
	blocker, err := first.DB.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Close()
	if _, err = blocker.ExecContext(t.Context(), `SELECT pg_advisory_lock(181,303)`); err != nil {
		t.Fatal(err)
	}
	defer blocker.ExecContext(context.Background(), `SELECT pg_advisory_unlock(181,303)`)
	currentExec(t, first.DB, `CREATE FUNCTION current_claim_gate() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF OLD.status='queued' AND NEW.status='running' AND NEW.job_kind='import.apply_v1' THEN PERFORM pg_advisory_xact_lock(181,303); END IF; RETURN NEW; END $$; CREATE TRIGGER current_claim_gate BEFORE UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION current_claim_gate()`)
	response := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+ids.session+"/apply", map[string]any{"client_txn_id": ids.txn})
	ids.job = httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	if got := waitForSQLCount(t, first.DB, 10*time.Second, 1, `SELECT count(*) FROM pg_locks WHERE locktype='advisory' AND classid=181 AND objid=303 AND NOT granted`); got != 1 {
		t.Fatal("current worker did not reach claim boundary")
	}
	first.Server.Close()
	currentExec(t, first.DB, `DROP TRIGGER current_claim_gate ON jobs; DROP FUNCTION current_claim_gate()`)
	options.Database = first.Database
	options.ObjectStore = first.ObjectStore
	if currentEffectCount(t, first.DB, family, ids.incident) != 0 || currentOutcomeBytes(t, first.DB, ids.unit) != "[]" {
		t.Fatal("paused admission wrote owner effects")
	}
	return first, options, ids
}

func currentEffectCount(t testing.TB, db *sql.DB, family, incident string) int {
	t.Helper()
	if family == "generic" {
		return dbassert.CountSQL(t, db, `SELECT count(*) FROM timeline_events WHERE incident_id::text=$1`, incident)
	}
	return dbassert.CountSQL(t, db, `SELECT count(*) FROM network_flow_rows WHERE incident_id::text=$1`, incident)
}

func assertCurrentGenericPreview(t testing.TB, h *appsupport.ServerHarness, login authcookietest.AuthCookies, ids currentImportIDs) {
	t.Helper()
	response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+"/api/v1/import-sessions/"+ids.session+"/units/"+ids.unit+"/preview", nil, httptestx.WithCookies(login.Session))
	preview := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	columns := preview["columns"].([]any)
	rows := preview["preview_rows"].([]any)
	if len(columns) != 2 || len(rows) != 1 {
		t.Fatalf("unexpected projection: %#v", preview)
	}
	column := columns[0].(map[string]any)
	if len(column) != 2 || column["source_column_ordinal"] != float64(1) || column["source_header_text"] != "summary" {
		t.Fatalf("invalid public column shape: %#v", column)
	}
	row := rows[0].(map[string]any)
	if len(row) != 2 || row["source_row_ref"] != float64(2) {
		t.Fatal("invalid preview row", row)
	}
	cells := row["cells"].([]any)
	if len(cells) != 1 {
		t.Fatal("invalid cell count")
	}
	cell := cells[0].(map[string]any)
	if len(cell) != 3 || cell["source_column_ordinal"] != float64(1) || cell["display_text"] != "current row" || cell["cell_kind"] != "string" {
		t.Fatal("invalid public cell shape", cell)
	}
}

func TestCurrentImportWorkflows_Integration(t *testing.T) {
	runCurrentWorkflows(t, []string{"queued", "interrupted", "cancel_requested", "source_changed", "mapping_changed", "mapping_content_changed", "binding_changed", "unknown_registry", "rows_changed", "generic_access_denied", "before_commit_fault", "after_commit_fault", "committed_before_restart"})
}
func TestCurrentImportCancellation_Integration(t *testing.T) {
	runCurrentWorkflows(t, []string{"precommit_cancel"})
}

func runCurrentWorkflows(t *testing.T, scenarios []string) {
	for _, family := range []string{"generic", "analytical"} {
		for _, scenario := range scenarios {
			if family == "generic" && scenario == "generic_access_denied" || family == "analytical" && scenario == "rows_changed" {
				continue
			}
			t.Run(family+"/"+scenario, func(t *testing.T) {
				runtime := appsupport.StartRuntime(t)
				first, options, ids := admitCurrentWorkflow(t, runtime, family)
				db := first.DB
				switch scenario {
				case "interrupted":
					currentExec(t, db, `UPDATE jobs SET status='running',started_at=now()-interval '1 hour',handler_attempt_id=gen_random_uuid(),handler_lease_expires_at=now()-interval '1 minute' WHERE job_id=$1`, ids.job)
				case "cancel_requested":
					currentExec(t, db, `UPDATE jobs SET status='cancel_requested',cancelable=false WHERE job_id=$1`, ids.job)
				case "source_changed":
					currentExec(t, db, `UPDATE import_sessions SET source_content_sha256=repeat('a',64) WHERE import_session_id=$1`, ids.session)
				case "mapping_changed":
					currentExec(t, db, `UPDATE import_units SET mapping_fingerprint=repeat('b',64) WHERE import_unit_id=$1`, ids.unit)
				case "mapping_content_changed":
					currentExec(t, db, `UPDATE import_units SET approved_mapping_json=approved_mapping_json || '{"unexpected":"changed"}'::jsonb WHERE import_unit_id=$1`, ids.unit)
				case "binding_changed":
					currentExec(t, db, `UPDATE import_apply_unit_plans SET owner_binding_id='unknown.import_create' WHERE import_unit_id=$1`, ids.unit)
				case "unknown_registry":
					currentExec(t, db, `UPDATE import_apply_unit_plans SET target_registry_sha256=repeat('c',64) WHERE import_unit_id=$1`, ids.unit)
				case "rows_changed":
					currentExec(t, db, `UPDATE import_units SET source_rows_json='[]' WHERE import_unit_id=$1`, ids.unit)
				case "generic_access_denied":
					currentExec(t, db, `REVOKE SELECT ON import_units FROM cartulary_runtime; DO $$ DECLARE columns text; BEGIN SELECT string_agg(quote_ident(attname),',') INTO columns FROM pg_attribute WHERE attrelid='public.import_units'::regclass AND attnum>0 AND NOT attisdropped AND attname<>'source_rows_json'; EXECUTE 'GRANT SELECT (' || columns || ') ON public.import_units TO cartulary_runtime'; END $$;`)
				}
				var rowDigest, rectangle sql.NullString
				if err := db.QueryRowContext(t.Context(), `SELECT source_rows_sha256,source_rect_a1 FROM import_apply_unit_plans WHERE import_unit_id=$1`, ids.unit).Scan(&rowDigest, &rectangle); err != nil {
					t.Fatal(err)
				}
				if rowDigest.Valid != (family == "generic") || rectangle.Valid != (family == "generic") {
					t.Fatal("wrong current plan authority shape")
				}
				illegal := `UPDATE import_apply_unit_plans SET source_rows_sha256=NULL WHERE import_unit_id=$1`
				if family == "analytical" {
					illegal = `UPDATE import_apply_unit_plans SET source_rows_sha256=repeat('a',64) WHERE import_unit_id=$1`
				}
				if _, err := db.ExecContext(t.Context(), illegal, ids.unit); err == nil {
					t.Fatal("invalid source authority admitted")
				}
				release := installCurrentWriteWitness(t, db, family, scenario)
				defer release()
				if scenario == "before_commit_fault" || scenario == "after_commit_fault" {
					installCurrentCommitFault(t, db, scenario)
				}
				first = runtime.StartServer(t, options)
				if scenario == "before_commit_fault" || scenario == "after_commit_fault" {
					first = recoverCurrentCommitFault(t, db, runtime, first, options, ids, family, scenario)
				}
				if scenario == "precommit_cancel" {
					cancelCurrentBlockedUnit(t, db, first, ids)
				}
				login := flowtest.LoginLocalUser(t, http.DefaultClient, first.Server.HTTP.URL, "current-import@example.test", "Current-import-test-42!", nil)
				waitCurrentJob(t, db, ids.job)
				terminal := waitImportJobTerminal(t, first.Server.HTTP.URL, login, ids.job)
				want, effects := "failed", 0
				switch scenario {
				case "queued", "interrupted", "generic_access_denied", "before_commit_fault", "after_commit_fault", "committed_before_restart":
					want, effects = "succeeded", 1
				case "cancel_requested", "precommit_cancel":
					want = "canceled"
				}
				if terminal["status"] != want {
					t.Fatalf("current job = %#v, want %s", terminal, want)
				}
				assertCurrentOutcome(t, db, ids, family, effects)
				assertCurrentWriteAttempts(t, db, scenario, effects)
				settled := currentOutcomeBytes(t, db, ids.unit)
				if want == "failed" {
					reason := "source_changed"
					if scenario == "unknown_registry" {
						reason = "target_kind_not_importable"
					}
					var code, got string
					if err := db.QueryRowContext(t.Context(), `SELECT error_code,reason_code FROM import_unit_apply_outcomes WHERE import_unit_id=$1`, ids.unit).Scan(&code, &got); err != nil {
						t.Fatal(err)
					}
					if code != "import_apply_blocked" || got != reason {
						t.Fatalf("mismatch: %s/%s", code, got)
					}
				}
				for range 2 {
					first.Server.Close()
					first = runtime.StartServer(t, options)
					login = flowtest.LoginLocalUser(t, http.DefaultClient, first.Server.HTTP.URL, "current-import@example.test", "Current-import-test-42!", nil)
					response := doImportJSON(t, first.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+ids.session+"/apply", map[string]any{"client_txn_id": ids.txn})
					if data := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any); data["job_id"] != ids.job {
						t.Fatal("replay replaced job identity")
					}
					assertCurrentOutcome(t, db, ids, family, effects)
					assertCurrentWriteAttempts(t, db, scenario, effects)
					if currentOutcomeBytes(t, db, ids.unit) != settled {
						t.Fatal("replay rewrote durable outcome")
					}
					if family == "generic" && want == "succeeded" {
						assertCurrentGenericPreview(t, first, login, ids)
					}
				}
			})
		}
	}
}
func assertCurrentOutcome(t testing.TB, db *sql.DB, ids currentImportIDs, family string, effects int) {
	t.Helper()
	table := "records"
	if family == "analytical" {
		table = "network_flow_tables"
	}
	if got := dbassert.CountSQL(t, db, `SELECT count(*) FROM `+table+` WHERE incident_id::text=$1`, ids.incident); got != effects {
		t.Fatalf("owner parent effects %d, want %d", got, effects)
	}
	if family == "generic" {
		if got := dbassert.CountSQL(t, db, `SELECT count(*) FROM import_apply_journal WHERE import_unit_id::text=$1`, ids.unit); got != effects {
			t.Fatalf("owner journal effects %d, want %d", got, effects)
		}
		if got := dbassert.CountSQL(t, db, `SELECT count(*) FROM change_sets WHERE incident_id::text=$1 AND source='imports.apply'`, ids.incident); got != effects {
			t.Fatalf("owner change sets %d, want %d", got, effects)
		}
	}
	if got := currentEffectCount(t, db, family, ids.incident); got != effects {
		t.Fatalf("owner effects %d, want %d", got, effects)
	}
	if got := dbassert.CountSQL(t, db, `SELECT count(*) FROM import_unit_apply_outcomes WHERE import_session_id::text=$1 AND import_unit_id::text=$2 AND apply_job_id::text=$3 AND unit_commit_id=$4`, ids.session, ids.unit, ids.job, "import-unit:"+ids.session+":"+ids.unit); got != 1 {
		t.Fatalf("durable outcomes = %d", got)
	}
	if got := dbassert.CountSQL(t, db, `SELECT count(*) FROM route_idempotency WHERE route_key='imports.sessions.apply' AND scope_key=$1 AND client_txn_id=$2`, ids.session, ids.txn); got != 1 {
		t.Fatalf("idempotency identities = %d", got)
	}
}

// Report durable attempt state while waiting for recovered jobs, including the
// controlled retry cases. This does not change runner or product deadlines.
func waitCurrentJob(t testing.TB, db *sql.DB, job string) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		var status string
		var attempts int
		var last sql.NullString
		if err := db.QueryRowContext(t.Context(), `SELECT status,handler_failure_count,handler_last_error FROM jobs WHERE job_id=$1`, job).Scan(&status, &attempts, &last); err != nil {
			t.Fatal(err)
		}
		if status == "succeeded" || status == "failed" || status == "canceled" {
			return
		}
		if time.Now().After(deadline) {
			t.Log(databaseLockGraph(db))
			t.Fatalf("current recovery did not settle: status=%s failures=%d last=%s", status, attempts, last.String)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
