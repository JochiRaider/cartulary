package reporting_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase11_I_11_REPORTING_01_SnapshotReplayAndReleaseProvenanceAreStable(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-provenance")

	adminLogin, adminUserID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-incident",
		"incident_key":  "IR-REPORTING-01",
		"title":         "Reporting Incident",
		"description":   strings.Repeat("external recipient narrative ", 12),
		"severity":      "high",
		"current_phase": "analysis",
	})
	incidentID := incident["incident_id"].(string)
	fixture := seedReportingWorkbookFixture(t, harness.DB, incidentID, adminUserID)

	createSnapshot := func() *http.Response {
		t.Helper()
		return phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/snapshots",
			map[string]any{
				"incident_id":   incidentID,
				"client_txn_id": "txn-reporting-snapshot",
			},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
	}
	firstSnapshot := createSnapshot()
	firstSnapshotJob := httptestx.RequireSuccessEnvelope(t, firstSnapshot, http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, firstSnapshotJob, "snapshot")
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	if snapshot["export_model_sha256"] == "" || snapshot["source_change_set_high_watermark"] == "" {
		t.Fatalf("snapshot must expose immutable export model and source boundary provenance: %#v", snapshot)
	}
	exportModel := requireSnapshotExportModel(t, harness.DB, snapshotID)
	requireExportModelCoverage(t, exportModel, fixture)
	initialWatermark := snapshot["source_change_set_high_watermark"]

	phase2test.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replaySnapshot := createSnapshot()
	replaySnapshotJob := httptestx.RequireSuccessEnvelope(t, replaySnapshot, http.StatusAccepted)["data"].(map[string]any)
	if replaySnapshotJob["job_id"] != firstSnapshotJob["job_id"] {
		t.Fatalf("snapshot replay must return original job: first=%#v replay=%#v", firstSnapshotJob, replaySnapshotJob)
	}
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
		t.Fatalf("snapshot idempotent replay must not create fresh state after live incident changed, got %d rows", got)
	}
	replayedSnapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	if replayedSnapshot["source_change_set_high_watermark"] != initialWatermark {
		t.Fatalf("snapshot replay must preserve original resolved source boundary: before=%v after=%v", initialWatermark, replayedSnapshot["source_change_set_high_watermark"])
	}

	createReleaseResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               "html",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createReleaseJob := httptestx.RequireSuccessEnvelope(t, createReleaseResp, http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, createReleaseJob, "release")
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_state"] != reporting.ReleaseStatePendingApproval {
		t.Fatalf("external release must start pending approval, got %#v", release["release_state"])
	}
	if release["redaction_profile_id"] != reporting.ExternalRedactionProfileID || release["redaction_profile_sha256"] == "" {
		t.Fatalf("release must bind immutable redaction profile identity, got %#v", release)
	}
	if release["output_sha256"] == "" || release["redaction_manifest_sha256"] == "" {
		t.Fatalf("release must expose output and manifest hashes, got %#v", release)
	}
	rendered, manifest := requireReleaseArtifacts(t, harness.DB, releaseID)
	if refs := release["recipient_partition_refs"].([]any); len(refs) != 0 {
		t.Fatalf("omitted recipient partitions must canonicalize to [], got %#v", refs)
	}
	if strings.Contains(rendered, "internal workbook context") {
		t.Fatalf("external output must not package working material: %s", rendered)
	}
	if manifest["profile_sha256"] != release["redaction_profile_sha256"] {
		t.Fatalf("manifest profile digest must match release provenance: manifest=%#v release=%#v", manifest["profile_sha256"], release["redaction_profile_sha256"])
	}
	entries := manifestEntriesByPath(t, manifest)
	internalEntry := entries["/incident/internal_note"]
	if internalEntry["action"] != reporting.ActionDrop || internalEntry["outcome"] != "dropped" {
		t.Fatalf("working material must be dropped with stable manifest outcome, got %#v", internalEntry)
	}
	descriptionEntry := entries["/incident/description"]
	if descriptionEntry["action"] != reporting.ActionTruncate || descriptionEntry["outcome"] != "truncated" {
		t.Fatalf("source evidence must be truncated for external release, got %#v", descriptionEntry)
	}
	partyEntry := entries["/parties/"+fixture["party"]]
	if partyEntry["outcome"] != "dropped_disclosure_partition" {
		t.Fatalf("party partition must be dropped when not requested, got %#v", partyEntry)
	}

	createPartitionedRelease := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-party",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               "html",
			"recipient_partition_refs":  []string{"party:" + fixture["party"], "party:" + fixture["party"]},
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	partitionedJob := httptestx.RequireSuccessEnvelope(t, createPartitionedRelease, http.StatusAccepted)["data"].(map[string]any)
	partitionedReleaseID := requireSucceededJobResourceID(t, harness, adminLogin, partitionedJob, "release")
	partitionedRelease := requireRelease(t, harness, adminLogin, partitionedReleaseID)
	partitionRefs := partitionedRelease["recipient_partition_refs"].([]any)
	if len(partitionRefs) != 1 || partitionRefs[0] != "party:"+fixture["party"] {
		t.Fatalf("recipient partitions must sort and coalesce, got %#v", partitionRefs)
	}
	if partitionedRelease["output_sha256"] == release["output_sha256"] || partitionedRelease["redaction_manifest_sha256"] == release["redaction_manifest_sha256"] {
		t.Fatalf("recipient partition must change output and manifest hashes: base=%#v partitioned=%#v", release, partitionedRelease)
	}
	_, partitionedManifest := requireReleaseArtifacts(t, harness.DB, partitionedReleaseID)
	partitionedEntries := manifestEntriesByPath(t, partitionedManifest)
	if partitionedEntries["/parties/"+fixture["party"]]["outcome"] == "dropped_disclosure_partition" {
		t.Fatalf("requested party partition must be retained in partitioned external release, got %#v", partitionedEntries["/parties/"+fixture["party"]])
	}
	manifestRefs := partitionedManifest["recipient_partition_refs"].([]any)
	if len(manifestRefs) != 1 || manifestRefs[0] != "party:"+fixture["party"] {
		t.Fatalf("redaction manifest must bind recipient partitions, got %#v", manifestRefs)
	}
}

