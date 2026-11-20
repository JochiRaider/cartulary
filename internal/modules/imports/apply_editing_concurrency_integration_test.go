package imports_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// AC-613: while an import apply unit is in progress in an incident, a Timeline
// patch of another record in that incident commits before the apply does. A
// time-conversion profile is present, so the apply unit and the patch both
// read it in their transactions.
func TestImportApplyUnitDoesNotBlockTimelinePatch_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "imports-apply-editing", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-apply-editing-incident", "incident_key": "IR-APPLY-EDITING", "title": "Apply does not block editing",
	})["incident_id"].(string)
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, base, login, http.MethodPut, "/api/v1/incidents/"+incidentID+"/timeline-time-conversion-profile",
		map[string]any{"base_profile_version": 1, "enabled": true, "local_offset_minutes": -240, "local_label": "UTC-04"}), http.StatusOK)
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM timeline_time_conversion_profiles WHERE incident_id::text=$1 AND enabled`, incidentID); got != 1 {
		t.Fatalf("time-conversion profiles = %d, want 1", got)
	}
	target := createApplyEditingTarget(t, base, login, incidentID)

	sessionID, unitID := startCSVImportSession(t, base, login, incidentID, "txn-apply-editing-upload",
		"summary\nfirst imported row\nheld imported row\nlast imported row\n", "apply-editing.csv")
	approveTimelineImportMapping(t, base, login, sessionID, unitID, []string{"summary"}, "apply-editing")
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	// The unit has created its first row and read the profile for its second
	// before the second row's insert waits.
	hold := holdImportApplyInsert(t, ctx, harness.DB, "timeline_events", "NEW.activity_synopsis_text = 'held imported row'")
	defer hold.release(t)
	jobID := selectAndApplyImport(t, base, login, sessionID, unitID, "apply-editing")
	applyPID := hold.waitHeld(t, ctx, base, login, jobID)
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM timeline_events WHERE incident_id::text=$1`, incidentID); got != 1 {
		t.Fatalf("uncommitted apply rows are visible: timeline events = %d, want 1", got)
	}

	response := hold.requireCommitsWhileHeld(t, ctx, applyPID, func(ctx context.Context) (*http.Response, error) {
		return sendApplyEditingPatch(ctx, base, login, target, "txn-apply-editing-patch")
	})
	patched := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)["row"].(map[string]any)
	if patched["record_id"] != target["record_id"] {
		t.Fatalf("patched row = %v, want %v", patched["record_id"], target["record_id"])
	}
	hold.requireStillHeld(t, ctx, applyPID)
	// The patch read the shared profile: its local counterpart was generated.
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM timeline_events WHERE record_id::text=$1 AND activity_utc_text='2026-04-10T10:00:00Z' AND activity_local_generated AND activity_local_text IS NOT NULL`, target["record_id"]); got != 1 {
		t.Fatal("committed patch did not apply the time-conversion profile")
	}

	hold.release(t)
	if job := waitImportJobTerminal(t, base, login, jobID); job["status"] != "succeeded" {
		t.Fatalf("apply job = %#v", job)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM timeline_events WHERE incident_id::text=$1 AND activity_synopsis_text LIKE '% imported row'`, incidentID); got != 3 {
		t.Fatalf("imported timeline events = %d, want 3", got)
	}
	// History order follows commit order (REQ-02-213): the patch committed first.
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT count(*)
  FROM change_sets patch, change_sets applied, import_unit_apply_outcomes outcome
 WHERE patch.incident_id::text = $1 AND patch.client_txn_id = 'txn-apply-editing-patch'
   AND outcome.import_unit_id::text = $2 AND outcome.outcome_status = 'applied'
   AND applied.change_set_id = outcome.change_set_id
   AND patch.history_order < applied.history_order`, incidentID, unitID); got != 1 {
		t.Fatal("the patch's change set does not precede the apply unit's in history order")
	}
}

// AC-613: while a Network Flow table import copies its rows in an incident, an
// ordinary patch in that incident commits. The import takes the incident lock
// only to publish its table.
func TestNetworkFlowImportCopyDoesNotBlockIncidentPatch_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "network-flow-copy-editing",
		Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	base := harness.Server.HTTP.URL
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-nf-copy-editing-incident", "incident_key": "IR-NF-COPY-EDITING", "title": "Network Flow copy does not block editing",
	})["incident_id"].(string)
	target := createApplyEditingTarget(t, base, login, incidentID)

	const acceptedRows = 300
	sessionID, unitID := startCSVImportSession(t, base, login, incidentID, "txn-nf-copy-editing-upload",
		sourceAdmissionHeader+strings.Repeat(sourceAdmissionRow, acceptedRows), "flows.csv")
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, base, login, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
		networkFlowMappingPayload("txn-nf-copy-editing-mapping")), http.StatusOK)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	// Rows are copied in batches of at most 128. Source row 200 is in the
	// second batch, after the table and the first batch were written.
	hold := holdImportApplyInsert(t, ctx, harness.DB, "network_flow_rows", "NEW.source_row_number = 200")
	defer hold.release(t)
	jobID := selectAndApplyImport(t, base, login, sessionID, unitID, "nf-copy-editing")
	applyPID := hold.waitHeld(t, ctx, base, login, jobID)

	response := hold.requireCommitsWhileHeld(t, ctx, applyPID, func(ctx context.Context) (*http.Response, error) {
		return sendApplyEditingPatch(ctx, base, login, target, "txn-nf-copy-editing-patch")
	})
	httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	hold.requireStillHeld(t, ctx, applyPID)
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM timeline_events WHERE record_id::text=$1 AND activity_utc_text='2026-04-10T10:00:00Z'`, target["record_id"]); got != 1 {
		t.Fatal("patch did not commit while the Network Flow copy was in progress")
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM network_flow_tables WHERE incident_id::text=$1`, incidentID); got != 0 {
		t.Fatalf("unpublished Network Flow tables are visible: %d", got)
	}

	hold.release(t)
	if job := waitImportJobTerminal(t, base, login, jobID); job["status"] != "succeeded" {
		t.Fatalf("Network Flow apply job = %#v", job)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT count(*)
  FROM network_flow_tables t
 WHERE t.incident_id::text = $1 AND t.display_name = 'flows' AND t.table_version = 1 AND t.table_status = 'active'
   AND t.row_count_accepted = $2
   AND (SELECT count(*) FROM network_flow_rows r WHERE r.network_flow_table_id = t.network_flow_table_id) = $2`, incidentID, acceptedRows); got != 1 {
		t.Fatal("Network Flow import did not publish one complete table under its derived name")
	}
}

