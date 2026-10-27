package reporting_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func TestSnapshotAdmissionAndCancellationDoNotWaitForSourceCapture_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-reporting-async-capture", TestRouteMode: httptestx.TestRouteModeDisabled})
	actor, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, actor, map[string]any{"incident_key": "IR-ASYNC-SNAPSHOT", "title": "Async capture", "client_txn_id": "async-incident"})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	barrier, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer barrier.Rollback(context.Background())
	if _, err := barrier.Exec(ctx, `LOCK TABLE timeline_events IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatal(err)
	}
	post := func(path string, body map[string]any) *http.Response {
		t.Helper()
		request := httptestx.NewJSONRequest(t, http.MethodPost, harness.Server.HTTP.URL+path, body).WithContext(ctx)
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie)(request)
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value)(request)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	body := map[string]any{"incident_id": incident["incident_id"], "client_txn_id": "async-snapshot"}
	job := httptestx.RequireSuccessEnvelope(t, post("/api/v1/snapshots", body), http.StatusAccepted)["data"].(map[string]any)
	id := job["job_id"].(string)
	// The worker must really be blocked in the source read. An immediate cancel
	// of a still-queued Job would not demonstrate interruptible capture.
	for {
		var waiting bool
		if err := harness.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, int32(barrier.Conn().PgConn().PID())).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("worker did not enter capture", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	replay := httptestx.RequireSuccessEnvelope(t, post("/api/v1/snapshots", body), http.StatusAccepted)["data"].(map[string]any)
	if replay["job_id"] != id {
		t.Fatal("pending replay allocated another Job")
	}
	httptestx.RequireSuccessEnvelope(t, post("/api/v1/jobs/"+id+"/cancel", map[string]any{"client_txn_id": "cancel-capture"}), http.StatusOK)
	requireJobStatus(t, harness, actor, id, "canceled")
	var count int
	if err := harness.DB.QueryRow(`SELECT (SELECT count(*) FROM reporting_job_payloads WHERE job_id=$1 AND request_json IS NOT NULL)+(SELECT count(*) FROM reporting_snapshots WHERE create_job_id=$1)`, id).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("interrupted capture published immutable state")
	}
	// The source lock is intentionally still held through all assertions.
}

func assertSnapshotPublicationCandidate(t *testing.T, harness *appsupport.ServerHarness, jobID, snapshotID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	candidate, err := reporting.ValidateSnapshotCandidate(ctx, harness.Pool, jobID)
	if err != nil {
		t.Fatal(err)
	}
	complete := func(proof reporting.SnapshotCandidate, id uuid.UUID) error {
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		return (&reporting.Store{}).CompleteSnapshotCreateJobTx(ctx, tx, jobID, id, proof)
	}
	if err := complete(candidate, snapshotID); err != nil {
		t.Fatal("validated snapshot replay rejected", err)
	}
	if complete(reporting.SnapshotCandidate{}, snapshotID) == nil || complete(candidate, uuid.New()) == nil {
		t.Fatal("unvalidated snapshot publication accepted")
	}
	var original []byte
	if err := harness.DB.QueryRow(`SELECT request_json FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&original); err != nil {
		t.Fatal(err)
	}
	restore := func() {
		t.Helper()
		if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1`, jobID, original); err != nil {
			t.Fatal(err)
		}
	}
	defer restore()
	// Keep embedded hashes unchanged: the optimistic publication fence must
	// detect the changed row without decoding case content under the Jobs lock.
	if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{source_capture,incident,title}','"substituted"') WHERE job_id=$1`, jobID); err != nil {
		t.Fatal(err)
	}
	if complete(candidate, snapshotID) == nil {
		t.Fatal("capture changed after validation was published")
	}
	if _, err := reporting.ValidateSnapshotCandidate(ctx, harness.Pool, jobID); err == nil {
		t.Fatal("altered capture validated")
	}
	restore()
	if complete(candidate, snapshotID) == nil {
		t.Fatal("restored bytes revived a stale publication proof")
	}
	candidate, err = reporting.ValidateSnapshotCandidate(ctx, harness.Pool, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if err := complete(candidate, snapshotID); err != nil {
		t.Fatal("fresh proof rejected", err)
	}
	if _, err := harness.DB.Exec(`UPDATE reporting_snapshots SET source_capture_sha256=repeat('0',64) WHERE snapshot_id=$1`, snapshotID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := harness.DB.Exec(`UPDATE reporting_snapshots SET source_capture_sha256=(SELECT request_json->>'source_capture_sha256' FROM reporting_job_payloads WHERE job_id=$2) WHERE snapshot_id=$1`, snapshotID, jobID); err != nil {
			t.Fatal(err)
		}
	}()
	if complete(candidate, snapshotID) == nil {
		t.Fatal("changed independent snapshot digest accepted")
	}
}
