package networkflow_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Extends the real saved-graph lifecycle scenario. The renderer remains an
// explicit protocol fixture; this establishes admission, recovery and execution
// independence, not rendering or process-confinement conformance.
func admitFrozenGraphPreview(t testing.TB, h *appsupport.ServerHarness, incident uuid.UUID, actor authcookietest.AuthCookies, snapshot string, refs []map[string]any) uuid.UUID {
	t.Helper()
	post := func(path string, body map[string]any, status int) map[string]any {
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+path, body,
			httptestx.WithCookies(actor.Session, actor.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRF.Value)), status)["data"].(map[string]any)
	}
	base := "/api/v1/incidents/" + incident.String() + "/report-compositions"
	composition := post(base, map[string]any{
		"client_txn_id": "txn-frozen-graph-composition", "template_id": reporting.DefaultTemplateID,
		"template_version": reporting.DefaultTemplateVersion, "deck_ops": []any{}, "diagram_decls": []any{}, "authored_texts": []any{},
	}, http.StatusCreated)
	var capture []byte
	var digest string
	if err := h.Pool.QueryRow(t.Context(), `SELECT p.request_json->'source_capture',s.source_capture_sha256 FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id=$1`, snapshot).Scan(&capture, &digest); err != nil {
		t.Fatal(err)
	}
	model, err := reporting.MaterializeSnapshotCapture(t.Context(), capture, digest)
	if err != nil {
		t.Fatal(err)
	}
	_, profileSHA, err := reporting.ResolveRedactionProfile(reporting.InternalRedactionProfileID, reporting.DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	preview := post(base+"/"+composition["composition_id"].(string)+"/preview", map[string]any{
		"client_txn_id": "txn-frozen-graph-preview", "source_kind": "draft", "snapshot_id": snapshot,
		"derivation_version": reporting.DerivationVersion, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
		"redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion,
		"redaction_profile_sha256": profileSHA, "render_environment_profile_id": "cartulary.reporting.local.v1",
		"output_kind": reporting.OutputKindSlidev, "output_options": map[string]any{}, "recipient_partition_refs": []any{}, "graph_projection_refs": refs,
	}, http.StatusAccepted)
	return uuid.MustParse(preview["render_attempt_id"].(string))
}