func createApplyEditingTarget(t *testing.T, base string, login authcookietest.AuthCookies, incidentID string) map[string]any {
	t.Helper()
	response := doImportJSON(t, base, login, http.MethodPost, "/api/v1/incidents/"+incidentID+"/views/cartulary.view.timeline.v2/rows",
		map[string]any{"client_txn_id": "txn-apply-editing-target", "timeline.activity_synopsis_text": "Edited while background work runs"})
	return httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)["row"].(map[string]any)
}

func sendApplyEditingPatch(ctx context.Context, base string, login authcookietest.AuthCookies, target map[string]any, clientTxnID string) (*http.Response, error) {
	body := fmt.Sprintf(`{"view_schema_id":"cartulary.view.timeline.v2","client_txn_id":%q,"base_row_version":%v,"changes":[{"field_key":"timeline.activity_utc_text","value":"2026-04-10T10:00:00Z"}]}`,
		clientTxnID, target["row_version"])
	request, err := http.NewRequestWithContext(ctx, http.MethodPatch, base+"/api/v1/records/"+target["record_id"].(string), strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	return http.DefaultClient.Do(request)
}

func selectAndApplyImport(t *testing.T, base string, login authcookietest.AuthCookies, sessionID, unitID, prefix string) string {
	t.Helper()
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, base, login, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
		map[string]any{"client_txn_id": "txn-" + prefix + "-select"}), http.StatusOK)
	response := doImportJSON(t, base, login, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-" + prefix + "-apply"})
	return httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
}

// importApplyHold stops one real owner insert inside an import apply unit's
// transaction on an advisory lock held by a dedicated test session. Waits are
// observed through pg_blocking_pids, never inferred from elapsed time.
type importApplyHold struct {
	db         *sql.DB
	blocker    *sql.Conn
	blockerPID int32
	released   bool
}