func TestPhase11_I_11_REPORTING_02_ExternalReleaseApprovalPublishAndStateConflicts(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-lifecycle")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	reviewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "report-reviewer@example.test", "Report Reviewer", "ReviewerPass1!", false, false, true)
	editorID := phase2test.SeedLocalUserFlags(t, harness.DB, "report-editor@example.test", "Report Editor", "EditorPass1!", false, false, true)
	reviewerSession, reviewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "report-reviewer@example.test", "ReviewerPass1!")
	reviewerLogin := phase2test.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	editorSession, editorCSRF := phase2test.LoginLocalUser(t, harness.Server, "report-editor@example.test", "EditorPass1!")
	editorLogin := phase2test.LoginResult{SessionCookie: editorSession, CSRFCookie: editorCSRF}

	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-lifecycle-incident",
		"incident_key":  "IR-REPORTING-02",
		"title":         "Reporting Lifecycle Incident",
		"description":   "approval lifecycle source text",
	})
	incidentID := incident["incident_id"].(string)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-reviewer-membership",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-editor-membership",
		"user_id":       editorID,
		"role":          "editor",
	})

	snapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-lifecycle-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")

	releaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-lifecycle-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               "html",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")

	editorApprove := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-editor-approve",
			"reason":        "editor cannot approve release",
		},
		phase2test.WithCookies(editorLogin.SessionCookie, editorLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, editorLogin.CSRFCookie.Value),
	)
	editorError := httptestx.RequireErrorEnvelope(t, editorApprove, http.StatusConflict, "release_approval_rejected")
	if editorError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "actor_lacks_approval_role" {
		t.Fatalf("editor approval must use release approval rejection reason, got %#v", editorError)
	}

	reviewerApprove := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-reviewer-approve",
			"reason":        "reviewed for release",
		},
		phase2test.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	reviewerBody := httptestx.RequireSuccessEnvelope(t, reviewerApprove, http.StatusOK)["data"].(map[string]any)
	if reviewerBody["release_state"] != reporting.ReleaseStatePendingApproval {
		t.Fatalf("reviewer approval alone must leave external release pending, got %#v", reviewerBody["release_state"])
	}

	adminApprove := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-admin-approve",
			"reason":        "approved for publication",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	adminBody := httptestx.RequireSuccessEnvelope(t, adminApprove, http.StatusOK)["data"].(map[string]any)
	if adminBody["release_state"] != reporting.ReleaseStateApproved || adminBody["approved_at"] == nil {
		t.Fatalf("distinct reviewer and admin approvals must approve external release, got %#v", adminBody)
	}

	publish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-publish",
			"reason":        "publish",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	publishBody := httptestx.RequireSuccessEnvelope(t, publish, http.StatusOK)["data"].(map[string]any)
	if publishBody["release_state"] != reporting.ReleaseStatePublished || publishBody["published_at"] == nil {
		t.Fatalf("approved release must publish synchronously, got %#v", publishBody)
	}

	republish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-publish-again",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	errorBody := httptestx.RequireErrorEnvelope(t, republish, http.StatusConflict, "release_state_conflict")
	if errorBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "already_published" {
		t.Fatalf("republish must fail with stable already_published reason, got %#v", errorBody)
	}
}

