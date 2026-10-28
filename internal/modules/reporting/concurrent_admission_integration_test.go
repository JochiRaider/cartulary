package reporting_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/jackc/pgx/v5"
)

// Hold an owner write, not the receipt lookup. Both requests must reach the
// database before either can commit, independently of scheduler timing.
func concurrentReportingMutation(t *testing.T, harness *appsupport.ServerHarness, actor authcookietest.AuthCookies, table, method, path string, body map[string]any, status int, identity string) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	barrier, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = barrier.Rollback(context.Background()) }()
	if _, err := barrier.Exec(ctx, "LOCK TABLE "+pgx.Identifier{table}.Sanitize()+" IN SHARE MODE"); err != nil {
		t.Fatal(err)
	}
	pid := int32(barrier.Conn().PgConn().PID())
	type outcome struct {
		response *http.Response
		err      error
	}
	results := make(chan outcome, 2)
	for range 2 {
		request := httptestx.NewJSONRequest(t, method, harness.Server.HTTP.URL+path, body).WithContext(ctx)
		httptestx.WithCookies(actor.Session, actor.CSRF)(request)
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRF.Value)(request)
		go func() { response, err := http.DefaultClient.Do(request); results <- outcome{response, err} }()
	}
	for {
		var waiting int
		err := harness.Pool.QueryRow(ctx, `WITH RECURSIVE blocked(pid) AS (
SELECT pid FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid))
UNION SELECT a.pid FROM pg_stat_activity a JOIN blocked b ON b.pid=ANY(pg_blocking_pids(a.pid))
) SELECT count(DISTINCT pid) FROM blocked`, pid).Scan(&waiting)
		if err != nil {
			t.Fatal("observe concurrent admission", err)
		}
		if waiting >= 2 {
			break
		}
		select {
		case early := <-results:
			if early.response != nil {
				_ = early.response.Body.Close()
			}
			t.Fatal("request settled before reaching the held owner mutation", early.err)
		case <-ctx.Done():
			t.Fatal("concurrent requests did not reach the owner mutation", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := barrier.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var first map[string]any
	for range 2 {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatal(result.err)
			}
			data := httptestx.RequireSuccessEnvelope(t, result.response, status)["data"].(map[string]any)
			if first == nil {
				first = data
				continue
			}
			if identity == "" {
				if !reflect.DeepEqual(first, data) {
					t.Fatal("concurrent exact replay changed the committed response")
				}
			} else if first[identity] == nil || first[identity] != data[identity] {
				t.Fatal("concurrent admission allocated distinct identities")
			}
		case <-ctx.Done():
			t.Fatal("concurrent admission did not settle", ctx.Err())
		}
	}
	return first
}

func assertConcurrentReportingAdmissions(t *testing.T, harness *appsupport.ServerHarness, actor authcookietest.AuthCookies, incidentID string) {
	t.Helper()
	snapshotJob := concurrentReportingMutation(t, harness, actor, "reporting_job_payloads", http.MethodPost, "/api/v1/snapshots", map[string]any{"incident_id": incidentID, "client_txn_id": "concurrent-snapshot"}, http.StatusAccepted, "job_id")
	snapshotID := requireSucceededJobResourceID(t, harness, actor, snapshotJob, "snapshot")
	compositionPath := "/api/v1/incidents/" + incidentID + "/report-compositions"
	draft := concurrentReportingMutation(t, harness, actor, "report_compositions", http.MethodPost, compositionPath, map[string]any{"client_txn_id": "concurrent-composition", "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion}, http.StatusCreated, "")
	compositionPath += "/" + draft["composition_id"].(string)
	updated := concurrentReportingMutation(t, harness, actor, "report_compositions", http.MethodPatch, compositionPath, map[string]any{"client_txn_id": "concurrent-update", "base_draft_version": 1, "authored_texts": []any{}}, http.StatusOK, "")
	if updated["draft_version"] != float64(2) {
		t.Fatal("concurrent update changed draft more than once")
	}
	concurrentReportingMutation(t, harness, actor, "report_composition_versions", http.MethodPost, compositionPath+"/versions", map[string]any{"client_txn_id": "concurrent-freeze", "base_draft_version": 2}, http.StatusCreated, "")
	raw, err := json.Marshal(requireSnapshotExportModel(t, harness.DB, snapshotID))
	if err != nil {
		t.Fatal(err)
	}
	var snapshot reporting.SnapshotModel
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	_, digest, err := reporting.ResolveRedactionProfile(reporting.InternalRedactionProfileID, reporting.DefaultRedactionProfileVersion, nil, snapshot.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	preview := concurrentReportingMutation(t, harness, actor, "report_composition_preview_attempts", http.MethodPost, compositionPath+"/preview", map[string]any{"client_txn_id": "concurrent-preview", "source_kind": "draft", "snapshot_id": snapshotID, "derivation_version": reporting.DerivationVersion, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion, "redaction_profile_sha256": digest, "render_environment_profile_id": "cartulary.reporting.local.v1", "output_kind": reporting.OutputKindSlidev, "output_options": map[string]any{}, "recipient_partition_refs": []any{}, "graph_projection_refs": []any{}}, http.StatusAccepted, "")
	requireJobStatus(t, harness, actor, preview["render_attempt_id"].(string), "succeeded")
	concurrentReportingMutation(t, harness, actor, "report_compositions", http.MethodDelete, compositionPath, map[string]any{"client_txn_id": "concurrent-retire", "base_draft_version": 2}, http.StatusOK, "")
	releaseJob := concurrentReportingMutation(t, harness, actor, "reporting_job_payloads", http.MethodPost, "/api/v1/releases", map[string]any{"client_txn_id": "concurrent-release", "snapshot_id": snapshotID, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalReview}, http.StatusAccepted, "job_id")
	releaseID := requireSucceededJobResourceID(t, harness, actor, releaseJob, "release")
	concurrentReportingMutation(t, harness, actor, "reporting_releases", http.MethodPost, "/api/v1/releases/"+releaseID+"/invalidate", map[string]any{"client_txn_id": "concurrent-invalidate", "reason": "review complete"}, http.StatusOK, "")
}