func holdImportApplyInsert(t *testing.T, ctx context.Context, db *sql.DB, table, condition string) *importApplyHold {
	t.Helper()
	blocker, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	hold := &importApplyHold{db: db, blocker: blocker}
	var namespace, key int32
	if err := blocker.QueryRowContext(ctx, "SELECT pg_backend_pid(), hashtext('imports-apply-editing'), hashtext($1)", t.Name()).Scan(&hold.blockerPID, &namespace, &key); err != nil {
		hold.release(t)
		t.Fatal(err)
	}
	if _, err := blocker.ExecContext(ctx, "SELECT pg_advisory_lock($1, $2)", namespace, key); err != nil {
		hold.release(t)
		t.Fatal(err)
	}
	fixtureSQL := fmt.Sprintf(`
CREATE FUNCTION public.hold_import_apply_insert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(%d, %d);
    RETURN NEW;
END;
$$;
CREATE TRIGGER hold_import_apply_insert BEFORE INSERT ON %s
FOR EACH ROW WHEN (%s) EXECUTE FUNCTION public.hold_import_apply_insert();
`, namespace, key, table, condition)
	if _, err := db.ExecContext(ctx, fixtureSQL); err != nil {
		hold.release(t)
		t.Fatal(err)
	}
	return hold
}

func (h *importApplyHold) release(t *testing.T) {
	t.Helper()
	if h.released {
		return
	}
	h.released = true
	closeImportAdvisoryGate(t, h.blocker)
}

// waitHeld returns the apply unit's backend once it waits on the hold.
func (h *importApplyHold) waitHeld(t *testing.T, ctx context.Context, base string, login authcookietest.AuthCookies, jobID string) int32 {
	t.Helper()
	for {
		var waiting []int32
		rows, err := h.db.QueryContext(ctx, `SELECT pid FROM pg_stat_activity WHERE $1::integer = ANY(pg_blocking_pids(pid))`, h.blockerPID)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var pid int32
			if err := rows.Scan(&pid); err != nil {
				t.Fatal(err)
			}
			waiting = append(waiting, pid)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
		if len(waiting) > 1 {
			t.Fatalf("more than one backend waits on the import hold: %v", waiting)
		}
		if len(waiting) == 1 {
			return waiting[0]
		}
		select {
		case <-ctx.Done():
			job := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, base+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.Session)), http.StatusOK)
			t.Fatalf("import apply never reached its held insert: %v; job %#v", ctx.Err(), job["data"])
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (h *importApplyHold) requireStillHeld(t *testing.T, ctx context.Context, applyPID int32) {
	t.Helper()
	var held bool
	if err := h.db.QueryRowContext(ctx, `SELECT $2::integer = ANY(pg_blocking_pids($1::integer))`, applyPID, h.blockerPID).Scan(&held); err != nil {
		t.Fatal(err)
	}
	if !held {
		t.Fatalf("import apply backend %d is no longer held uncommitted", applyPID)
	}
}

// requireCommitsWhileHeld fails as soon as any backend waits on the held apply
// transaction, and otherwise returns the edit's response.
func (h *importApplyHold) requireCommitsWhileHeld(t *testing.T, ctx context.Context, applyPID int32, send func(context.Context) (*http.Response, error)) *http.Response {
	t.Helper()
	requestCtx, cancelRequest := context.WithCancel(ctx)
	done := make(chan struct{})
	var response *http.Response
	var requestErr error
	returned := false
	t.Cleanup(func() {
		cancelRequest()
		select {
		case <-done:
			if response != nil && !returned {
				_ = response.Body.Close()
			}
		case <-time.After(5 * time.Second):
			t.Error("concurrent edit did not settle during cleanup")
		}
	})
	go func() {
		defer close(done)
		response, requestErr = send(requestCtx)
	}()
	for {
		select {
		case <-done:
			if requestErr != nil {
				t.Fatalf("concurrent edit failed: %v", requestErr)
			}
			returned = true
			return response
		case <-ctx.Done():
			t.Fatalf("concurrent edit neither committed nor waited on the apply: %v", ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
		var waiting string
		if err := h.db.QueryRowContext(ctx, `
SELECT coalesce(string_agg(format('pid=%s wait=%s/%s query=%s', pid, wait_event_type, wait_event, left(regexp_replace(query, '\s+', ' ', 'g'), 120)), '; '), '')
  FROM pg_stat_activity
 WHERE $1::integer = ANY(pg_blocking_pids(pid))`, applyPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting != "" {
			t.Fatalf("ordinary edit waited on the in-progress import apply backend %d: %s", applyPID, waiting)
		}
	}
}