func TestPhase11_I_11_REPORTING_03_BoundaryReplayDefaultsAndActionIdempotency(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-idempotency")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-idempotency-incident",
		"incident_key":  "IR-REPORTING-03",
		"title":         "Reporting Idempotency Incident",
		"description":   "recipient-visible source text",
	})
	incidentID := incident["incident_id"].(string)

	snapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-idempotency-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	watermark := snapshot["source_change_set_high_watermark"].(string)

	phase2test.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replayWithExplicitBoundary := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":                      incidentID,
			"client_txn_id":                    "txn-reporting-idempotency-snapshot",
			"source_change_set_high_watermark": watermark,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayJob := httptestx.RequireSuccessEnvelope(t, replayWithExplicitBoundary, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != snapshotJob["job_id"] {
		t.Fatalf("explicit snapshot replay must return original job: first=%#v replay=%#v", snapshotJob, replayJob)
	}
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
		t.Fatalf("explicit original boundary replay must not create a new snapshot, got %d rows", got)
	}

	failedReleaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-render-failed-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             reporting.ReleaseScopeExternal,
			"output_kind":               reporting.OutputKindReenactment,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	failedReleaseID := requireFailedReleaseJob(t, harness, adminLogin, failedReleaseJob, "template_render_failed")
	failedRelease := requireRelease(t, harness, adminLogin, failedReleaseID)
	if failedRelease["release_state"] != reporting.ReleaseStateRenderFailed ||
		failedRelease["render_failed_reason_code"] != "template_render_failed" ||
		failedRelease["output_sha256"] != nil ||
		failedRelease["redaction_manifest_sha256"] != nil ||
		failedRelease["output_media_type"] != nil {
		t.Fatalf("render failure must persist nullable render-failed release resource, got %#v", failedRelease)
	}
	for _, action := range []string{"approve", "publish", "invalidate"} {
		resp := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+failedReleaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-render-failed-" + action},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "release_state_conflict")
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "render_failed" {
			t.Fatalf("%s on render_failed release must return render_failed conflict, got %#v", action, body)
		}
	}

	releaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-idempotency-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               "markdown",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_scope"] != reporting.ReleaseScopeInternalDraft || release["release_state"] != reporting.ReleaseStateApproved {
		t.Fatalf("omitted release_scope must default to approved internal draft, got %#v", release)
	}

	publish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-idempotency-publish",
			"reason":        "",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	firstPublish := httptestx.RequireSuccessEnvelope(t, publish, http.StatusOK)["data"].(map[string]any)
	replayPublish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-idempotency-publish",
			"reason":        nil,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	secondPublish := httptestx.RequireSuccessEnvelope(t, replayPublish, http.StatusOK)["data"].(map[string]any)
	if firstPublish["published_at"] != secondPublish["published_at"] || secondPublish["release_state"] != reporting.ReleaseStatePublished {
		t.Fatalf("publish replay with null/empty reason must return original published resource: first=%#v second=%#v", firstPublish, secondPublish)
	}
}

