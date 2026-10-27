package imports_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestCancellationInterruptsUncommittedImportUnit_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "imports-uncommitted-cancellation", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-uncommitted-cancel-incident",
		"incident_key":  "IR-UNCOMMITTED-CANCEL", "title": "Uncommitted import cancellation",
	})
	incidentID := incident["incident_id"].(string)
	sessionID, unitID := prepareReadyTimelineImport(t, harness, login, incidentID, "uncommitted-cancel")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	blocker, err := harness.DB.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Close()
	var pid, namespace, key int32
	if err := blocker.QueryRowContext(ctx, "SELECT pg_backend_pid(), hashtext($1), hashtext($2)", "imports-cancel-test", t.Name()).Scan(&pid, &namespace, &key); err != nil {
		t.Fatal(err)
	}
	if _, err := blocker.ExecContext(ctx, "SELECT pg_advisory_lock($1, $2)", namespace, key); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = blocker.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1, $2)", namespace, key)
	}()
	// Stop actual owner SQL before its first effect. Cancellation must complete
	// and interrupt this query without the test releasing the blocker.
	fixtureSQL := fmt.Sprintf(`
CREATE FUNCTION public.block_uncommitted_import_row() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(%d, %d);
    RETURN NEW;
END;
$$;
CREATE TRIGGER block_uncommitted_import_row BEFORE INSERT ON timeline_events
FOR EACH ROW EXECUTE FUNCTION public.block_uncommitted_import_row();
`, namespace, key)
	if _, err := harness.DB.ExecContext(ctx, fixtureSQL); err != nil {
		t.Fatal(err)
	}
	response := doImportJSON(t, harness.Server.HTTP.URL, login, http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-uncommitted-cancel-apply"})
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	jobID := job["job_id"].(string)
	waiting := false
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := harness.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !waiting {
		t.Fatal("import owner did not reach the controlled uncommitted boundary")
	}
	cancelCtx, cancelRequest := context.WithTimeout(ctx, 5*time.Second)
	defer cancelRequest()
	request, err := http.NewRequestWithContext(cancelCtx, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", bytes.NewBufferString(`{"client_txn_id":"txn-uncommitted-cancel"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	request.AddCookie(login.SessionCookie)
	request.AddCookie(login.CSRFCookie)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("cancellation was blocked by uncommitted owner work: %v", err)
	}
	canceled := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	if canceled["status"] != "cancel_requested" {
		t.Fatalf("cancellation status = %v", canceled["status"])
	}
	terminal := waitImportJobTerminal(t, harness.Server.HTTP.URL, login, jobID)
	if terminal["status"] != "canceled" {
		t.Fatalf("terminal status = %v", terminal["status"])
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text=$1`, incidentID); got != 0 {
		t.Fatalf("canceled owner rows = %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM import_apply_journal WHERE import_unit_id::text=$1`, unitID); got != 0 {
		t.Fatalf("canceled journal rows = %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM import_unit_apply_outcomes WHERE import_unit_id::text=$1 AND outcome_status='canceled'`, unitID); got != 1 {
		t.Fatalf("durable canceled outcomes = %d", got)
	}
}