func retainedReportingInputSnapshot(t testing.TB, h *appsupport.ServerHarness, ids ...uuid.UUID) string {
	t.Helper()
	var result string
	if err := h.Pool.QueryRow(t.Context(), `SELECT coalesce(jsonb_agg(jsonb_build_array(job_id,job_kind,incident_id,actor_user_id,request_json,created_at,render_admission_sha256,admitted_graph_sources_sha256,admitted_profile_sha256,admitted_template_manifest_sha256,snapshot_admission_json,snapshot_admission_sha256) ORDER BY job_id),'[]')::text FROM reporting_job_payloads WHERE job_id=ANY($1::uuid[])`, ids).Scan(&result); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertFrozenReportingRestore(t testing.TB, runtime *appsupport.Runtime, h *appsupport.ServerHarness, actor authcookietest.AuthCookies, frozen string, ids ...uuid.UUID) {
	t.Helper()
	if retainedReportingInputSnapshot(t, h, ids...) != frozen {
		t.Fatal("Graph reconstruction or retirement changed Reporting's immutable input")
	}
	receipts := retainedReportingReceiptSnapshot(t, h, ids...)
	var sourceSessions int
	if err := h.Pool.QueryRow(t.Context(), `SELECT count(*) FROM user_sessions`).Scan(&sourceSessions); err != nil || sourceSessions == 0 {
		t.Fatalf("source fixture lacks session invalidation evidence: count=%d err=%v", sourceSessions, err)
	}
	h.Server.Close()
	// Graceful Jobs shutdown must release attempts without charging a failure.
	for _, id := range ids {
		var status string
		var failures int
		var attempt *uuid.UUID
		if err := h.Pool.QueryRow(t.Context(), `SELECT status,handler_failure_count,handler_attempt_id FROM jobs WHERE job_id=$1`, id).Scan(&status, &failures, &attempt); err != nil || (status != "queued" && status != "running") || failures != 0 || attempt != nil {
			t.Fatalf("shutdown did not preserve retry: status=%s failures=%d attempt=%v err=%v", status, failures, attempt, err)
		}
	}
	if _, err := h.Pool.Exec(t.Context(), `DELETE FROM graph_projection_results`); err != nil {
		t.Fatal(err)
	}
	// Advance the application clock beyond the former consumer-lease lifetime.
	// Session cookies are unnecessary: retained jobs use current actor authority.
	restarted := runtime.CaptureRestoreServer(t, h, appsupport.ServerOptions{
		Prefix: "frozen-reporting-source-unclaimed",
		Env:    map[string]string{"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": "false"}, TestRouteMode: httptestx.TestRouteModeDisabled,
		ConfigureRuntime: func(options *server.Options) {
			options.Now = func() time.Time { return time.Now().UTC().Add(time.Hour) }
		},
	})
	var restoredSessions int
	if err := restarted.Pool.QueryRow(t.Context(), `SELECT count(*) FROM user_sessions`).Scan(&restoredSessions); err != nil || restoredSessions != 0 {
		t.Fatalf("restore retained session authority: count=%d err=%v", restoredSessions, err)
	}
	response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet,
		restarted.Server.HTTP.URL+"/api/v1/jobs/"+ids[0].String(), nil, httptestx.WithCookies(actor.Session))
	httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "session_required")
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	for _, id := range ids {
		for {
			var status string
			var failures int
			var attempted bool
			var retryAt *time.Time
			var failureClass string
			if err := restarted.Pool.QueryRow(ctx, `SELECT status,handler_failure_count,handler_attempt_id IS NOT NULL,handler_next_attempt_at,COALESCE(handler_last_error,'') FROM jobs WHERE job_id=$1`, id).Scan(&status, &failures, &attempted, &retryAt, &failureClass); err != nil {
				t.Fatal(err)
			}
			if status == "succeeded" {
				if failures != 0 {
					t.Fatalf("source-independent retry charged %d failures", failures)
				}
				break
			}
			if status == "failed" || status == "canceled" {
				t.Fatalf("retained Reporting job reached %s with its source unclaimed", status)
			}
			select {
			case <-ctx.Done():
				t.Fatalf("retained Reporting job did not execute independently: status=%s failures=%d attempted=%t retry_scheduled=%t failure_class=%s: %v", status, failures, attempted, retryAt != nil, failureClass, ctx.Err())
			case <-time.After(25 * time.Millisecond):
			}
		}
	}
	if retainedReportingInputSnapshot(t, restarted, ids...) != frozen {
		t.Fatal("ordinary Jobs recovery or retry changed frozen Reporting input")
	}
	if retainedReportingReceiptSnapshot(t, restarted, ids...) != receipts {
		t.Fatal("current restore or execution changed admitted receipt identity or immutable preview acknowledgement")
	}
	var derived int
	if err := restarted.Pool.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM graph_projection_results)+(SELECT count(*) FROM graph_projection_result_vertices)+(SELECT count(*) FROM graph_projection_result_edges)`).Scan(&derived); err != nil || derived != 0 {
		t.Fatalf("unclaimed source unexpectedly reconstructed Graph state: count=%d err=%v", derived, err)
	}
}

// Common Job receipts may acquire the terminal resource under their owner
// contract; preview acknowledgements remain immutable. Both retain the exact
// initiating identity, normalized request hash and HTTP status across restore.
func retainedReportingReceiptSnapshot(t testing.TB, h *appsupport.ServerHarness, ids ...uuid.UUID) string {
	t.Helper()
	var count int
	var result string
	err := h.Pool.QueryRow(t.Context(), `SELECT count(*), COALESCE(jsonb_agg(jsonb_build_array(j.job_id,r.route_key,r.actor_user_id,r.scope_key,r.client_txn_id,r.request_hash,r.status_code,CASE WHEN r.route_key='report_compositions.preview' THEN r.response_json ELSE NULL END) ORDER BY j.job_id)::text,'[]') FROM jobs j JOIN route_idempotency r ON r.route_key=j.extension_idempotency_route_key AND r.actor_user_id=j.submitted_by_user_id AND r.scope_key=j.extension_idempotency_scope_key AND r.client_txn_id=j.extension_idempotency_identity->>'client_txn_id' WHERE j.job_id=ANY($1::uuid[])`, ids).Scan(&count, &result)
	if err != nil || count != len(ids) {
		t.Fatalf("retained route receipt coverage: want=%d got=%d err=%v", len(ids), count, err)
	}
	return result
}

func assertFrozenReportingRestoreFailures(t *testing.T, runtime *appsupport.Runtime, source *appsupport.ServerHarness, ids ...uuid.UUID) {
	t.Helper()
	if len(ids) != 2 {
		t.Fatal("fixture requires release and preview identities")
	}
	for _, fault := range []string{"authorization lost", "actor inactive", "release role lost", "incident closed", "revoked during release", "revoked during preview", "preparation changed during release", "inconsistent digest rejected"} {
		t.Run(fault, func(t *testing.T) {
			if _, err := source.Pool.Exec(t.Context(), `UPDATE jobs SET handler_next_attempt_at=clock_timestamp()+interval '2 hours' WHERE job_id=ANY($1::uuid[])`, ids); err != nil {
				t.Fatal(err)
			}
			resume := make(chan struct{})
			gate := &restoredReportingRuntime{started: make(chan struct{}, 1), resume: resume}
			late := fault == "revoked during release" || fault == "revoked during preview" || fault == "preparation changed during release"
			target := runtime.CaptureRestoreServer(t, source, appsupport.ServerOptions{
				Prefix: "frozen-reporting-rejection", Env: map[string]string{"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": "false"}, TestRouteMode: httptestx.TestRouteModeDisabled,
				ConfigureRuntime: func(options *server.Options) {
					options.Now = func() time.Time { return time.Now().UTC().Add(time.Hour) }
					if late {
						options.ReportingRuntime = gate
					}
				},
			})
			defer target.Server.Close()
			var attempts int
			if err := target.Pool.QueryRow(t.Context(), `SELECT count(*) FROM jobs WHERE job_id=ANY($1::uuid[]) AND (handler_attempt_id IS NOT NULL OR status NOT IN ('queued','running'))`, ids).Scan(&attempts); err != nil || attempts != 0 {
				t.Fatalf("restore did not preserve deferred attempts: count=%d err=%v", attempts, err)
			}
			if late {
				id := ids[0]
				if fault == "revoked during preview" {
					id = ids[1]
				}
				if _, err := target.Pool.Exec(t.Context(), `UPDATE jobs SET handler_next_attempt_at=NULL WHERE job_id=$1`, id); err != nil {
					t.Fatal(err)
				}
				select {
				case <-gate.started:
				case <-time.After(15 * time.Second):
					var status, code string
					err := target.Pool.QueryRow(t.Context(), `SELECT status,COALESCE(error_summary_json->>'code','') FROM jobs WHERE job_id=$1`, id).Scan(&status, &code)
					t.Fatalf("worker did not reach the declared render protocol: status=%s code=%s err=%v", status, code, err)
				}
			}
			before := retainedReportingInputSnapshot(t, target, ids...)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			tx, err := target.Pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(t.Context())
			switch fault {
			case "authorization lost", "revoked during release", "revoked during preview":
				_, err = tx.Exec(ctx, `DELETE FROM incident_memberships WHERE (incident_id,user_id) IN (SELECT incident_id,submitted_by_user_id FROM jobs WHERE job_id=ANY($1::uuid[]))`, ids)
			case "preparation changed during release":
				var changed pgconn.CommandTag
				changed, err = tx.Exec(ctx, `UPDATE reporting_job_payloads SET render_preparation_json=jsonb_set(render_preparation_json,'{source_facts_sha256}',to_jsonb(repeat('0',64))) WHERE job_id=$1 AND render_preparation_json IS NOT NULL`, ids[0])
				if err == nil && changed.RowsAffected() != 1 {
					t.Fatal("fixture did not alter prepared input")
				}
			case "actor inactive":
				_, err = tx.Exec(ctx, `UPDATE users SET is_active=false WHERE id IN (SELECT submitted_by_user_id FROM jobs WHERE job_id=ANY($1::uuid[]))`, ids)
			case "release role lost":
				_, err = tx.Exec(ctx, `UPDATE incident_memberships SET role='viewer' WHERE (incident_id,user_id) IN (SELECT incident_id,submitted_by_user_id FROM jobs WHERE job_id=ANY($1::uuid[]))`, ids)
			case "incident closed":
				_, err = tx.Exec(ctx, `UPDATE incidents SET status='closed',closed_at=clock_timestamp() WHERE id IN (SELECT incident_id FROM jobs WHERE job_id=ANY($1::uuid[]))`, ids)
			case "inconsistent digest rejected":
				_, err = tx.Exec(ctx, `UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{graph_sources_sha256}',to_jsonb(repeat('0',64))) WHERE job_id=ANY($1::uuid[])`, ids)
				var violation *pgconn.PgError
				if !errors.As(err, &violation) || violation.Code != "23514" || violation.ConstraintName != "reporting_render_recipe_v1_ck" {
					t.Fatal("inconsistent retained digest was not rejected by its storage contract")
				}
				if err := tx.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
				if retainedReportingInputSnapshot(t, target, ids...) != before {
					t.Fatal("rejected digest update changed immutable input")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `UPDATE jobs SET handler_next_attempt_at=NULL WHERE job_id=ANY($1::uuid[])`, ids); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if late {
				close(resume)
			}
			for index, id := range ids {
				want := "failed"
				if fault == "incident closed" || (fault == "release role lost" || fault == "preparation changed during release") && index == 1 {
					want = "succeeded"
				}
				wantReason := "initiating_actor_unavailable"
				if fault == "preparation changed during release" {
					wantReason = "retained_input_mismatch"
				}
				for {
					var status, code, reason string
					var retryable bool
					if err := target.Pool.QueryRow(ctx, `SELECT status,COALESCE(error_summary_json->>'code',''),COALESCE(error_summary_json->'details'->>'reason_code',''),COALESCE((error_summary_json->>'retryable')::boolean,false) FROM jobs WHERE job_id=$1`, id).Scan(&status, &code, &reason, &retryable); err != nil {
						t.Fatal(err)
					}
					if status == "succeeded" || status == "failed" || status == "canceled" {
						if status != want || want == "failed" && (code != "release_render_failed" || reason != wantReason || retryable) {
							t.Fatalf("restored authority outcome: want=%s got=%s code=%s reason=%s retryable=%t", want, status, code, reason, retryable)
						}
						break
					}
					select {
					case <-ctx.Done():
						t.Fatalf("restored authority outcome did not complete: status=%s: %v", status, ctx.Err())
					case <-time.After(25 * time.Millisecond):
					}
				}
				if want == "failed" {
					var outputs int
					if err := target.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM reporting_releases WHERE create_job_id=$1 AND release_state!='render_failed')+(SELECT count(*) FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1)+(SELECT count(*) FROM extension_job_commit_proofs WHERE job_id=$1)`, id).Scan(&outputs); err != nil || outputs != 0 {
						t.Fatalf("authority-denied render published output or proof: count=%d err=%v", outputs, err)
					}
				}
			}
			if retainedReportingInputSnapshot(t, target, ids...) != before {
				t.Fatal("authority transition changed immutable input")
			}
			if fault == "authorization lost" {
				journal := func(h *appsupport.ServerHarness) string {
					t.Helper()
					var value string
					var count int
					if err := h.Pool.QueryRow(t.Context(), `SELECT count(*),coalesce(jsonb_agg(to_jsonb(f) ORDER BY job_id),'[]')::text FROM reporting_render_failure_records f WHERE job_id=ANY($1::uuid[])`, ids).Scan(&count, &value); err != nil || count != len(ids) {
						t.Fatal("independent failure journal missing", count, err)
					}
					return value
				}
				beforeJournal := journal(target)
				target.Server.Close()
				again := runtime.CaptureRestoreServer(t, target, appsupport.ServerOptions{Prefix: "retained-failure-journal", Env: map[string]string{"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": "false"}, TestRouteMode: httptestx.TestRouteModeDisabled, ConfigureRuntime: func(options *server.Options) {
					options.Now = func() time.Time { return time.Now().UTC().Add(time.Hour) }
				}})
				defer again.Server.Close()
				if journal(again) != beforeJournal {
					t.Fatal("current capture/restore changed independent failure bytes or identities")
				}
			}
			if fault == "incident closed" {
				terminal := func() string {
					t.Helper()
					var value string
					if err := target.Pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(j) ORDER BY job_id)::text FROM jobs j WHERE job_id=ANY($1::uuid[])`, ids).Scan(&value); err != nil {
						t.Fatal(err)
					}
					return value
				}
				committed := terminal()
				if _, err := target.Pool.Exec(ctx, `DELETE FROM incident_memberships WHERE (incident_id,user_id) IN (SELECT incident_id,submitted_by_user_id FROM jobs WHERE job_id=ANY($1::uuid[]))`, ids); err != nil {
					t.Fatal(err)
				}
				if terminal() != committed {
					t.Fatal("later revocation changed committed Job outcomes")
				}
				var proofs int
				if err := target.Pool.QueryRow(ctx, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id=ANY($1::uuid[])`, ids).Scan(&proofs); err != nil || proofs != len(ids) {
					t.Fatalf("later revocation lost committed success proofs: count=%d err=%v", proofs, err)
				}
			}
		})
	}
}

// Corruption must satisfy the storage shape so this exercises the shipped
// retained-input reader, rather than merely observing a rejected fixture UPDATE.
func assertFrozenReportingInputCorruption(t *testing.T, runtime *appsupport.Runtime, source *appsupport.ServerHarness, ids ...uuid.UUID) {
	t.Helper()
	for _, fault := range []string{"graph bytes", "missing input", "admission digest", "missing snapshot input", "snapshot capture bytes"} {
		t.Run("retained input "+fault, func(t *testing.T) {
			if _, err := source.Pool.Exec(t.Context(), `UPDATE jobs SET handler_next_attempt_at=clock_timestamp()+interval '2 hours' WHERE job_id=ANY($1::uuid[])`, ids); err != nil {
				t.Fatal(err)
			}
			resume := make(chan struct{})
			close(resume)
			renderer := &restoredReportingRuntime{started: make(chan struct{}, len(ids)), resume: resume}
			target := runtime.CaptureRestoreServer(t, source, appsupport.ServerOptions{
				Prefix: "frozen-reporting-corruption", Env: map[string]string{"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": "false"}, TestRouteMode: httptestx.TestRouteModeDisabled,
				ConfigureRuntime: func(options *server.Options) {
					options.Now = func() time.Time { return time.Now().UTC().Add(time.Hour) }
					options.ReportingRuntime = renderer
				},
			})
			defer target.Server.Close()
			var snapshotJobs []uuid.UUID
			if err := target.Pool.QueryRow(t.Context(), `SELECT array_agg(DISTINCT s.create_job_id) FROM reporting_snapshots s JOIN reporting_job_payloads p ON s.snapshot_id=(p.request_json->>'snapshot_id')::uuid WHERE p.job_id=ANY($1::uuid[])`, ids).Scan(&snapshotJobs); err != nil || len(snapshotJobs) == 0 {
				t.Fatal("fixture lacks frozen snapshot inputs", err)
			}
			observed := append(append([]uuid.UUID(nil), ids...), snapshotJobs...)
			original := retainedReportingInputSnapshot(t, target, observed...)
			ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
			defer cancel()
			tx, err := target.Pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(t.Context())
			query := `UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{graph_sources,results}','[]'::jsonb) WHERE job_id=ANY($1::uuid[]) AND jsonb_array_length(request_json->'graph_sources'->'results')>0`
			wantReason := "retained_input_mismatch"
			if fault == "missing input" {
				query = `DELETE FROM reporting_job_payloads WHERE job_id=ANY($1::uuid[])`
				wantReason = "retained_input_missing"
			}
			if fault == "admission digest" {
				query = `UPDATE reporting_job_payloads SET render_admission_sha256=repeat('0',64),render_preparation_json=CASE WHEN render_preparation_json IS NULL THEN NULL ELSE jsonb_set(render_preparation_json,'{admission_sha256}',to_jsonb(repeat('0',64))) END WHERE job_id=ANY($1::uuid[])`
			}
			expectedChanges := len(ids)
			if fault == "missing snapshot input" || fault == "snapshot capture bytes" {
				expectedChanges = len(snapshotJobs)
				query = `DELETE FROM reporting_job_payloads WHERE job_id IN (SELECT s.create_job_id FROM reporting_snapshots s JOIN reporting_job_payloads p ON s.snapshot_id=(p.request_json->>'snapshot_id')::uuid WHERE p.job_id=ANY($1::uuid[]))`
				wantReason = "retained_input_missing"
				if fault == "snapshot capture bytes" {
					wantReason = "retained_input_mismatch"
					query = `UPDATE reporting_job_payloads SET request_json='{}'::jsonb WHERE job_id IN (SELECT s.create_job_id FROM reporting_snapshots s JOIN reporting_job_payloads p ON s.snapshot_id=(p.request_json->>'snapshot_id')::uuid WHERE p.job_id=ANY($1::uuid[]))`
				}
			}
			tag, err := tx.Exec(ctx, query, ids)
			if err != nil || tag.RowsAffected() != int64(expectedChanges) {
				t.Fatalf("corruption fixture did not reach stored logical bytes: changed=%d err=%v", tag.RowsAffected(), err)
			}
			if _, err := tx.Exec(ctx, `UPDATE jobs SET handler_next_attempt_at=NULL WHERE job_id=ANY($1::uuid[])`, ids); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			corrupted := retainedReportingInputSnapshot(t, target, observed...)
			if corrupted == original {
				t.Fatal("corruption fixture left input unchanged")
			}
			for _, id := range ids {
				for {
					var status, code, reason, failure string
					var retries int
					var retryable, retainedDiagnostic bool
					if err := target.Pool.QueryRow(ctx, `SELECT j.status,COALESCE(j.error_summary_json->>'code',''),COALESCE(j.error_summary_json->'details'->>'reason_code',''),COALESCE((SELECT convert_from(f.record_json,'UTF8')::jsonb->'summary'->>'failure_code' FROM reporting_render_failure_records f WHERE f.job_id=j.job_id),''),j.handler_failure_count,COALESCE((j.error_summary_json->>'retryable')::boolean,false),EXISTS(SELECT 1 FROM reporting_render_failure_records f WHERE f.job_id=j.job_id) FROM jobs j WHERE j.job_id=$1`, id).Scan(&status, &code, &reason, &failure, &retries, &retryable, &retainedDiagnostic); err != nil {
						t.Fatal(err)
					}
					if status == "failed" || status == "succeeded" || status == "canceled" {
						if status != "failed" || code != "release_render_failed" || reason != wantReason || failure != "retained_input_invalid" || retries != 0 || retryable || !retainedDiagnostic {
							t.Errorf("corrupt retained input lacks its terminal owner outcome: status=%s code=%s reason=%s failure=%s retries=%d retryable=%t diagnostic=%t", status, code, reason, failure, retries, retryable, retainedDiagnostic)
						}
						break
					}
					select {
					case <-ctx.Done():
						t.Fatal("corrupt input did not reach terminal rejection", ctx.Err())
					case <-time.After(25 * time.Millisecond):
					}
				}
			}
			select {
			case <-renderer.started:
				t.Error("corrupt retained input reached rendering")
			default:
			}
			var outputs int
			if err := target.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM reporting_releases WHERE create_job_id=ANY($1::uuid[]))+(SELECT count(*) FROM reporting_composition_preview_outputs WHERE render_attempt_id=ANY($1::uuid[]))+(SELECT count(*) FROM extension_job_commit_proofs WHERE job_id=ANY($1::uuid[]))`, ids).Scan(&outputs); err != nil || outputs != 0 {
				t.Fatalf("corrupt retained input published owner output or proof: count=%d err=%v", outputs, err)
			}
			if retainedReportingInputSnapshot(t, target, observed...) != corrupted {
				t.Fatal("corruption handling rewrote retained input")
			}
		})

	}
}