func seedReportingWorkbookFixture(t testing.TB, db *sql.DB, incidentID string, actorID string) map[string]string {
	t.Helper()
	ids := map[string]string{}
	ids["timeline"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'timeline_event', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO timeline_events (record_id, incident_id, summary, details, source_text, capture_state, created_by_user_id, updated_by_user_id)
    SELECT record_id, $1, 'Initial timeline summary', 'Timeline details', 'Timeline source text', 'rough', $2, $2
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	ids["host"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'host', $2, $2)
    RETURNING record_id
), host AS (
    INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
    SELECT record_id, $1, 'host-reporting-01', 'host-reporting-01.example.test', 'canonical', $2, $2
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO host_grid_projection (record_id, incident_id, row_version, display_name, hostname, host_state, edited_at)
    SELECT record_id, $1, 1, 'host-reporting-01', 'host-reporting-01.example.test', 'canonical', now()
      FROM host
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["identity"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'identity', $2, $2)
    RETURNING record_id
), identity AS (
    INSERT INTO identities (record_id, incident_id, display_name, upn, email, identity_state, created_by_user_id, updated_by_user_id)
    SELECT record_id, $1, 'Report Identity', 'report.identity@example.test', 'report.identity@example.test', 'canonical', $2, $2
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO identity_grid_projection (record_id, incident_id, row_version, display_name, upn, email, identity_state, edited_at)
    SELECT record_id, $1, 1, 'Report Identity', 'report.identity@example.test', 'report.identity@example.test', 'canonical', now()
      FROM identity
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["party"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'party', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO parties (record_id, incident_id, display_name, party_kind, organization_name, primary_email)
    SELECT record_id, $1, 'Recipient Organization', 'organization', 'Recipient Org', 'recipient@example.test'
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	ids["evidence"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'evidence', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, requested_at, blob_hash, storage_ref)
    SELECT record_id, $1, 'Evidence metadata only', 'received', 'complete', now(), 'raw-blob-hash-must-not-export', 'object://evidence/reporting'
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	ids["task"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'task_request', $2, $2)
    RETURNING record_id
), task AS (
    INSERT INTO task_requests (record_id, incident_id, title, status, owner_user_id, priority, task_kind, workstream)
    SELECT record_id, $1, 'Contain reporting host', 'open', $2, 'high', 'remediation', 'containment'
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO task_request_grid_projection (
        record_id, incident_id, row_version, title, status, owner_user_id,
        priority, task_kind, workstream, linked_record_count, updated_at, no_owner
    )
    SELECT record_id, $1, 1, 'Contain reporting host', 'open', $2,
           'high', 'remediation', 'containment', 0, now(), false
      FROM task
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["decision"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'decision', $2, $2)
    RETURNING record_id
), decision AS (
    INSERT INTO decisions (record_id, incident_id, summary, status, owner_user_id, decision_type, rationale)
    SELECT record_id, $1, 'Escalate reporting case', 'approved', $2, 'scope', 'Evidence supports escalation.'
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO decision_grid_projection (
        record_id, incident_id, row_version, summary, status, owner_user_id,
        decision_type, rationale, affected_record_count, updated_at, is_superseded
    )
    SELECT record_id, $1, 1, 'Escalate reporting case', 'approved', $2,
           'scope', 'Evidence supports escalation.', 0, now(), false
      FROM decision
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["artifact"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'artifact', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
    SELECT record_id, $1, 'note', 'Reporting analyst note', 'Working note body', $2
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	if _, err := db.Exec(`
INSERT INTO record_links (incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id)
VALUES ($1, $2, $3, 'supported_by', 'manual', $4, $4)
`, incidentID, ids["task"], ids["evidence"], actorID); err != nil {
		t.Fatalf("seed reporting support link: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO record_tags (incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, 'reporting-tag', 'reporting-tag', $3)
`, incidentID, ids["host"], actorID); err != nil {
		t.Fatalf("seed reporting tag: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, ordinal, created_by_user_id
)
VALUES ($1, 'host', 'timeline.summary', 'manual', 'summary:1', 'host-reporting-01', 'host-reporting-01', 'unresolved', 1, $2)
`, ids["timeline"], actorID); err != nil {
		t.Fatalf("seed reporting entity mention: %v", err)
	}
	return ids
}

func querySeedID(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var id string
	if err := db.QueryRow(query, args...).Scan(&id); err != nil {
		t.Fatalf("seed query failed: %v", err)
	}
	return id
}

func requireSnapshotExportModel(t testing.TB, db *sql.DB, snapshotID string) map[string]any {
	t.Helper()
	var raw []byte
	if err := db.QueryRow(`
SELECT export_model_json
  FROM reporting_snapshots
 WHERE snapshot_id::text = $1
`, snapshotID).Scan(&raw); err != nil {
		t.Fatalf("query snapshot export model: %v", err)
	}
	var model map[string]any
	if err := json.Unmarshal(raw, &model); err != nil {
		t.Fatalf("decode snapshot export model: %v", err)
	}
	if strings.Contains(string(raw), "raw-blob-hash-must-not-export") || strings.Contains(string(raw), "blob_hash") {
		t.Fatalf("snapshot export model must exclude raw blob bytes/hash fields: %s", string(raw))
	}
	return model
}

func requireExportModelCoverage(t testing.TB, model map[string]any, ids map[string]string) {
	t.Helper()
	if model["schema_id"] != "cartulary.export_model.v2" || model["derivation_version"] != reporting.DerivationVersion {
		t.Fatalf("snapshot export model must use v2 identity, got %#v", model)
	}
	fields := model["fields"].([]any)
	byPath := make(map[string]map[string]any, len(fields))
	for _, item := range fields {
		field := item.(map[string]any)
		byPath[field["path"].(string)] = field
	}
	wantPaths := []string{
		"/timeline/" + ids["timeline"],
		"/hosts/" + ids["host"],
		"/identities/" + ids["identity"],
		"/parties/" + ids["party"],
		"/evidence/" + ids["evidence"],
		"/task_requests/" + ids["task"],
		"/decisions/" + ids["decision"],
		"/artifacts/" + ids["artifact"],
	}
	for _, path := range wantPaths {
		if byPath[path] == nil {
			t.Fatalf("snapshot export model missing workbook path %s; paths=%#v", path, byPath)
		}
	}
	if byPath["/relationships/"+ids["task"]] == nil {
		for path := range byPath {
			if strings.HasPrefix(path, "/relationships/") {
				goto relationshipPresent
			}
		}
		t.Fatalf("snapshot export model missing active relationship metadata; paths=%#v", byPath)
	}
relationshipPresent:
	tagPresent := false
	for path := range byPath {
		if strings.HasPrefix(path, "/tags/") {
			tagPresent = true
			break
		}
	}
	if !tagPresent {
		t.Fatalf("snapshot export model missing active tag metadata; paths=%#v", byPath)
	}
	mentionPresent := false
	for path := range byPath {
		if strings.HasPrefix(path, "/entity_mentions/") {
			mentionPresent = true
			break
		}
	}
	if !mentionPresent {
		t.Fatalf("snapshot export model missing entity mention metadata; paths=%#v", byPath)
	}
	supportRefs := byPath["/task_requests/"+ids["task"]]["support_refs"].([]any)
	if len(supportRefs) != 1 || supportRefs[0] != "/record_envelopes/"+ids["evidence"] {
		t.Fatalf("task export field must carry deterministic support refs, got %#v", supportRefs)
	}
	partyPartitions := byPath["/parties/"+ids["party"]]["disclosure_partition_refs"].([]any)
	if len(partyPartitions) != 1 || partyPartitions[0] != "party:"+ids["party"] {
		t.Fatalf("party export field must carry disclosure partition refs, got %#v", partyPartitions)
	}
}

func requireSucceededJobResourceID(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, job map[string]any, wantKind string) string {
	t.Helper()
	finalJob := requireJobStatus(t, harness, actor, job["job_id"].(string), "succeeded")
	summary := finalJob["result_summary"].(map[string]any)
	refs := summary["resource_refs"].([]any)
	if len(refs) != 1 {
		t.Fatalf("job must emit exactly one resource ref, got %#v", refs)
	}
	ref := refs[0].(map[string]any)
	if ref["kind"] != wantKind || ref["id"] == "" {
		t.Fatalf("unexpected job resource ref: got %#v want kind %q", ref, wantKind)
	}
	return ref["id"].(string)
}

func requireFailedReleaseJob(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, job map[string]any, wantReason string) string {
	t.Helper()
	finalJob := requireJobStatus(t, harness, actor, job["job_id"].(string), "failed")
	summary := finalJob["error_summary"].(map[string]any)
	if summary["code"] != "release_render_failed" {
		t.Fatalf("release render failure job code = %#v want release_render_failed: %#v", summary["code"], summary)
	}
	details := summary["details"].(map[string]any)
	if details["reason_code"] != wantReason || details["release_id"] == "" {
		t.Fatalf("release render failure details = %#v want reason %q and release id", details, wantReason)
	}
	return details["release_id"].(string)
}

func requireJobStatus(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, jobID string, wantStatus string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID,
			nil,
			phase2test.WithCookies(actor.SessionCookie),
		)
		last = httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		switch last["status"] {
		case "queued", "running", "cancel_requested":
			time.Sleep(25 * time.Millisecond)
			continue
		default:
			if last["status"] != wantStatus {
				t.Fatalf("job %s terminal status = %#v want %q: %#v", jobID, last["status"], wantStatus, last)
			}
			return last
		}
	}
	t.Fatalf("timed out waiting for job %s status %q after %#v", jobID, wantStatus, last)
	return nil
}

func requireSnapshot(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, snapshotID string) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		phase2test.WithCookies(actor.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireRelease(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, releaseID string) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		phase2test.WithCookies(actor.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireReleaseArtifacts(t testing.TB, db *sql.DB, releaseID string) (string, map[string]any) {
	t.Helper()
	var rendered string
	var manifestBytes []byte
	if err := db.QueryRow(`
SELECT rendered_output, redaction_manifest_json
  FROM reporting_releases
 WHERE release_id::text = $1
`, releaseID).Scan(&rendered, &manifestBytes); err != nil {
		t.Fatalf("query release artifacts: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode redaction manifest: %v", err)
	}
	return rendered, manifest
}

func manifestEntriesByPath(t testing.TB, manifest map[string]any) map[string]map[string]any {
	t.Helper()
	rows := manifest["entries"].([]any)
	entries := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		entry := row.(map[string]any)
		entries[entry["path"].(string)] = entry
	}
	return entries
}
